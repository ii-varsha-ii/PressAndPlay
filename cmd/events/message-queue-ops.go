package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/adarshsrinivasan/PressAndPlay/libraries/common"

	"github.com/Shopify/sarama"
	"github.com/sirupsen/logrus"
)

const (
	KAFKA_HOST_ENV               = "KAFKA_HOST"
	KAFKA_PORT_ENV               = "KAFKA_PORT"
	KAFKA_USER_DELETE_TOPIC_ENV  = "KAFKA_USER_DELETE_TOPIC"
	KAFKA_COURT_DELETE_TOPIC_ENV = "KAFKA_COURT_DELETE_TOPIC"
	KAFKA_SLOT_BOOKED_TOPIC_ENV  = "KAFKA_SLOT_BOOKED_TOPIC_ENV"
	KAFKA_RETRY_ENV              = "KAFKA_RETRY"
)

var (
	userDeletedTopic string
	slotBookedTopic  string
)

type SlotBookedNotification struct {
	UserId    string    `json:"user_id"`
	ManagerId string    `json:"manager_id"`
	CourtId   string    `json:"court_id"`
	SlotId    string    `json:"slot_id"`
	Timestamp time.Time `json:"timestamp"`
}

func (s *SlotBookedNotification) ObjToString() string {
	bytesAddress, _ := json.Marshal(s)
	return string(bytesAddress)
}

func (s *SlotBookedNotification) StringToObj(stringAddress string) {
	json.Unmarshal([]byte(stringAddress), s)
}
func newKafkaHandler() (sarama.SyncProducer, sarama.Consumer, error) {
	host := common.GetEnv(KAFKA_HOST_ENV, "localhost")
	port := common.GetEnv(KAFKA_PORT_ENV, "9092")
	userDeletedTopic = common.GetEnv(KAFKA_USER_DELETE_TOPIC_ENV, "user-deleted")
	slotBookedTopic = common.GetEnv(KAFKA_SLOT_BOOKED_TOPIC_ENV, "slot-booked")
	retry, _ := strconv.Atoi(common.GetEnv(KAFKA_RETRY_ENV, "5"))

	config := sarama.NewConfig()
	config.Admin.Retry.Max = 15
	config.Admin.Retry.Backoff = 200 * time.Millisecond
	config.Admin.Timeout = 3 * time.Second
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = retry
	config.Producer.Return.Successes = true
	config.Consumer.Return.Errors = true
	producerObj, err := sarama.NewSyncProducer([]string{fmt.Sprintf("%s:%s", host, port)}, config)
	if err != nil {
		return nil, nil, err
	}
	//if err := common.RetryOnError(15, 5*time.Second, func() error {
	//	if _, err := sarama.NewConsumer([]string{fmt.Sprintf("%s:%s", host, port)}, config); err != nil {
	//		return err
	//	}
	//	return nil
	//}); err != nil {
	//	return nil, nil, err
	//}
	logrus.Infof("Initializing kafka consumer ar %v", []string{fmt.Sprintf("%s:%s", host, port)})
	consumerObj, err := sarama.NewConsumer([]string{fmt.Sprintf("%s:%s", host, port)}, config)
	if err != nil {
		return nil, nil, err
	}

	return producerObj, consumerObj, nil
}
func initializeConsumers(topic string, notificationHandler func(*sarama.ConsumerMessage)) error {
	partitions, err := messageQueueConsumer.Partitions(topic)
	if err != nil {
		err = fmt.Errorf("exception while fetching partitions for topic %s . %v", topic, err)
		logrus.Errorf(err.Error())
		return err
	}
	// this only consumes partition no 1, you would probably want to consume all partitions
	consumer, err := messageQueueConsumer.ConsumePartition(topic, partitions[0], sarama.OffsetOldest)
	if err != nil {
		err = fmt.Errorf("exception while initialized consumser for %s topic. %v", topic, err)
		logrus.Errorf(err.Error())
		return err
	}
	logrus.Infof("Start consuming topic %v", topic)
	go func(topic string, consumer sarama.PartitionConsumer) {
		for {
			select {
			case consumerError := <-consumer.Errors():
				logrus.Errorf("received error from topic %s. %v ", topic, consumerError.Err)

			case msg := <-consumer.Messages():
				notificationHandler(msg)
				logrus.Infof("received message from topic %s. %v ", topic, string(msg.Value))
			}
		}
	}(topic, consumer)
	return err
}

func notifyUserDeletedEvent(userID string) error {
	if err := validateMessageQueueProducer(messageQueueProducer); err != nil {
		return err
	}
	msg := &sarama.ProducerMessage{
		Topic: userDeletedTopic,
		Value: sarama.StringEncoder(userID),
	}
	partition, offset, err := messageQueueProducer.SendMessage(msg)
	if err != nil {
		return err
	}
	logrus.Infof("Sent message: %s, on topic %s. partition: %v, offset: %v", userID, userDeletedTopic, partition, offset)
	return nil
}

func handleSlotBookedNotifications(message *sarama.ConsumerMessage) {
	eventMessage := string(message.Value)
	logrus.Infof("handling slot booking notification: %s", eventMessage)
	slotModel := SlotBookedNotification{}
	slotModel.StringToObj(eventMessage)
	logrus.Infof("handling slot booking notification: %s", slotModel)
	eventsModel := EventsDBData{
		UserID:           slotModel.UserId,
		ManagerID:        slotModel.ManagerId,
		SlotID:           slotModel.SlotId,
		CourtID:          slotModel.CourtId,
		BookingTimestamp: slotModel.Timestamp,
	}
	if _, err := eventsModel.createEvent(); err != nil {
		logrus.Errorf("exception while creating events %s. %v", eventsModel, err)
	}
}

func handleUserDeletedNotifications(message *sarama.ConsumerMessage) {
	userIdFromMsg := string(message.Value)
	eventsObj := EventsDBData{
		UserID: userIdFromMsg,
	}
	if _, err := eventsObj.deleteByUserID(); err != nil {
		logrus.Errorf("exception while deleting events by user id %s. %v", userIdFromMsg, err)
	}
}

func handleCourtDeletedNotifications(message *sarama.ConsumerMessage) {
	courtIdFromMsg := string(message.Value)
	eventsObj := EventsDBData{
		CourtID: courtIdFromMsg,
	}
	if _, err := eventsObj.deleteByCourtID(); err != nil {
		logrus.Errorf("exception while deleting events by court id %s. %v", courtIdFromMsg, err)
	}
}

func validateMessageQueueProducer(producer sarama.SyncProducer) error {
	if producer == nil {
		return fmt.Errorf("kafka producer connection not initialized")
	}
	return nil
}

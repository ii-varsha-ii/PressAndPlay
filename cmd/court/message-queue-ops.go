package main

import (
	"encoding/json"
	"fmt"

	"github.com/adarshsrinivasan/PressAndPlay/libraries/common"

	"strconv"
	"time"

	"github.com/Shopify/sarama"
	"github.com/sirupsen/logrus"
)

const (
	KAFKA_HOST_ENV               = "KAFKA_HOST"
	KAFKA_PORT_ENV               = "KAFKA_PORT"
	KAFKA_COURT_DELETE_TOPIC_ENV = "KAFKA_COURT_DELETE_TOPIC"
	KAFKA_SLOT_BOOKED_TOPIC_ENV  = "KAFKA_SLOT_BOOKED_TOPIC"
	KAFKA_USER_DELETE_TOPIC_ENV  = "KAFKA_USER_DELETE_TOPIC"
	KAFKA_RETRY_ENV              = "KAFKA_RETRY"
)

var (
	courtDeletedTopic string
	slotBookedTopic   string
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
	courtDeletedTopic = common.GetEnv(KAFKA_COURT_DELETE_TOPIC_ENV, "court-deleted")
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

func handleUserDeletedNotification(message *sarama.ConsumerMessage) {
	userIdFromMsg := string(message.Value)
	logrus.Infof("handling user deleted notication for userID: %s", userIdFromMsg)
	courtModel := CourtModel{
		ManagerId: userIdFromMsg,
	}
	if _, err := courtModel.deleteByManagerID(); err != nil {
		logrus.Errorf("exception while deleting courts uer details for usierID %s. %v", userIdFromMsg, err)
	}
}

func notifySlotBookedEvent(userID, managerID, courtID, slotID string) error {
	if err := validateMessageQueueProducer(messageQueueProducer); err != nil {
		return err
	}
	slotBookedObj := SlotBookedNotification{
		UserId:    userID,
		ManagerId: managerID,
		CourtId:   courtID,
		SlotId:    slotID,
		Timestamp: time.Now(),
	}
	msg := &sarama.ProducerMessage{
		Topic: slotBookedTopic,
		Value: sarama.StringEncoder(slotBookedObj.ObjToString()),
	}
	partition, offset, err := messageQueueProducer.SendMessage(msg)
	if err != nil {
		return err
	}
	logrus.Infof("Sent message: %s, on topic %s. partition: %v, offset: %v", courtID, slotBookedTopic, partition, offset)
	return nil
}

func notifyCourtDeletedEvent(courtID string) error {
	if err := validateMessageQueueProducer(messageQueueProducer); err != nil {
		return err
	}
	msg := &sarama.ProducerMessage{
		Topic: courtDeletedTopic,
		Value: sarama.StringEncoder(courtID),
	}
	partition, offset, err := messageQueueProducer.SendMessage(msg)
	if err != nil {
		return err
	}
	logrus.Infof("Sent message: %s, on topic %s. partition: %v, offset: %v", courtID, courtDeletedTopic, partition, offset)
	return nil
}

func validateMessageQueueProducer(producer sarama.SyncProducer) error {
	if producer == nil {
		return fmt.Errorf("kafka producer connection not initialized")
	}
	return nil
}

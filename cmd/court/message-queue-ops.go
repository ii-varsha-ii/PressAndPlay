package main

import (
	"encoding/json"
	"fmt"
	"github.com/adarshsrinivasan/PressAndPlay/libraries/common"
	"strconv"

	"github.com/Shopify/sarama"
	"github.com/sirupsen/logrus"
)

const (
	KAFKA_HOST_ENV               = "KAFKA_HOST"
	KAFKA_PORT_ENV               = "KAFKA_PORT"
	KAFKA_COURT_DELETE_TOPIC_ENV = "KAFKA_COURT_DELETE_TOPIC"
	KAFKA_SLOT_BOOKED_TOPIC_ENV  = "KAFKA_SLOT_BOOKED_TOPIC"
	KAFKA_RETRY_ENV              = "KAFKA_RETRY"
)

var (
	courtDeletedTopic string
	slotBookedTopic   string
)

type SlotBookedNotification struct {
	UserId  string `json:"user_id"`
	CourtId string `json:"court_id"`
	SlotId  string `json:"slot_id"`
}

func (s *SlotBookedNotification) ObjToString() string {
	bytesAddress, _ := json.Marshal(s)
	return string(bytesAddress)
}

func (s *SlotBookedNotification) StringToObj(stringAddress string) {
	json.Unmarshal([]byte(stringAddress), s)
}

func newKafkaHandler() (sarama.SyncProducer, error) {
	host := common.GetEnv(KAFKA_HOST_ENV, "localhost")
	port := common.GetEnv(KAFKA_PORT_ENV, "9092")
	courtDeletedTopic = common.GetEnv(KAFKA_COURT_DELETE_TOPIC_ENV, "court-deleted")
	slotBookedTopic = common.GetEnv(KAFKA_SLOT_BOOKED_TOPIC_ENV, "slot-booked")
	retry, _ := strconv.Atoi(common.GetEnv(KAFKA_RETRY_ENV, "5"))

	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = retry
	config.Producer.Return.Successes = true
	if _, err := sarama.NewClient([]string{fmt.Sprintf("%s:%s", host, port)}, config); err != nil {
		return nil, err
	}
	producer, err := sarama.NewSyncProducer([]string{fmt.Sprintf("%s:%s", host, port)}, config)
	if err != nil {
		return nil, err
	}
	return producer, nil
}

func notifySlotBookedEvent(userID, courtID, slotID string) error {
	if err := validateMessageQueueProducer(messageQueueProducer); err != nil {
		return err
	}
	slotBookedObj := SlotBookedNotification{
		UserId:  userID,
		CourtId: courtID,
		SlotId:  slotID,
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

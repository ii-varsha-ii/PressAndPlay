package main

import (
	"fmt"
	"github.com/adarshsrinivasan/PressAndPlay/libraries/common"
	"strconv"

	"github.com/Shopify/sarama"
	"github.com/sirupsen/logrus"
)

const (
	KAFKA_HOST_ENV              = "KAFKA_HOST"
	KAFKA_PORT_ENV              = "KAFKA_PORT"
	KAFKA_USER_DELETE_TOPIC_ENV = "KAFKA_USER_DELETE_TOPIC"
	KAFKA_RETRY_ENV             = "KAFKA_RETRY"
)

var (
	topic string
)

func newKafkaHandler() (sarama.SyncProducer, error) {
	host := common.GetEnv(KAFKA_HOST_ENV, "localhost")
	port := common.GetEnv(KAFKA_PORT_ENV, "9092")
	topic = common.GetEnv(KAFKA_USER_DELETE_TOPIC_ENV, "court-deleted")
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

func notifyCourtDeletedEvent(courtID string) error {
	if err := validateMessageQueueProducer(messageQueueProducer); err != nil {
		return err
	}
	msg := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.StringEncoder(courtID),
	}
	partition, offset, err := messageQueueProducer.SendMessage(msg)
	if err != nil {
		return err
	}
	logrus.Infof("Sent message: %s, on topic %s. partition: %v, offset: %v", courtID, topic, partition, offset)
	return nil
}

func validateMessageQueueProducer(producer sarama.SyncProducer) error {
	if producer == nil {
		return fmt.Errorf("kafka producer connection not initialized")
	}
	return nil
}

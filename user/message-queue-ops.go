package main

import (
	"fmt"
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
	host := getEnv(KAFKA_HOST_ENV, "localhost")
	port := getEnv(KAFKA_PORT_ENV, "9092")
	topic = getEnv(KAFKA_USER_DELETE_TOPIC_ENV, "user-deleted")
	retry, _ := strconv.Atoi(getEnv(KAFKA_RETRY_ENV, "5"))

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

func notifyUserDeletedEvent(userID string) error {
	if err := validateMessageQueueProducer(messageQueueProducer); err != nil {
		return err
	}
	msg := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.StringEncoder(userID),
	}
	partition, offset, err := messageQueueProducer.SendMessage(msg)
	if err != nil {
		return err
	}
	logrus.Infof("Sent message: %s, on topic %s. partition: %v, offset: %v", userID, topic, partition, offset)
	return nil
}

func validateMessageQueueProducer(producer sarama.SyncProducer) error {
	if producer == nil {
		return fmt.Errorf("kafka producer connection not initialized")
	}
	return nil
}

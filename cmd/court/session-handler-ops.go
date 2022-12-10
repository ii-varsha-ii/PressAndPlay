package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/adarshsrinivasan/PressAndPlay/libraries/common"
	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
)

const (
	REDIS_HOST_ENV     = "REDIS_HOST"
	REDIS_PORT_ENV     = "REDIS_PORT"
	REDIS_PASSWORD_ENV = "REDIS_PASSWORD"
	REDIS_DB_ENV       = "REDIS_DB"
)

type sessionBody struct {
	SessionID     string    `json:"session_id"`
	UserID        string    `json:"user_id"`
	LastLoginTime time.Time `json:"last_login_time"`
}

func (s *sessionBody) toString() string {
	converted, _ := json.Marshal(s)
	return string(converted)
}

func (s *sessionBody) toStruct(stringSessionDetails string) {
	json.Unmarshal([]byte(stringSessionDetails), s)
}

func newRedisHandler() (*redis.Client, error) {
	host := common.GetEnv(REDIS_HOST_ENV, "localhost")
	port := common.GetEnv(REDIS_PORT_ENV, "6379")
	password := common.GetEnv(REDIS_PASSWORD_ENV, "admin")
	dbName, _ := strconv.Atoi(common.GetEnv(REDIS_DB_ENV, "0"))

	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", host, port),
		Password: password,
		DB:       dbName,
	})
	return client, client.Ping(ctx).Err()
}

func validateSessionID(sessionID string) bool {
	if err := verifyRedisConnection(sessionCLient); err != nil {
		logrus.Errorf("validateSessionID(%s): exception. %v", sessionID, err)
		return false
	}
	_, err := sessionCLient.Get(ctx, sessionID).Result()
	return err == nil
}

func getUserIdFromSession(sessionID string) string {
	if err := verifyRedisConnection(sessionCLient); err != nil {
		logrus.Errorf("getUserIdFromSession(%s): exception. %v", sessionID, err)
		return ""
	}
	value, err := sessionCLient.Get(ctx, sessionID).Result()
	if err != nil {
		logrus.Errorf("getUserIdFromSession(%s): exception. %v", sessionID, err)
		return ""
	}
	session := sessionBody{}
	session.toStruct(value)
	return session.UserID
}

func verifyRedisConnection(redisClient *redis.Client) error {
	if redisClient == nil {
		return fmt.Errorf("redis connection not initialized")
	}
	return nil
}

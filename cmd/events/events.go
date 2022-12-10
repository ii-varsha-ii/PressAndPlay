package main

import (
	"context"
	"fmt"
	"github.com/Shopify/sarama"
	"github.com/adarshsrinivasan/PressAndPlay/libraries/common"
	"github.com/adarshsrinivasan/PressAndPlay/libraries/proto"
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"github.com/uptrace/bun"
	"google.golang.org/grpc"
	"log"
	"net"
	"net/http"
	"strconv"
)

const (
	SERVICE_NAME              = "events"
	GRPC_SERVER_PORT_ENV      = "GRPC_SERVER_PORT"
	GRPC_USER_CLIENT_HOST_ENV = "GRPC_USER_CLIENT_HOST"
	GRPC_USER_CLIENT_PORT_ENV = "GRPC_USER_CLIENT_PORT"
	HTTP_SERVER_HOST_ENV      = "HTTP_SERVER_HOST"
	HTTP_SERVER_PORT_ENV      = "HTTP_SERVER_PORT"
)

var (
	dbClient              *bun.DB
	sessionCLient         *redis.Client
	err                   error
	ctx                   = context.Background()
	messageQueueProducer  sarama.SyncProducer
	httpRouter            *mux.Router
	gRPCServerPort, _     = strconv.Atoi(common.GetEnv(GRPC_SERVER_PORT_ENV, "50005"))
	gRPCUserClientHost    = common.GetEnv(GRPC_USER_CLIENT_HOST_ENV, "localhost")
	gRPCUserClientPort, _ = strconv.Atoi(common.GetEnv(GRPC_USER_CLIENT_PORT_ENV, "50011"))
	httpServerHost        = common.GetEnv(HTTP_SERVER_HOST_ENV, "localhost")
	httpServerPort, _     = strconv.Atoi(common.GetEnv(HTTP_SERVER_PORT_ENV, "50004"))
	gRPCUserClient        proto.UserClient
)

func initializeDB() error {
	dbClient, err = newPostgresClient()
	if err != nil {
		return fmt.Errorf("exception while initializing postgres client. %v", err)
	}
	if err := createUserTable(); err != nil {
		return fmt.Errorf("exception while creating user tabel. %v", err)
	}
	return nil
}

func initializeSessionHandler() error {
	sessionCLient, err = newRedisHandler()
	if err != nil {
		return fmt.Errorf("exception while initializing redis client. %v", err)
	}
	return nil
}

func initializeMessageQueueProducer() error {
	messageQueueProducer, err = newKafkaHandler()
	if err != nil {
		return fmt.Errorf("exception while initializing kafka producer client. %v", err)
	}
	return nil
}

func initializeHTTPRouter() error {
	initializeMuxRoutes()
	if httpRouter == nil {
		return fmt.Errorf("http router not initialized")
	}
	return nil
}

func initializeGRPCServer() error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", gRPCServerPort))
	// Check for errors
	if err != nil {
		return fmt.Errorf("exception while initializing grpc server. %v", err)
	}
	// Instantiate the server
	s := grpc.NewServer()
	proto.RegisterUserServer(s, &userGRPCService{})
	logrus.Infof("grpc server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		return fmt.Errorf("exception while starting grpc server. %v", err)
	}
	return nil
}

func initializeGRPCUserClient() error {
	var conn *grpc.ClientConn
	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", gRPCUserClientHost, gRPCUserClientPort), grpc.WithInsecure())
	if err != nil {
		return fmt.Errorf("exception while initializing grpc client. %v", err)
	}

	gRPCUserClient = proto.NewUserClient(conn)
	return nil
}

func initialize() error {
	go initializeGRPCServer()
	if err := initializeDB(); err != nil {
		return fmt.Errorf("db initialization error. %v", err)
	}
	if err := initializeSessionHandler(); err != nil {
		return fmt.Errorf("session handler initialization error. %v", err)
	}
	if err := initializeMessageQueueProducer(); err != nil {
		return fmt.Errorf("message queue producer initialization error. %v", err)
	}
	if err := initializeHTTPRouter(); err != nil {
		return fmt.Errorf("http router initialization error. %v", err)
	}
	if err := initializeGRPCUserClient(); err != nil {
		return fmt.Errorf("gRPC client initialization error. %v", err)
	}

	return nil
}

func main() {
	if err := initialize(); err != nil {
		logrus.Panic(err)
	}
	log.Fatal(http.ListenAndServe(fmt.Sprintf("%s:%d", httpServerHost, httpServerPort), httpRouter))
}

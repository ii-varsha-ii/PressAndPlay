package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"strconv"

	"github.com/Shopify/sarama"
	"github.com/adarshsrinivasan/PressAndPlay/libraries/common"
	"github.com/adarshsrinivasan/PressAndPlay/libraries/proto"
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
	"google.golang.org/grpc"
)

const (
	SERVICE_NAME               = "court"
	GRPC_EVENT_CLIENT_HOST_ENV = "GRPC_EVENT_CLIENT_HOST"
	GRPC_EVENT_CLIENT_PORT_ENV = "GRPC_EVENT_CLIENT_PORT"
	GRPC_USER_CLIENT_HOST_ENV  = "GRPC_USER_CLIENT_HOST"
	GRPC_USER_CLIENT_PORT_ENV  = "GRPC_USER_CLIENT_PORT"
	GRPC_SERVER_PORT_ENV       = "GRPC_SERVER_PORT"
	HTTP_SERVER_HOST_ENV       = "HTTP_SERVER_HOST"
	HTTP_SERVER_PORT_ENV       = "HTTP_SERVER_PORT"
)

var (
	dbClient               *mongo.Collection
	sessionCLient          *redis.Client
	err                    error
	ctx                    = context.Background()
	messageQueueProducer   sarama.SyncProducer
	messageQueueConsumer   sarama.Consumer
	httpRouter             *mux.Router
	gRPCServerPort, _      = strconv.Atoi(common.GetEnv(GRPC_SERVER_PORT_ENV, "50003"))
	httpServerHost         = common.GetEnv(HTTP_SERVER_HOST_ENV, "localhost")
	httpServerPort, _      = strconv.Atoi(common.GetEnv(HTTP_SERVER_PORT_ENV, "50002"))
	gRPCEventClientHost    = common.GetEnv(GRPC_EVENT_CLIENT_HOST_ENV, "localhost")
	gRPCEventClientPort, _ = strconv.Atoi(common.GetEnv(GRPC_EVENT_CLIENT_PORT_ENV, "50005"))
	gRPCUserClientHost     = common.GetEnv(GRPC_USER_CLIENT_HOST_ENV, "localhost")
	gRPCUserClientPort, _  = strconv.Atoi(common.GetEnv(GRPC_USER_CLIENT_PORT_ENV, "50011"))
	gRPCEventClient        proto.EventsClient
	gRPCUserClient         proto.UserClient
)

func initializeDB() error {
	dbClient, err = newMongoClient()
	if err != nil {
		return fmt.Errorf("exception while initializing mongo client. %v", err)
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

func initializeMessageQueue() error {
	messageQueueProducer, messageQueueConsumer, err = newKafkaHandler()
	if err != nil {
		return fmt.Errorf("exception while initializing kafka producer client. %v", err)
	}
	userDeletedTopic := common.GetEnv(KAFKA_USER_DELETE_TOPIC_ENV, "user-deleted")
	if err := initializeConsumers(userDeletedTopic, handleUserDeletedNotification); err != nil {
		return fmt.Errorf("exception while initializing kafka consumer for topic %s. %v", userDeletedTopic, err)
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
	proto.RegisterCourtServer(s, &courtGRPCService{})
	logrus.Infof("grpc server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		return fmt.Errorf("exception while starting grpc server. %v", err)
	}
	return nil
}

func initializeGRPCEventClient() error {
	var conn *grpc.ClientConn
	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", gRPCEventClientHost, gRPCEventClientPort), grpc.WithInsecure())
	if err != nil {
		return fmt.Errorf("exception while initializing grpc client. %v", err)
	}

	gRPCEventClient = proto.NewEventsClient(conn)
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
	if err := initializeMessageQueue(); err != nil {
		return fmt.Errorf("message queue producer initialization error. %v", err)
	}
	if err := initializeHTTPRouter(); err != nil {
		return fmt.Errorf("http router initialization error. %v", err)
	}
	if err := initializeGRPCUserClient(); err != nil {
		return fmt.Errorf("gRPC client initialization error. %v", err)
	}
	//if err := initializeGRPCEventClient(); err != nil {
	//	return fmt.Errorf("gRPC client initialization error. %v", err)
	//}
	return nil
}

func main() {
	if err := initialize(); err != nil {
		logrus.Panic(err)
	}
	log.Fatal(http.ListenAndServe(fmt.Sprintf("%s:%d", httpServerHost, httpServerPort), httpRouter))
}

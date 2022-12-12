#
# Copyright 2018 Zededa Inc.
#
# Makefile for building and running zedcloud components in docker
#

.PHONY: SERVICES DEBUG_SERVICES COV_SERVICES ztests srvs/zswagger-ui/Dockerfile all docker debug docker-debug push test zservices clean check-version man-pages coverage docker-cov

SERVICES := user \
			court \
			events

DOCKER_BUILD := DOCKER_BUILDKIT=1 docker build
DOCKER_PUSH := DOCKER_BUILDKIT=1 docker push
DOCKERUSER=adarshzededa

$(SERVICES): gen-proto Dockerfile.service
	@echo "Building $@ ..."
	$(DOCKER_BUILD) --build-arg service=$@ -t $(DOCKERUSER)/pressandplay-$@:latest -f Dockerfile.service .
	@echo "Pushing $@ ..."
	$(DOCKER_PUSH) $(DOCKERUSER)/pressandplay-$@:latest
	: $@: Succeeded

gen-proto:
	@echo "Generating Proto..."
	protoc --proto_path=$(GOPATH)/src/github.com/adarshsrinivasan/PressAndPlay/libraries/proto --plugin=$(GOPATH)/bin/protoc-gen-go-grpc --go_out=$(GOPATH)/src --go-grpc_out=$(GOPATH)/src common.proto
	protoc --proto_path=$(GOPATH)/src/github.com/adarshsrinivasan/PressAndPlay/libraries/proto --plugin=$(GOPATH)/bin/protoc-gen-go-grpc --go_out=$(GOPATH)/src --go-grpc_out=$(GOPATH)/src user.proto
	protoc --proto_path=$(GOPATH)/src/github.com/adarshsrinivasan/PressAndPlay/libraries/proto --plugin=$(GOPATH)/bin/protoc-gen-go-grpc --go_out=$(GOPATH)/src --go-grpc_out=$(GOPATH)/src court.proto
	protoc --proto_path=$(GOPATH)/src/github.com/adarshsrinivasan/PressAndPlay/libraries/proto --plugin=$(GOPATH)/bin/protoc-gen-go-grpc --go_out=$(GOPATH)/src --go-grpc_out=$(GOPATH)/src events.proto
	: $@: Succeeded

deploy-docker:
	@echo "Deploying PressAndPlay on docker..."
	cd deployment/docker && docker-compose up -d
	: $@: Succeeded

undeploy-docker:
	@echo "Removing PressAndPlay from docker..."
	cd deployment/docker && docker-compose down -v
	: $@: Succeeded

deploy-kubernetes:
	@echo "Deploying PressAndPlay on kubernetes..."

	kubectl apply -f deployment/kuberneter/mongo/mongo-persistent-volume.yaml
	kubectl apply -f deployment/kuberneter/mongo/mongodb-deployment.yaml
	kubectl apply -f deployment/kuberneter/mongo/mongodb-svc.yaml

	kubectl apply -f deployment/kuberneter/postgres/postgres-persistent-volume.yaml
	kubectl apply -f deployment/kuberneter/postgres/postgres-deployment.yaml
	kubectl apply -f deployment/kuberneter/postgres/postgres-svc.yaml

	kubectl apply -f deployment/kuberneter/redis/redis-persistent-volume.yaml
	kubectl apply -f deployment/kuberneter/redis/redis-deployment.yaml
	kubectl apply -f deployment/kuberneter/redis/redis-svc.yaml

	kubectl apply -f deployment/kuberneter/kafka/zookeeper-deployment.yaml
	kubectl apply -f deployment/kuberneter/kafka/zookeeper-svc.yaml
	kubectl apply -f deployment/kuberneter/kafka/kafka-deployment.yaml
	kubectl apply -f deployment/kuberneter/kafka/kafka-svc.yaml

	sleep 60
	kubectl apply -f deployment/kuberneter/user/user-deployment.yaml
	kubectl apply -f deployment/kuberneter/user/user-svc.yaml

	sleep 30
	kubectl apply -f deployment/kuberneter/court/court-deployment.yaml
	kubectl apply -f deployment/kuberneter/court/court-svc.yaml

	sleep 30
	kubectl apply -f deployment/kuberneter/events/events-deployment.yaml
	kubectl apply -f deployment/kuberneter/events/events-svc.yaml

	kubectl apply -f deployment/kuberneter/pressandplay-ingress.yaml
	: $@: Succeeded

undeploy-kubernetes:
	@echo "Deploying PressAndPlay on kubernetes..."

	kubectl delete -f deployment/kuberneter/pressandplay-ingress.yaml

	kubectl delete -f deployment/kuberneter/mongo/mongodb-svc.yaml
	kubectl delete -f deployment/kuberneter/mongo/mongodb-deployment.yaml
	kubectl delete -f deployment/kuberneter/mongo/mongo-persistent-volume.yaml

	kubectl delete -f deployment/kuberneter/postgres/postgres-svc.yaml
	kubectl delete -f deployment/kuberneter/postgres/postgres-deployment.yaml
	kubectl delete -f deployment/kuberneter/postgres/postgres-persistent-volume.yaml

	kubectl delete -f deployment/kuberneter/redis/redis-svc.yaml
	kubectl delete -f deployment/kuberneter/redis/redis-deployment.yaml
	kubectl delete -f deployment/kuberneter/redis/redis-persistent-volume.yaml

	kubectl delete -f deployment/kuberneter/kafka/zookeeper-svc.yaml
	kubectl delete -f deployment/kuberneter/kafka/zookeeper-deployment.yaml
	kubectl delete -f deployment/kuberneter/kafka/kafka-svc.yaml
	kubectl delete -f deployment/kuberneter/kafka/kafka-deployment.yaml

	kubectl delete -f deployment/kuberneter/user/user-svc.yaml
	kubectl delete -f deployment/kuberneter/user/user-deployment.yaml

	kubectl delete -f deployment/kuberneter/court/court-svc.yaml
	kubectl delete -f deployment/kuberneter/court/court-deployment.yaml

	kubectl delete -f deployment/kuberneter/events/events-svc.yaml
	kubectl delete -f deployment/kuberneter/events/events-deployment.yaml

	: $@: Succeeded

help:
	@echo "List of available targets:"
	@echo "all: build all services in docker containers"
	@echo "      make"
	@echo "      make all"
	@echo "      make build"
	@echo "<target> : build and push the specific service"
	@echo "      make user"
	@echo "      make court"
	@echo "      make events"
	@echo "gen-proto: generate proto"
	@echo "deploy-docker: deploy PressAndPlay using Docker"
	@echo "undeploy-docker: undeploy PressAndPlay from Docker"


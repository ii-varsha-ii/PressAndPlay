#
# Copyright 2018 Zededa Inc.
#
# Makefile for building and running zedcloud components in docker
#

.PHONY: SERVICES DEBUG_SERVICES COV_SERVICES ztests srvs/zswagger-ui/Dockerfile all docker debug docker-debug push test zservices clean check-version man-pages coverage docker-cov

SERVICES := user \
			court

DOCKER_BUILD := DOCKER_BUILDKIT=1 docker build
DOCKER_PUSH := DOCKER_BUILDKIT=1 docker push
DOCKERUSER=adarshzededa

$(SERVICES): gen-proto Dockerfile.service
	@echo "Building $@ ..."
	$(DOCKER_BUILD) --build-arg service=$@ -t $(DOCKERUSER)/pressandplay-$@:latest -f Dockerfile.service .
#	@echo "Pushing $@ ..."
#	$(DOCKER_PUSH) $(DOCKERUSER)/pressandplay-$@:latest
	: $@: Succeeded

gen-proto:
	@echo "Generating Proto..."
	protoc --proto_path=$(GOPATH)/src/github.com/adarshsrinivasan/PressAndPlay/libraries/proto --plugin=$(GOPATH)/bin/protoc-gen-go-grpc --go_out=$(GOPATH)/src --go-grpc_out=$(GOPATH)/src common.proto
	protoc --proto_path=$(GOPATH)/src/github.com/adarshsrinivasan/PressAndPlay/libraries/proto --plugin=$(GOPATH)/bin/protoc-gen-go-grpc --go_out=$(GOPATH)/src --go-grpc_out=$(GOPATH)/src user.proto
	protoc --proto_path=$(GOPATH)/src/github.com/adarshsrinivasan/PressAndPlay/libraries/proto --plugin=$(GOPATH)/bin/protoc-gen-go-grpc --go_out=$(GOPATH)/src --go-grpc_out=$(GOPATH)/src court.proto
	protoc --proto_path=$(GOPATH)/src/github.com/adarshsrinivasan/PressAndPlay/libraries/proto --plugin=$(GOPATH)/bin/protoc-gen-go-grpc --go_out=$(GOPATH)/src --go-grpc_out=$(GOPATH)/src events.proto
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
	@echo "gen-proto: generate proto"

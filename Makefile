#
# Copyright 2018 Zededa Inc.
#
# Makefile for building and running zedcloud components in docker
#

.PHONY: SERVICES DEBUG_SERVICES COV_SERVICES ztests srvs/zswagger-ui/Dockerfile all docker debug docker-debug push test zservices clean check-version man-pages coverage docker-cov

SERVICES := user \

DOCKER_BUILD := DOCKER_BUILDKIT=1 docker build
DOCKERUSER=adarshzededa

$(SERVICES): Dockerfile.service
	@echo "Building $@ ..."
	$(DOCKER_BUILD) --build-arg service=$@ -t $(DOCKERUSER)/pressandplay-$@:latest -f Dockerfile.service .
	: $@: Succeeded

proto:
	@echo "Generating Proto..."
	protoc --plugin=$(GOPATH)/bin/protoc-gen-go-grpc --go_out=user --go-grpc_out=user user/proto/user.proto
	: $@: Succeeded

help:
	@echo "List of available targets:"
	@echo "all: build all services in docker containers"
	@echo "      make"
	@echo "      make all"
	@echo "      make build"
	@echo "debug: build all services in docker containers with debug enabled"
	@echo "<target> : build the specific service"
	@echo "      make gilas"
	@echo "      make seine"
	@echo "<target>-debug : build the specific service with debug enabled"
	@echo "      make gilas-debug"
	@echo "      make seine-debug"
	@echo "run-<target> : build and run the specific service"
	@echo "      make run-gilas"
	@echo "      make run-seine"
	@echo "all-in-one: build a container that has all services"
	@echo "run-all-in-one: build and run container that has all services"
	@echo "update-dependencies: scan all of our dependencies and raise a PR updating outdated dependencies"
	@echo "		'GITHUB_TOKEN' environment variable must be set before running this command"
	@echo "scan-dep-vulnerability: scan our golang dependencies to check for vulnerabilities"
	@echo "yetus: run branch-level yetus tests"
	@echo "gen-dep-report: generate a report of dependent third party modules"

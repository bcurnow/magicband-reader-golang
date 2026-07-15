#!/usr/bin/make

SHELL := /bin/bash
currentDir := $(dir $(abspath $(lastword $(MAKEFILE_LIST))))
imageName := $(notdir $(patsubst %/,%,$(currentDir)))
currentUser := $(shell id -u)
currentGroup := $(shell id -g)
binaryName := magicband-reader
# Must match the dev_image WORKDIR in the Dockerfile.
containerWorkdir := /workspace

.PHONY: build-docker build-docker-dev run-docker dev run-docker-prod prod build format vet lint test run tidy local clear lr clean

build-docker:
	docker buildx build \
	  --platform linux/arm/v6 \
	  --build-arg USER_ID=${currentUser} \
	  --build-arg GROUP_ID=${currentGroup} \
	  -t ${imageName}:production  \
	  ${currentDir}

build-docker-dev:
	docker buildx build \
	  --target dev_image \
	  --platform linux/arm/v6 \
	  --build-arg USER_ID=${currentUser} \
	  --build-arg GROUP_ID=${currentGroup} \
	  -t ${imageName}:latest  \
	  ${currentDir}

# Mounts the repo itself (the module doesn't need to live at a GOPATH-derived path) at the
# same path the Dockerfile builds from, so edits on the host are reflected live in the container.
run-docker:
	docker run --platform linux/arm/v6 -it --privileged --mount src=/dev,target=/dev,type=bind --mount src=${currentDir}sounds,target=/sounds,type=bind --mount src="${currentDir}",target=${containerWorkdir},type=bind ${imageName}:latest /bin/bash

dev: run-docker

run-docker-prod:
	docker run --platform linux/arm/v6 -d --privileged --mount src=/dev,target=/dev,type=bind --mount src=${currentDir}sounds,target=/sounds,type=bind ${imageName}:production

prod: run-docker-prod

build:
# Build for armv6 (which is what the RPi0 has)
# Also, d2xx doesn't work with CGO_ENABLED so set the build flag to skip it, we don't use that driver
	env GOARCH=arm GOOS=linux GOARM=6 CGO_ENABLED=1 go build -tags no_d2xx -o bin/${binaryName}

format:
	gofmt -l -w -s .

vet:
	go vet -tags no_d2xx ./...

lint:
	golangci-lint run --build-tags=no_d2xx

test:
	go test -tags no_d2xx ./...

run:
	sudo bin/${binaryName} ${READER_ARGS}

tidy:
	go mod tidy

local: clear tidy format vet test build

clear:
	clear

lr: local run

clean:
	rm -rf bin/

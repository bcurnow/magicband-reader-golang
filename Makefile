#!/usr/bin/make

SHELL := /bin/bash
currentDir := $(dir $(abspath $(lastword $(MAKEFILE_LIST))))
imageName := $(notdir $(patsubst %/,%,$(dir $(currentDir))))
rootDir := $(abspath ${currentDir}../../../../)
currentUser := $(shell id -u)
currentGroup := $(shell id -g)
binaryName := magicband-reader

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

run-docker:
	docker run --platform linux/arm/v6 -it --privileged --mount src=/dev,target=/dev,type=bind --mount src=${currentDir}sounds,target=/sounds,type=bind --mount src="${rootDir}",target=/go,type=bind ${imageName}:latest /bin/bash

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

run:
	sudo bin/${binaryName} ${READER_ARGS}

tidy:
	go mod tidy

local: clear tidy format build

clear:
	clear

lr: local run

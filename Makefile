#!/usr/bin/make

SHELL := /bin/bash
currentDir := $(dir $(abspath $(lastword $(MAKEFILE_LIST))))
imageName := $(notdir $(patsubst %/,%,$(dir $(currentDir))))
rootDir := $(abspath ${currentDir}../../../../)
currentUser := $(shell id -u)
currentGroup := $(shell id -g)

build-docker:
	docker buildx build --no-cache \
	  --platform linux/arm/v6 \
	  --build-arg USER_ID=${currentUser} \
	  --build-arg GROUP_ID=${currentGroup} \
	  -t ${imageName}:latest  \
	  ${currentDir}

bd: build-docker

run-docker:
	docker run --platform linux/arm/v6 -it --privileged --mount src=/dev,target=/dev,type=bind --mount src=${currentDir}sounds,target=/sounds,type=bind --mount src="${rootDir}",target=/go,type=bind ${imageName}:latest /bin/bash

rd: run-docker

build:
	# Build for armv6 (which is what the RPi0 has)
	# Also, d2xx doesn't work with CGO_ENABLED so set the build flag to skip it, we don't use that driver
	env GOARCH=arm GOOS=linux GOARM=6 CGO_ENABLED=1 go build -tags no_d2xx -o bin/${imageName}

format:
	gofmt -l -w -s .

sudo:
	sudo bin/${imageName}

sudodb:
	sudo bin/${imageName} --log-level debug

tidy:
	go mod tidy

local: clear tidy format build

clear:
	clear

lr: local sudo

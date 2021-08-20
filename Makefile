#!/usr/bin/make

SHELL := /bin/bash
currentDir := $(dir $(abspath $(lastword $(MAKEFILE_LIST))))
imageName := $(notdir $(patsubst %/,%,$(dir $(currentDir))))
rootDir := $(abspath ${currentDir}../../../../)
currentUser := $(shell id -u)
currentGroup := $(shell id -g)

build-docker:
	docker build \
	  --build-arg USER_ID=${currentUser} \
	  --build-arg GROUP_ID=${currentGroup} \
	  -t ${imageName}:latest  \
	  ${currentDir}

bd: build-docker

run-docker:
	docker run -it --privileged --group-add audio --mount src=/dev,target=/dev,type=bind --mount src=${currentDir}sounds,target=/sounds,type=bind --mount src="${rootDir}",target=/go,type=bind ${imageName}:latest /bin/bash

rd: run-docker

build:
	go build -o bin/magicband-reader

format:
	gofmt -l -w -s .

sudo:
	sudo bin/magicband-reader

tidy:
	go mod tidy

local: tidy format build

lr: local sudo

.PHONY: build

MKFILE_PATH := $(abspath $(lastword $(MAKEFILE_LIST)))
MKFILE_DIR := $(dir $(MKFILE_PATH))

build:
	CGO_ENABLED=0 go build -v -o ${MKFILE_DIR}bin/consuldemo ${MKFILE_DIR}cmd/consul/main.go
	CGO_ENABLED=0 go build -v -o ${MKFILE_DIR}bin/meshdemo ${MKFILE_DIR}cmd/mesh/main.go

build_docker:
	docker build -t megaease/consuldemo:latest .

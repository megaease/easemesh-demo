.PHONY: build

MKFILE_PATH := $(abspath $(lastword $(MAKEFILE_LIST)))
MKFILE_DIR := $(dir $(MKFILE_PATH))

build:
	CGO_ENABLED=0 go build -v -o ${MKFILE_DIR}bin/consuldemo ${MKFILE_DIR}cmd/consul/main.go
	CGO_ENABLED=0 go build -v -o ${MKFILE_DIR}bin/meshdemo ${MKFILE_DIR}cmd/mesh/main.go
	CGO_ENABLED=0 go build -v -o ${MKFILE_DIR}bin/http2kafka ${MKFILE_DIR}cmd/http2kafka/main.go

build_docker:
	docker buildx build --platform linux/amd64 --load -t megaease/consuldemo:latest .
	docker tag megaease/consuldemo:latest megaease/consuldemo:canary_version

build_docker_http2kafka:
	docker build -t megaease/http2kafka:latest -f Dockerfile.http2kafka .

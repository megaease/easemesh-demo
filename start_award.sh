#!/usr/bin/env bash

CONSUL_ADDRESS=127.0.0.1:8500 \
	POD_IP=127.0.0.1 \
	POD_PORT=10095 \
	SERVICE_NAME=award \
	INSTANCE_ID=award-001 \
	bin/consuldemo

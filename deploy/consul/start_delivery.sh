#!/usr/bin/env bash

CONSUL_ADDRESS=127.0.0.1:8500 \
	POD_IP=127.0.0.1 \
	POD_PORT=10090 \
	SERVICE_NAME=delivery \
	INSTANCE_ID=delivery-001 \
	bin/consuldemo

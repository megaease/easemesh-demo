#!/usr/bin/env bash

CONSUL_ADDRESS=127.0.0.1:8500 \
	POD_IP=127.0.0.1 \
	POD_PORT=10092 \
	SERVICE_NAME=order \
	INSTANCE_ID=order-001 \
	bin/consuldemo

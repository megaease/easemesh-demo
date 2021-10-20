#!/usr/bin/env bash

CONSUL_ADDRESS=127.0.0.1:8500 \
	POD_IP=127.0.0.1 \
	POD_PORT=10091 \
	SERVICE_NAME=restaurant \
	INSTANCE_ID=restaurant-001 \
	bin/consuldemo

#!/usr/bin/env bash

# 0. Change to namespace mesh-service
kcd mesh-service

# 1. Apply all services
kubectl apply -f deploy/mesh/k8s_order.yaml
kubectl apply -f deploy/mesh/k8s_award.yaml
kubectl apply -f deploy/mesh/k8s_restaurant.yaml
kubectl apply -f deploy/mesh/k8s_delivery.yaml

# 2. Show EaseMesh Registry
emctl get serviceinstance

# 3. Request service
curl http://192.168.0.200:31098/ -d '{"order_id": "abc1234", "food": "bread"}' | jq

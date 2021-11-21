#!/usr/bin/env bash

# 0. Change to namespace default
kcd default

# 1. Show Order, Restaurant, Delivery, Award
kubectl get pod

# 2. Show Consul registry via browser

# 3. Show EaseMesh registry
emctl get serviceinstance

# 4. Try to request service
curl http://192.168.0.200:31121/ -d '{"order_id": "abc1234", "food": "bread"}' | jq

# ---

# 5. Create syncer
egctl --server 192.168.0.200:30146 object create -f deploy/consul-service-registry.yaml

# 6. Show Consul registry via browser

# 7. Show EaseMesh registry
emctl get serviceinstance

# 8. Request service again
curl http://192.168.0.200:31121/ -d '{"order_id": "abc1234", "food": "bread"}' | jq

# 9. Clean syncer
egctl --server 192.168.0.200:30146 object delete consul-service-registry

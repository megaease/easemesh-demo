#!/usr/bin/env bash

emctl delete servicecanary delivery-mesh-beijing
emctl delete servicecanary restaurant-mesh-beijing
emctl delete servicecanary refund-android

kubectl delete -f deploy/mesh/k8s_delivery_beijing.yaml
kubectl delete -f deploy/mesh/k8s_restaurant_beijing.yaml
kubectl delete -f deploy/mesh/k8s_delivery_android.yaml
kubectl delete -f deploy/mesh/k8s_restaurant_android.yaml

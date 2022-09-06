#!/usr/bin/env bash

# Service Canary
emctl delete servicecanary beijing
emctl delete servicecanary android
emctl delete servicecanary beijing-android

# Primary
kubectl delete -f deploy/mesh/k8s_order.yaml
kubectl delete -f meshdemo-restaurant/k8s_restaurant.yaml
kubectl delete -f deploy/mesh/k8s_delivery.yaml

# beijing
kubectl delete -f deploy/mesh/k8s_delivery_beijing.yaml

# android
kubectl delete -f deploy/mesh/k8s_delivery_android.yaml

# beijing-android
kubectl delete -f meshdemo-restaurant/k8s_restaurant_android.yaml
kubectl delete -f meshdemo-restaurant/k8s_restaurant_beijing_android.yaml

# Namespace and mesh config
emctl delete -f deploy/mesh/easemesh_order.yaml
emctl delete -f meshdemo-restaurant/easemesh_restaurant.yaml
emctl delete -f deploy/mesh/easemesh_delivery.yaml
emctl delete -f deploy/mesh/easemesh_tenant.yaml

kubectl create ns mesh-service
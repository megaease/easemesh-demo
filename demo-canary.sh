#!/usr/bin/env bash

# 1. Deploy primary stack
kubectl apply -f deploy/mesh/k8s_order.yaml
kubectl apply -f meshdemo-restaurant/k8s_restaurant.yaml
kubectl apply -f deploy/mesh/k8s_delivery.yaml

# Primary traffic
curl http://127.0.0.1:30188/ -d '{"order_id": "abc1234", "food": "bread"}'

# 2. Deploy canary Delivery Beijing (Green)
kubectl apply -f deploy/mesh/k8s_delivery_beijing.yaml

echo '
apiVersion: mesh.megaease.com/v1alpha1
kind: ServiceCanary
metadata:
  name: beijing
spec:
  priority: 3
  selector:
    matchServices: [delivery-mesh]
    matchInstanceLabels: {release: delivery-mesh-beijing}
  trafficRules:
    headers:
      X-Location:
        exact: Beijing
' | emctl apply -f -

# Beijing traffic
curl http://127.0.0.1:30188/ -d '{"order_id": "abc1234", "food": "bread"}' -H 'X-Location: Beijing'

# 3. Deploy canary Restaurant/Delivery Android
kubectl apply -f deploy/mesh/k8s_delivery_android.yaml
kubectl apply -f meshdemo-restaurant/k8s_restaurant_android.yaml

echo '
apiVersion: mesh.megaease.com/v1alpha1
kind: ServiceCanary
metadata:
  name: android
spec:
  priority: 2
  selector:
    matchServices: [restaurant-mesh, delivery-mesh]
    matchInstanceLabels: {release: refund-android}
  trafficRules:
    headers:
      X-Phone-Os:
        exact: Android
' | emctl apply -f -

# Android traffic
curl http://127.0.0.1:30188/ -d '{"order_id": "abc1234", "food": "bread"}' -H 'X-Phone-Os: Android'

# 4. Deploy canary Restaurant Beijing&Android
kubectl apply -f meshdemo-restaurant/k8s_restaurant_beijing_android.yaml

echo '
apiVersion: mesh.megaease.com/v1alpha1
kind: ServiceCanary
metadata:
  name: beijing-android
spec:
  priority: 1
  selector:
    matchServices: [restaurant-mesh]
    matchInstanceLabels: {release: restaurant-mesh-beijing-android}
  trafficRules:
    headers:
      X-Location:
        exact: Beijing
      X-Phone-Os:
        exact: Android
' | emctl apply -f -

# Beijing&Android traffic
curl http://127.0.0.1:30188/ -d '{"order_id": "abc1234", "food": "bread"}' -H 'X-Location: Beijing' -H 'X-Phone-Os: Android'

# 5. Change priorty

echo '
apiVersion: mesh.megaease.com/v1alpha1
kind: ServiceCanary
metadata:
  name: beijing-android
spec:
  priority: 2
  selector:
    matchServices: [restaurant-mesh]
    matchInstanceLabels: {release: restaurant-mesh-beijing-android}
  trafficRules:
    headers:
      X-Location:
        exact: Beijing
      X-Phone-Os:
        exact: Android
' | emctl apply -f -

echo '
apiVersion: mesh.megaease.com/v1alpha1
kind: ServiceCanary
metadata:
  name: android
spec:
  priority: 1
  selector:
    matchServices: [restaurant-mesh, delivery-mesh]
    matchInstanceLabels: {release: refund-android}
  trafficRules:
    headers:
      X-Phone-Os:
        exact: Android
' | emctl apply -f -

# Check all kinds of traffic.
# Primary Traffic
curl http://127.0.0.1:30188/ -d '{"order_id": "abc1234", "food": "bread"}'

# Beijing Traffic
curl http://127.0.0.1:30188/ -d '{"order_id": "abc1234", "food": "bread"}' -H 'X-Location: Beijing'

# Android Traffic
curl http://127.0.0.1:30188/ -d '{"order_id": "abc1234", "food": "bread"}' -H 'X-Phone-Os: Android'

# Beijing&Android traffic
curl http://127.0.0.1:30188/ -d '{"order_id": "abc1234", "food": "bread"}' -H 'X-Location: Beijing' -H 'X-Phone-Os: Android'

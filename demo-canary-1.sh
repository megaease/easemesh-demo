#!/usr/bin/env bash

# Traffic Shadow Problem (Demo)

echo '
apiVersion: mesh.megaease.com/v1alpha1
kind: ServiceCanary
metadata:
  name: beijing-android
spec:
  priority: 1
  selector:
    matchServices: [restaurant-mesh]
    matchInstanceLabels: {release: restaurant-mesh-beijing}
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
  priority: 2
  selector:
    matchServices: [restaurant-mesh, delivery-mesh]
    matchInstanceLabels: {release: refund-android}
  trafficRules:
    headers:
      X-Phone-Os:
        exact: Android
' | emctl apply -f -

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

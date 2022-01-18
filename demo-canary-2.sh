#!/usr/bin/env bash

# Fix Intersection Based on Priority (Demo)

# Check priorty of canary releases.
emctl get servicecanary

# Check all kinds of traffic.
# Primary Traffic
curl http://127.0.0.1:32539/ -d '{"order_id": "abc1234", "food": "bread"}'

# Beijing Traffic
curl http://127.0.0.1:32539/ -d '{"order_id": "abc1234", "food": "bread"}' -H 'X-Location: Beijing'

# Android Traffic
curl http://127.0.0.1:32539/ -d '{"order_id": "abc1234", "food": "bread"}' -H 'X-Phone-Os: Android'

# Beijing&Android traffic
curl http://127.0.0.1:32539/ -d '{"order_id": "abc1234", "food": "bread"}' -H 'X-Location: Beijing' -H 'X-Phone-Os: Android'

# ---------------------------

# Traffic Shade Problem (Demo)

# Change Priority

echo '
apiVersion: mesh.megaease.com/v1alpha1
kind: ServiceCanary
metadata:
  name: beijing-android
spec:
  priority: 2
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
curl http://127.0.0.1:32539/ -d '{"order_id": "abc1234", "food": "bread"}'

# Beijing Traffic
curl http://127.0.0.1:32539/ -d '{"order_id": "abc1234", "food": "bread"}' -H 'X-Location: Beijing'

# Android Traffic
curl http://127.0.0.1:32539/ -d '{"order_id": "abc1234", "food": "bread"}' -H 'X-Phone-Os: Android'

# Beijing&Android traffic
curl http://127.0.0.1:32539/ -d '{"order_id": "abc1234", "food": "bread"}' -H 'X-Location: Beijing' -H 'X-Phone-Os: Android'

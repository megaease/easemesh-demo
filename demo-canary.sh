#!/usr/bin/env bash

# 1. Team-delivery

# 1.1 Check primary traffic.
curl http://127.0.0.1:32539/ -d '{"order_id": "abc1234", "food": "bread"}'

# 1.2 Deploy delivery Beijing canary.
kubectl apply -f deploy/mesh/k8s_delivery_beijing.yaml

# 1.3 Check Beijing traffic.
curl http://127.0.0.1:32539/ -d '{"order_id": "abc1234", "food": "bread"}' -H 'X-Location: Beijing'

# 1.4 Apply service canary rules(This should be done before 1.3).
emctl apply -f deploy/mesh/easemesh_delivery_beijing.yaml

# 1.5 Check Beijing traffic.
curl http://127.0.0.1:32539/ -d '{"order_id": "abc1234", "food": "bread"}' -H 'X-Location: Beijing'

#----------------------------------------------------------------------------------

# 2. Team-restaurant

# 2.1 Deploy restaurant Beijing canary.
kubectl apply -f deploy/mesh/k8s_restaurant_beijing.yaml

# 2.2 Check Beijing traffic.
curl http://127.0.0.1:32539/ -d '{"order_id": "abc1234", "food": "bread"}' -H 'X-Location: Beijing'

# 2.3 Apply service canary rules(this should be done before 2.1).
emctl apply -f deploy/mesh/easemesh_restaurant_beijing.yaml

# 2.4 Check Beijing traffic.
curl http://127.0.0.1:32539/ -d '{"order_id": "abc1234", "food": "bread"}' -H 'X-Location: Beijing'

# 2.5 Check Beijing traffic with specifying as restaurant canary.
curl http://127.0.0.1:32539/ -d '{"order_id": "abc1234", "food": "bread"}' -H 'X-Location: Beijing' -H 'X-Mesh-Service-Canary: restaurant-mesh-beijing'

#----------------------------------------------------------------------------------

# Team-delivery co-operates with Team-restaurant

# 3.1 Check Android traffic.
curl http://127.0.0.1:32539/ -d '{"order_id": "abc1234", "food": "bread"}' -H 'X-Phone-Os: Android'

# 3.2 Apply service canary rules.
emctl apply -f deploy/mesh/easemesh_android.yaml

# 3.3 Deploy delivery Android canary.
kubectl apply -f deploy/mesh/k8s_delivery_android.yaml

# 3.4 Deploy restaurant Android canary.
kubectl apply -f deploy/mesh/k8s_restaurant_android.yaml

# 3.5 Check Android traffic.
curl http://127.0.0.1:32539/ -d '{"order_id": "abc1234", "food": "bread"}' -H 'X-Phone-Os: Android'

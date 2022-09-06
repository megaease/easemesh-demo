# EaseMesh Demo

The repository contains demo code and scripts for EaseMesh.

## Requirements

We need to [install EaseMesh and CoreDNS](https://github.com/megaease/easemesh/blob/main/docs/install.md).

## MicroServices

We wrote 3 services, and their call-chain is like `order(Go) -> Restuarnt(Java) -> Delivery(Go)`. It demonstrates the non-Java language applications could bidirectionally communicate with Java Spring applications.

## Demo Guide

The script is `./demo-canary.sh` which contains the whole procedure of demo for service canary.

And please notice the access port `30188` could differ from your environment. You can replace it by the node port from `kubectl get service -n mesh-service order-mesh-public`, whose precise output woule be ` kubectl get service -n mesh-service order-mesh-public -o yaml --output jsonpath='{.spec.ports[0].nodePort}'`.

After finishing testings, run `./clean-canary` would clean all of them.

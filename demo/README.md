# System Demonstration Scripts

This directory contains automated bash scripts designed to evaluate the core distributed systems concepts outlined in the project proposal: AP architecture, eventual consistency, and vector clock-based state reconciliation.

## Prerequisites

Before executing any tests, ensure your local Kubernetes cluster is actively running with the Calico network plugin and the application manifests are deployed. Run these commands from the root directory of the project:

```bash
# 1. Boot the cluster with Calico enabled
minikube start --cni=calico

# 2. Compile and load the application image
docker build -t warehouse-node:latest .
minikube image load warehouse-node:latest

# 3. Deploy the peer-to-peer network
kubectl apply -f k8s/
```
## Execution Commands

**1. Normal Execution**
Demonstrates standard AP availability. Injects data into a single node and verifies that the gossip protocol successfully propagates the state to all peer nodes without a central database. **Run:** `./demo/01-normal-execution.sh`

**2. Node Failure & Recovery**
Simulates a hardware/process crash. A node is taken offline while the remaining cluster continues to accept writes. Upon restoration, the recovered node synchronizes its ledger with the active cluster. **Run:** `./demo/02-node-failure.sh`

**3. Network Partition & Reconciliation (`03-network-partition.sh`)**
Demonstrates Vector Clock causality. A Kubernetes NetworkPolicy isolates a node. Concurrent updates are injected into the split network. Once healed, the additive merge protocol mathematically resolves the state without data loss. **Run:** `./demo/03-network-partition.sh`
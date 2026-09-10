# Distributed Multi-Warehouse Inventory Manager

A decentralized, fault-tolerant inventory management system built in Go. This project explores distributed systems concepts by removing the central database in favor of local node state, prioritizing Availability and Partition Tolerance (AP in the CAP theorem). 

This project is intended for the Distributed Systems course at the University of Bologna.

## Architecture Overview

Traditional inventory systems rely on a centralized database, creating a single point of failure and latency bottlenecks. This system replaces the central database with a network of independent warehouse nodes. 

* **Eventual Consistency:** Each warehouse maintains its own local state via an embedded SQLite database, allowing for immediate local reads and writes even during WAN network partitions.
* **Conflict Resolution:** The system implements **Vector Clocks** to accurately capture causality and detect concurrent transactions across disconnected nodes.
* **State Reconciliation:** Upon network reconnection, a custom additive reconciliation protocol mathematically merges conflicting state updates without data loss.
* **Decentralized Communication:** Nodes communicate asynchronously via a background gRPC gossip protocol.
* **Orchestration:** The distributed environment is containerized via Docker and orchestrated locally using Minikube (Kubernetes).

## Technologies Used
* **Language:** Go
* **Networking:** gRPC / Protocol Buffers
* **Storage:** SQLite
* **Deployment:** Docker, Kubernetes (Minikube)
* **Network Policies:** Project Calico (CNI)

## Prerequisites for Local Development
* [Go](https://go.dev/) (1.21+)
* [Docker](https://www.docker.com/)
* [Minikube](https://minikube.sigs.k8s.io/docs/)
* [kubectl](https://kubernetes.io/docs/tasks/tools/)

---

## Quick Start & Deployment

**1. Bootstrap the Cluster**  
Start Minikube with Calico enabled to support the network partition simulations:
```bash
minikube start --cni=calico
```
**2. Build and Load the Image**  
Compile the node container and load it into Minikube's local registry:
```bash
docker build -t warehouse-node:latest .
minikube image load warehouse-node:latest
```
**3. Deploy the Infrastructure**  
Apply the Kubernetes manifests to spin up the distributed nodes:
```bash
kubectl apply -f k8s/
```
Wait until all pods are running via `kubectl get pods`.

## Usage
The system is operated via a local CLI client that communicates with the nodes over gRPC.

**1. Connect to a Local Node**
```bash
kubectl port-forward svc/node-a-svc 50051:50051
```
**2. Execute CLI Commands**  
Open a new terminal window to interact with the forwarded node:
```bash
# Add 50 units of a monitor
go run ./cmd/cli -target localhost:50051 -item monitor -add 50

# Deduct 20 units
go run ./cmd/cli -target localhost:50051 -item monitor -add -20

# Query a specific item
go run ./cmd/cli -target localhost:50051 -item monitor -get

# Audit the full local ledger and Vector Clock state
go run ./cmd/cli -target localhost:50051 -all
```

## Automated Validation Testing

The system's core capabilities are validated via automated bash scripts located in the `/demo` directory. More information in the folder's `README.md`.

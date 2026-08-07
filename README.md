# Distributed Multi-Warehouse Inventory Manager

A decentralized, fault-tolerant inventory management system built in Go. This project explores distributed systems concepts by removing the central database in favor of local node state, prioritizing Availability and Partition Tolerance (AP in the CAP theorem). 

This project was developed for the Distributed Systems course at the University of Bologna.

## Architecture Overview

Traditional inventory systems rely on a centralized database, creating a single point of failure and latency bottlenecks. This system replaces the central database with a network of independent warehouse nodes. 

* **Eventual Consistency:** Each warehouse maintains its own local state via SQLite, allowing for immediate local reads and writes even during network partitions.
* **Conflict Resolution:** The system implements **Vector Clocks** to accurately capture causality and detect concurrent transactions across disconnected nodes.
* **State Reconciliation:** Upon network reconnection, a custom reconciliation protocol safely merges conflicting state updates without data loss.
* **Decentralized Communication:** Nodes communicate directly with each other via gRPC.
* **Orchestration:** The distributed environment is containerized via Docker and orchestrated locally using Minikube (Kubernetes).

## Technologies Used
* **Language:** Go
* **Networking:** gRPC / Protocol Buffers
* **Storage:** SQLite
* **Deployment:** Docker, Kubernetes (Minikube)

## Key System Scenarios

The system is designed to handle the following distributed scenarios:
1. **Normal Execution:** Nodes discover each other and synchronize inventory updates in real-time.
2. **Node Failures:** If a node crashes, Kubernetes spins up a replacement that recovers its state from the local disk.
3. **Network Partitions:** If a node is isolated from the network, it continues to process local orders (Availability).
4. **State Reconciliation:** When a partition heals, nodes use vector clocks to identify concurrent updates and merge their ledgers cleanly.

## Prerequisites for Local Development
* [Go](https://go.dev/)
* [Docker](https://www.docker.com/)
* [Minikube](https://minikube.sigs.k8s.io/docs/)
* [kubectl](https://kubernetes.io/docs/tasks/tools/)

#!/bin/bash
echo "0. Resetting Cluster State"
kubectl rollout restart deploy node-a node-b node-c
echo "Waiting for fresh pods to initialize..."
kubectl rollout status deploy/node-c --timeout=30s
sleep 5

echo "1. Starting Port Forwards"
kubectl port-forward svc/node-a-svc 50051:50051 > /dev/null 2>&1 &
PID_A=$!
kubectl port-forward svc/node-b-svc 50052:50051 > /dev/null 2>&1 &
PID_B=$!
kubectl port-forward svc/node-c-svc 50053:50051 > /dev/null 2>&1 &
PID_C=$!
sleep 3

echo "2. Severing Network (Node A)"
kubectl apply -f k8s/network-partition.yaml
sleep 2

echo "3. Injecting Split-Brain Data"
echo "Node A (Isolated): Selling 20 monitors (-20)"
go run ./cmd/cli -target localhost:50051 -item monitor -add -20
echo "Node B (Active): Receiving 50 monitors (+50)"
go run ./cmd/cli -target localhost:50052 -item monitor -add 50

echo "4. Healing Network"
kubectl delete -f k8s/network-partition.yaml
echo "Waiting 15 seconds for gossip synchronization..."
sleep 15

echo "5. Verifying Global State"
go run ./cmd/cli -target localhost:50051 -all
go run ./cmd/cli -target localhost:50052 -all
go run ./cmd/cli -target localhost:50053 -all

echo "6. Cleaning Up"
kill $PID_A $PID_B $PID_C
echo "Network partition test complete."
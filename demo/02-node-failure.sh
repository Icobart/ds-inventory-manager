#!/bin/bash
echo "1. Starting Port Forward (Node A)"
kubectl port-forward svc/node-a-svc 50051:50051 > /dev/null 2>&1 &
PID_A=$!
sleep 3

echo "2. Simulating Node C Failure"
kubectl scale deploy node-c --replicas=0
sleep 5

echo "3. Injecting Data During Failure"
go run ./cmd/cli -target localhost:50051 -item table -add 15

echo "4. Recovering Node C"
kubectl scale deploy node-c --replicas=1
echo "Waiting for pod to initialize..."
kubectl wait --for=condition=ready pod -l app=warehouse-node-c --timeout=30s
sleep 5

echo "5. Verifying Recovered State"
kubectl port-forward svc/node-c-svc 50053:50051 > /dev/null 2>&1 &
PID_C=$!
sleep 3
go run ./cmd/cli -target localhost:50053 -item table -get

echo "6. Cleaning Up"
kill $PID_A $PID_C
echo "Node failure test complete."
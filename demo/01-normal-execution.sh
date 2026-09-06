#!/bin/bash
echo "1. Starting Port Forwards"
kubectl port-forward svc/node-a-svc 50051:50051 > /dev/null 2>&1 &
PID_A=$!
kubectl port-forward svc/node-b-svc 50052:50051 > /dev/null 2>&1 &
PID_B=$!
sleep 3

echo "2. Injecting Data into Node A"
go run ./cmd/cli -target localhost:50051 -item chair -add 40

echo "3. Waiting for Gossip Protocol (5s) ==="
sleep 5

echo "4. Verifying State on Node B"
go run ./cmd/cli -target localhost:50052 -item chair -get

echo "5. Cleaning Up"
kill $PID_A $PID_B
echo "Normal execution test complete."
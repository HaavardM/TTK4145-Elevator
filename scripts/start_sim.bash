#!/bin/bash

if [ ! -f SimElevatorServer ]; then
    wget https://github.com/TTK4145/Simulator-v2/releases/download/v1.5/SimElevatorServer
    chmod +x SimElevatorServer
fi
start_elevator () {
    sleep 2
    go run ../cmd/main.go
}

start_elevator&
./SimElevatorServer --port=8080


#!/bin/bash

if [ ! -f SimElevatorServer ]; then
    wget https://github.com/TTK4145/Simulator-v2/releases/download/v1.5/SimElevatorServer
    chmod +x SimElevatorServer
fi

echo "Enter ID"
read id
echo "Enter port"
read port

start_elevator () {
    sleep 2
    go run --race ../main.go --id=$id --elevator-port=$port
}

start_elevator&
./SimElevatorServer --port=$port
docker kill $(docker ps -a -q)


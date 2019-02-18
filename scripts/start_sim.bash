#!/bin/bash

if [ ! -f SimElevatorServer ]; then
    wget https://github.com/TTK4145/Simulator-v2/releases/download/v1.5/SimElevatorServer
    chmod +x SimElevatorServer
fi
start_elevator () {
    sleep 2
    docker run --network host thefuturezebras/project
}

start_elevator&
./SimElevatorServer --port=15657


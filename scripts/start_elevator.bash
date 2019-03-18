#!/bin/bash

echo "Enter ID"
read id

start_elevator () {
    sleep 2
    docker run --network host thefuturezebras/project:$(git rev-parse HEAD) ./main --id=$id
}

start_elevator&
ElevatorServer --port=8080


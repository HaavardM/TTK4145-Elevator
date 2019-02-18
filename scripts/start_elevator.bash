#!/bin/bash

start_elevator () {
    sleep 2
    docker run --network host thefuturezebras/project:$(git rev-parse HEAD)
}

start_elevator&
ElevatorServer --port=8080


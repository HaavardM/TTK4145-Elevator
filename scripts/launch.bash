#!/bin/bash

go build -o main ../main.go

tmux new-session -s sim -n sim1 -d './SimElevatorServer --port=1234'
tmux split-window -t sim:0 -d './SimElevatorServer --port=2345'
tmux split-window -t sim:0 -d './restartIfError.bash ./main --elevator-port=1234 --id=1 --folder=orders1.json'
tmux split-window -t sim:0 -d './restartIfError.bash ./main --elevator-port=2345 --id=2 --folder=orders2.json'

tmux select-layout -t sim:0 tiled

tmux attach -t sim

#!/bin/bash

go build -o main ../main.go

tmux new-session -s sim -n sim1 -d './SimElevatorServer --port=1234'
tmux new-window -t sim:1 -n sim2 './SimElevatorServer --port=2345'
tmux new-window -t sim:2 -n elev1 './restartIfError.bash ./main --elevator-port=1234 --id=1 --folder=/home/haavard/orders1'
tmux new-window -t sim:3 -n elev2 './restartIfError.bash ./main --elevator-port=2345 --id=2 --folder=/home/haavard/orders2'

tmux join-pane 

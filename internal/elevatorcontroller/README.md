Module: Elevator Controller
===========================

- The elevator controller module implements a simple fsm for the elevator.
- It will only execute one order at a time, sent from the scheduler.
- It sends a message back to the scheduler once the order is completed.

## External packages
|Package Name|Description|Reason|
|------------|-----------|------|
|[context](https://golang.org/x/net/context)|Goroutine context management (included in standard library from Golang 1.7)|To stop the goroutine if the context is no longer valid|

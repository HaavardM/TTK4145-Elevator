Module: Elevator driver
======================
The elevator driver adapts commands for setting lights, motors etc. in Golang so that it is understood by the hardware.

## External packages
|Package Name|Description|Reason|
|------------|-----------|------|
|[context](https://golang.org/x/net/context)|Goroutine context management (included in standard library from Golang 1.7)|To stop the goroutine if the context is no longer valid|
|[driver-go](github.com/TTK4145/driver-go/elevio)|Handout from TTK4145|To control the hardware of the elevator|
Module: Elevator driver
======================
The elevator driver adapts commands for setting lights, motors etc. in Golang so that it is understood by the hardware.

## External packages
|Package Name|Description|Reason|
|------------|-----------|------|
|[context](https://golang.org/x/net/context)| |To stop the goroutine if the context expires|
|[driver-go](github.com/TTK4145/driver-go/elevio)| |To set the hardware of the elevator|

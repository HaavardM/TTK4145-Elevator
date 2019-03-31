Module: Scheduler
=====================
- The scheduler module schedules new orders to the cheapest elevator
- Orders are assigned by the elevator that received the order
- Orders are assigned a deadline, and if it is expired, the order is reassigned a new elevator. This also applies to "offline" elevators
- Worst case normal execution time
- Can function in single elevator mode. Should then finish all orders it has been assigned in addition to own cab calls


## External packages
|Package Name|Description|Reason|
|------------|-----------|------|
|[context](https://golang.org/x/net/context)| |To stop the goroutine if the context expires|
|[driver-go](github.com/TTK4145/driver-go/elevio)| |To set the hardware of the elevator|
|[go-spew](https://github.com/davecgh/go-spew/spew)| | |
|[xid](https://github.com/rs/xid) | Generates globally unique IDs | Used to assign messages unique message ids |


	
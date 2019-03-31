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
|[context](https://golang.org/x/net/context)|Goroutine context management (included in standard library from Golang 1.7)|To stop the goroutine if the context is no longer valid|
|[driver-go](github.com/TTK4145/driver-go/elevio)|Handout from TTK4145|To control the hardware of the elevator|
|[go-spew](https://github.com/davecgh/go-spew/spew)|Pretty printer for complex data structures|To dump order list to console|
|[xid](https://github.com/rs/xid) | Generates globally unique IDs | Used to assign unique order ids |


	
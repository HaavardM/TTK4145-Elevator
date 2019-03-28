Module: Scheduler
=====================
- The scheduler module schedules new orders to the cheapest elevator
- Orders are assigned by the elevator that received the order
- Orders are assigned a deadline, and if it is expired, the order is reassigned a new elevator. This also applies to "offline" elevators
- Worst case normal execution time
- Can function in single elevator mode. Should then finish all orders it has been assigned in addition to own cab calls


Used Packages
---------------
In this several golang packages have been used
- "time": 							to make deadlines for orders
- "math": 							to assign infinite values for comparison of cheapest elevator cost
- "context":							to stop the goroutine if the context expires
- "sync":								to make sure the main routine waits for this goroutine to finish
- "encoding/json":					to convert structs of orders into json format before saving to file, and then convert back before reading them again.
- "io/ioutil":						to be able to read the json-files
- "os":								to open the json-files
- "github.com/davecgh/go-spew/spew":	to format the data types	
- "errors":							to handle errors if they occur
	
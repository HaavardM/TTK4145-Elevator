Module: Elevator driver
======================
The elevator driver adapts commands for setting lights, motors etc. in Golang so that it is understood by the hardware.

Used Packages
--------------
- "context": 	to to stop the goroutine if the context expires
- "errors":	to handle errors if they occur
Module: Elevator Controller
===========================

- The elevator controller module implements a simple FSM for the elevator.
- It will only execute one order at a time, sent from the scheduler as seen in the figure below.
- It sends a message back to the scheduler once the order is completed.


Used Packages
---------------
Two golang packages have been used
- "context" :	to stop the goroutine if the context expires
- "time":		to implement timers
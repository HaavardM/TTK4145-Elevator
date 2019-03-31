package common

//ElevatorStatus contains information about the current status of the elevator
type ElevatorStatus struct {
	OrderDir Direction
	Moving   bool
	Floor    int
	Error    error
}

package common

//ElevatorStatus contains information about the current status of the elevator
type ElevatorStatus struct {
	Dir    Direction
	Moving bool
	Floor  int
}

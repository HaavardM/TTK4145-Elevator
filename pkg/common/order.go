package common

//Struct containing information about the floor and direction of the order (Up,Down,ButtonCall)
type Order struct {
	Dir   Direction `json:"direction"`
	Floor int       `json:"floor"`
}

package common

import "fmt"

//Direction specifies a direction for an order or the elevator
type Direction int

const (
	//UpDir is the direction upwards
	UpDir Direction = iota + 1
	//DownDir is the direction downwards
	DownDir
	//NoDir is when no direction is specified.
	//Typically used by cab orders
	NoDir
)

//Returns a string representation of the direction
func (d Direction) String() string {
	switch d {
	case UpDir:
		return "Up"
	case DownDir:
		return "Down"
	case NoDir:
		return "NoDir"
	}
	return fmt.Sprintf("%d", d)
}

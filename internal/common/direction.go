package common

import "fmt"

type Direction int

const (
	UpDir Direction = iota + 1
	DownDir
	NoDir
)

func (d Direction) String() string {
	switch d {
	case UpDir:
		return "Up"
	case DownDir:
		return "Down"
	case NoDir:
		return "NoDir"
	}
	return fmt.Sprint(d)
}

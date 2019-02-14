package elevatordriver

import "fmt"

type CommandID int
type Event int

const (
	CloseDoor           CommandID = 1
	OpenDoor            CommandID = 2
	MoveUp              CommandID = 3
	MoveDown            CommandID = 4
	Stop                CommandID = 5
	SetOrderLightUp     CommandID = 6
	ClearOrderLightUp   CommandID = 7
	SetOrderLightDown   CommandID = 8
	ClearOrderLightDown CommandID = 9
	SetInternalLight    CommandID = 10
	ClearInternalLight  CommandID = 11
)

type Command struct {
	command CommandID
	floor   int
}

type Config struct {
	commands       <-chan Command
	events         chan<- Event
	arrivedAtFloor chan<- int
}

func Run(config Config) {
	fmt.Println("Hello")
}

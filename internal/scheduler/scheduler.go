package scheduler

import (
	"context"

	"github.com/TTK4145-students-2019/project-thefuturezebras/internal/elevatordriver"
	"github.com/TTK4145/driver-go/elevio"
)

type Order struct {
	floor int
	id    int
}

type Config struct {
	ButtonPressed  <-chan elevio.ButtonEvent
	OrderCompleted <-chan Order
	CurrentOrder   chan<- Order
	Lights         chan<- elevatordriver.LightState
}

func Run(ctx context.Context)

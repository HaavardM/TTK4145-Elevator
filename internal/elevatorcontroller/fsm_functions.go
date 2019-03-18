package elevatorcontroller
import (
	"context"
	"log"
	"time"

	"github.com/TTK4145-students-2019/project-thefuturezebras/internal/elevatordriver"
)

ordersUp = [NumberOfFloors]int
ordersDown = [NumberOfFloors]int
ordersNoDir = [NumberOfFloors]int


func shouldStop(floor int) bool {
	switch f.state {
	case stateMovingUp:
		if ordersUp[floor] == 1 || ordersNoDir[floor] == 1 || !ordersAbove(floor){
			return true
		}
		return false
	case stateMovingDown:
		if ordersDown[floor] == 1 || ordersNoDir[floor] == 1 || !ordersBelow(floor){
			return true
		}
		return false
	/*Vi skal vel ikke håndtere disse tre statene i should stop. DoorOpen og closed er jo allerede på 
	etasje og bør ikke være nødvendig å spørre om å skulle stoppe
	case stateDoorOpen:
	case stateDoorClosed:
	case stateEmergency:*/
	}
	return false
}

func ordersAbove(floor int) bool {
	for i := floor + 1; i < NumberOfFloors; i++ {
		if ordersUp[i] == 1 || ordersNoDir[i] == 1 {
			return true
		}
	}
	return false
}

func ordersBelow(floor int) bool {
	for i := 0; i < floor; i++ {
		if ordersDown[i] == 1 || ordersNoDir[i] == 1 {
			return true
		}
	}
	return false
}

func addOrders(orders []Order) {
	switch Order.Dir {
	case DOWN:
		ordersDown[Order.Floor] == 1
	case UP:
		ordersUp[Order.Floor] == 1
	case NoDirection:
		ordersNoDir[Order.Floor] == 1
	}	
}

func deleteOrders(floor int) {
	//sletter vi alltid alle ordre i alle retninger samtidig??
	ordersDown[floor] == 0
	ordersUp[floor] == 0
	ordersNoDir[floor] == 0
	
}
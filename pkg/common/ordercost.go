package common

//OrderCosts contains ID of an elevator, number of orders the elevator has and its cost to the different floors.
type OrderCosts struct {
	ID         int       `json:"id"`
	OrderCount int       `json:"order_count"`
	HallUp     []float64 `json:"cost_up"`
	HallDown   []float64 `json:"cost_down"`
	Cab        []float64 `json:"cost_cab"`
}

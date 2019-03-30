package common

//OrderCosts contains the cost of an order to the different floors in one direction
type OrderCosts struct {
	ID         int       `json:"id"`
	OrderCount int       `json:"order_count"`
	HallUp     []float64 `json:"cost_up"`
	HallDown   []float64 `json:"cost_down"`
	Cab        []float64 `json:"cost_cab"`
}

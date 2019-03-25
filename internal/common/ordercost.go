package common

//OrderCosts contains the cost of an order to the different floors in one direction
type OrderCosts struct {
	ID        int       `json:"id"`
	CostsUp   []float64 `json:"cost_up"`
	CostsDown []float64 `json:"cost_down"`
	CostsCab  []float64 `json:"cost_cab"`
}

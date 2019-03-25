package common

type OrderCosts struct {
	ID        int       `json"id"`
	CostsUp   []float64 `json:"cost_up"`
	CostsDown []float64 `json:"cost_down"`
}

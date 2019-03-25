package common

type Order struct {
	Dir   Direction `json:"direction"`
	Floor int       `json:"floor"`
}

package models

type FlightPathRequest struct {
	Flights [][]string `json:"flights"`
}

type Flight struct {
	Source      string `json:"source"`
	Destination string `json:"destination"`
}

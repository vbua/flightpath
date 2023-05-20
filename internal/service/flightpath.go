package service

import (
	"github.com/vbua/flightpath/internal/models"
)

type FlightPath struct{}

func NewFlightPath() *FlightPath {
	return &FlightPath{}
}

func (f *FlightPath) FindStartAndEndOfPath(flights [][]string) models.Flight {
	sources := make(map[string]struct{})
	destinations := make(map[string]struct{})

	for _, flight := range flights {
		sources[flight[0]] = struct{}{}
		destinations[flight[1]] = struct{}{}
	}

	var startEndFlight models.Flight

	for _, flight := range flights {
		if _, found := destinations[flight[0]]; !found {
			startEndFlight.Source = flight[0]
		} else {
			delete(destinations, flight[0])
		}

		if _, found := sources[flight[1]]; !found {
			startEndFlight.Destination = flight[1]
		} else {
			delete(sources, flight[1])
		}
	}

	return startEndFlight
}

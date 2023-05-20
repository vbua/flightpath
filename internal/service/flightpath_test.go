package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vbua/flightpath/internal/models"
)

func TestFindStartAndEndOfPath(t *testing.T) {
	flightpathService := NewFlightPath()

	type test struct {
		input [][]string
		want  models.Flight
	}

	tests := []test{
		{
			input: [][]string{
				{
					"IND",
					"EWR",
				},
				{
					"SFO",
					"ATL",
				},
				{
					"GSO",
					"IND",
				},
				{
					"ATL",
					"GSO",
				},
			}, want: models.Flight{Source: "SFO", Destination: "EWR"},
		},
		{
			input: [][]string{
				{
					"ATL",
					"EWR",
				},
				{
					"SFO",
					"ATL",
				},
			}, want: models.Flight{Source: "SFO", Destination: "EWR"},
		},
		{
			input: [][]string{
				{
					"SFO",
					"EWR",
				},
			}, want: models.Flight{Source: "SFO", Destination: "EWR"},
		},
		{
			input: [][]string{}, want: models.Flight{Source: "", Destination: ""},
		},
		{
			input: [][]string{
				{
					"EWR",
					"IND",
				},
				{
					"IND",
					"EWR",
				},
				{
					"SFO",
					"ATL",
				},
				{
					"GSO",
					"IND",
				},
				{
					"ATL",
					"GSO",
				},
				{
					"IND",
					"EWR",
				},
			}, want: models.Flight{Source: "IND", Destination: "EWR"},
		},
	}

	for _, tc := range tests {
		got := flightpathService.FindStartAndEndOfPath(tc.input)
		assert.Equal(t, tc.want, got)
	}
}

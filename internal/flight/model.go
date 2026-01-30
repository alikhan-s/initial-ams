package flight

import (
	"fmt"
	"time"
)

// Status represents the lifecycle of a flight
type Status string

const (
	StatusScheduled Status = "SCHEDULED"
	StatusDelayed   Status = "DELAYED"
	StatusBoarding  Status = "BOARDING"
	StatusDeparted  Status = "DEPARTED"
	StatusCancelled Status = "CANCELLED"
)

// Flight represents a row in the "flights" table
type Flight struct {
	ID            int64     `json:"id"`
	FlightNo      string    `json:"flight_no"`
	Origin        string    `json:"origin"`
	Destination   string    `json:"destination"`
	GateID        *int64    `json:"gate_id,omitempty"` // Pointer because it can be null
	DepartureTime time.Time `json:"departure_time"`
	ArrivalTime   time.Time `json:"arrival_time"`
	Status        Status    `json:"status"`
	Version       int       `json:"-"` // Internal version for optimistic locking
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// Validate Validation Logic attached to the model
func (f *Flight) Validate() error {
	if f.Origin == f.Destination {
		return fmt.Errorf("origin and destination cannot be the same")
	}
	if f.ArrivalTime.Before(f.DepartureTime) {
		return fmt.Errorf("arrival time cannot be before departure time")
	}
	return nil
}

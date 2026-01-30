package booking

import (
	"time"
)

type Ticket struct {
	ID          int64     `json:"id"`
	FlightID    int64     `json:"flight_id"`
	PassengerID int64     `json:"passenger_id"`
	SeatNo      string    `json:"seat_no,omitempty"` // Can be empty initially
	Price       float64   `json:"price"`
	Status      string    `json:"status"` // "ACTIVE", "CANCELLED"
	CreatedAt   time.Time `json:"created_at"`
}

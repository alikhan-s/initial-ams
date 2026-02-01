package airportops

type Gate struct {
	ID         int64  `json:"id"`
	TerminalID int64  `json:"terminal_id"`
	Code       string `json:"code"`   // e.g. "A1"
	Status     string `json:"status"` // OPEN, CLOSED, MAINTENANCE
}

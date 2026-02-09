package passenger

type Passenger struct {
	ID         int64  `json:"id"`
	UserID     int64  `json:"user_id"`
	PassportNo string `json:"passport_no"`
	Phone      string `json:"phone"`
}

package models

// UserBalance represents the balance of a user
type UserBalance struct {
	UserID     int64   `json:"user_id"`
	UserName   string  `json:"user_name"`
	EmployeeID string  `json:"employee_id"`
	Balance    float64 `json:"balance"`
}

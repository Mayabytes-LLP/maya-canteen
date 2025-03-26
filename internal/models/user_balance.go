package models

import "time"

// UserBalance represents the balance of a user
type UserBalance struct {
	UserID           int64      `json:"user_id"`
	UserName         string     `json:"user_name"`
	EmployeeID       string     `json:"employee_id"`
	Department       string     `json:"user_department"`
	UserActive       bool       `json:"user_active"`
	LastNotification *time.Time `json:"last_notification"`
	Phone            string     `json:"user_phone"`
	Balance          float64    `json:"balance"`
}

package models

import (
	"time"
)

// User represents a user in the system
type User struct {
	ID         int64     `json:"id"`
	Name       string    `json:"name"`
	EmployeeId string    `json:"employee_id"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// GetID returns the user ID
func (u *User) GetID() int64 {
	return u.ID
}

// SetID sets the user ID
func (u *User) SetID(id int64) {
	u.ID = id
}

// SetCreatedAt sets the created timestamp
func (u *User) SetCreatedAt(timestamp interface{}) {
	if t, ok := timestamp.(time.Time); ok {
		u.CreatedAt = t
	}
}

// SetUpdatedAt sets the updated timestamp
func (u *User) SetUpdatedAt(timestamp interface{}) {
	if t, ok := timestamp.(time.Time); ok {
		u.UpdatedAt = t
	}
}

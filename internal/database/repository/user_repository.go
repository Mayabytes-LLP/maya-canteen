package repository

import (
	"database/sql"
	"maya-canteen/internal/models"
	"time"
)

// UserRepository handles all database operations related to users
type UserRepository struct {
	db *sql.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// InitTable initializes the users table
func (r *UserRepository) InitTable() error {
	query := `
		CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			employee_id TEXT NOT NULL,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL
		)
	`
	_, err := r.db.Exec(query)
	return err
}

// Create inserts a new user into the database
func (r *UserRepository) Create(user *models.User) error {
	query := `
		INSERT INTO users (name, employee_id, created_at, updated_at)
		VALUES (?, ?, ?, ?)
	`
	now := time.Now()
	result, err := r.db.Exec(
		query,
		user.Name,
		user.EmployeeId,
		now,
		now,
	)
	if err != nil {
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	user.ID = id
	user.CreatedAt = now
	user.UpdatedAt = now
	return nil
}

// GetAll retrieves all users from the database
func (r *UserRepository) GetAll() ([]models.User, error) {
	query := `SELECT * FROM users ORDER BY name ASC`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		err := rows.Scan(
			&user.ID,
			&user.Name,
			&user.EmployeeId,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, nil
}

// Get retrieves a single user by ID
func (r *UserRepository) Get(id int64) (*models.User, error) {
	query := `SELECT * FROM users WHERE id = ?`
	var user models.User
	err := r.db.QueryRow(query, id).Scan(
		&user.ID,
		&user.Name,
		&user.EmployeeId,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Update updates an existing user
func (r *UserRepository) Update(user *models.User) error {
	query := `
		UPDATE users
		SET name = ?, employee_id = ?, updated_at = ?
		WHERE id = ?
	`
	now := time.Now()
	_, err := r.db.Exec(
		query,
		user.Name,
		user.EmployeeId,
		now,
		user.ID,
	)
	if err != nil {
		return err
	}
	user.UpdatedAt = now
	return nil
}

// Delete removes a user by ID
func (r *UserRepository) Delete(id int64) error {
	query := `DELETE FROM users WHERE id = ?`
	_, err := r.db.Exec(query, id)
	return err
}

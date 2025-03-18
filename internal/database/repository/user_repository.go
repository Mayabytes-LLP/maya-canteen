package repository

import (
	"database/sql"
	"fmt"
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
			employee_id TEXT NOT NULL UNIQUE,
    	phone TEXT,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL
		)
	`
	_, err := r.db.Exec(query)
	if err != nil {
		return err
	}

	err1 := r.Create(&models.User{
		Name:       "Abdul Rafay",
		EmployeeId: "10058",
		Phone:      "+923452324442",
	})
	err2 := r.Create(&models.User{
		Name:       "Qasim Imtiaz",
		EmployeeId: "10037",
		Phone:      "+923452565003",
	})

	err3 := r.Create(&models.User{
		Name:       "Syed Kazim Raza",
		EmployeeId: "10024",
		Phone:      "+923422949447",
	})

	if err1 != nil || err2 != nil || err3 != nil {
		fmt.Printf("Error in adding admin possibly already exists to the database %v\n %v\n %v\n", err1, err2, err3)
	}

	return nil
}

// Create inserts a new user into the database
func (r *UserRepository) Create(user *models.User) error {
	query := `
		INSERT INTO users (name, employee_id, phone, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)
	`
	now := time.Now()
	result, err := r.db.Exec(
		query,
		user.Name,
		user.EmployeeId,
		user.Phone,
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
			&user.Phone,
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
	fmt.Println("Get user by ID", id)
	// make sure id int has 5 digits
	idStr := fmt.Sprintf("%05d", id)
	query := `SELECT * FROM users WHERE employee_id = ?`
	var user models.User
	err := r.db.QueryRow(query, idStr).Scan(
		&user.ID,
		&user.Name,
		&user.EmployeeId,
		&user.Phone,
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
		SET name = ?, employee_id = ?, phone = ?, updated_at = ?
		WHERE id = ?
	`
	now := time.Now()
	_, err := r.db.Exec(
		query,
		user.Name,
		user.EmployeeId,
		user.Phone,
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

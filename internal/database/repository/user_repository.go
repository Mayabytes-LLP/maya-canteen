package repository

import (
	"database/sql"
	"fmt"
	"maya-canteen/internal/models"
	"time"

	log "github.com/sirupsen/logrus"
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
	// First check if the active column exists, if not, add it
	r.addActiveColumnIfNeeded()
	r.addLastNotificationColumnIfNeeded()

	query := `
		CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			department TEXT NOT NULL,
			employee_id TEXT NOT NULL UNIQUE,
			phone TEXT,
			active BOOLEAN NOT NULL DEFAULT 1,
      last_notification DATETIME,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL
		)
	`
	_, err := r.db.Exec(query)
	if err != nil {
		log.Errorf("Error creating users table: %v", err)
		return err
	}
	log.Info("Created Users Table")

	err1 := r.Create(&models.User{
		Name:       "Abdul Rafay",
		EmployeeId: "10081",
		Department: "Development Dept",
		Phone:      "+923452324442",
		Active:     true,
	})
	err2 := r.Create(&models.User{
		Name:       "Qasim Imtiaz",
		EmployeeId: "1023",
		Department: "Development Dept",
		Phone:      "+923452565003",
		Active:     true,
	})

	err3 := r.Create(&models.User{
		Name:       "Syed Kazim Raza",
		EmployeeId: "10024",
		Department: "Admin Dept",
		Phone:      "+923422949447",
		Active:     true,
	})

	if err1 != nil || err2 != nil || err3 != nil {
		log.Errorf("Error in adding admin possibly already exists to the database:\n %v\n %v\n %v\n", err1, err2, err3)
	}

	return nil
}

// addActiveColumnIfNeeded checks if the active column exists and adds it if needed
func (r *UserRepository) addActiveColumnIfNeeded() {
	// Check if the column exists
	var colExists bool
	err := r.db.QueryRow(`
		SELECT COUNT(*) > 0
		FROM pragma_table_info('users')
		WHERE name = 'active'
	`).Scan(&colExists)

	if err != nil || colExists {
		if err != nil {
			log.Errorf("Error checking if active column exists: %v", err)
		}
		log.Info("Active column already exists in users table")
		return // Either error occurred or column already exists
	}

	// Add the column if it doesn't exist
	_, err = r.db.Exec(`ALTER TABLE users ADD COLUMN active BOOLEAN NOT NULL DEFAULT 1`)
	if err != nil {
		log.Errorf("Error adding active column to users table: %v", err)
	} else {
		log.Info("Added active column to users table")
	}
}

func (r *UserRepository) addLastNotificationColumnIfNeeded() {
	var colExists bool
	err := r.db.QueryRow(`
    SELECT COUNT(*) > 0
    FROM pragma_table_info('users')
    WHERE name = 'last_notification'
  `).Scan(&colExists)

	if err != nil || colExists {
		if err != nil {
			log.Errorf("Error checking if last_notification column exists: %v", err)
		}
		log.Info("last_notification column already exists in users table")
		return
	}

	_, err = r.db.Exec(`ALTER TABLE users ADD COLUMN last_notification DATETIME DEFAULT NULL`)
	if err != nil {
		log.Errorf("Error adding last_notification column to users table: %v", err)
	} else {
		log.Info("Added last_notification column to users table")
	}
}

// Create inserts a new user into the database
func (r *UserRepository) Create(user *models.User) error {
	query := `
		INSERT INTO users (name, employee_id, department, phone, active, last_notification, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`
	now := time.Now()
	// If Active field is not explicitly set, default to true (active)
	if !user.Active {
		user.Active = true
	}

	// LastNotification can be NULL, so handle it accordingly
	var lastNotification any
	if user.LastNotification != nil {
		lastNotification = *user.LastNotification
	} else {
		lastNotification = nil
	}

	result, err := r.db.Exec(
		query,
		user.Name,
		user.EmployeeId,
		user.Department,
		user.Phone,
		user.Active,
		lastNotification,
		now,
		now,
	)
	if err != nil {
		log.Errorf("Error inserting user: %v", err)
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		log.Errorf("Error getting last insert ID: %v", err)
		return err
	}
	user.ID = id
	user.CreatedAt = now
	user.UpdatedAt = now
	return nil
}

// GetAll retrieves all users from the database
func (r *UserRepository) GetAll() ([]models.User, error) {
	query := `SELECT id, name, employee_id, department, phone, active, last_notification, created_at, updated_at FROM users ORDER BY name ASC`
	rows, err := r.db.Query(query)
	if err != nil {
		log.Errorf("Error getting all users: %v", err)
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		var lastNotificationNull sql.NullTime

		err := rows.Scan(
			&user.ID,
			&user.Name,
			&user.EmployeeId,
			&user.Department,
			&user.Phone,
			&user.Active,
			&lastNotificationNull,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			log.Errorf("Error scanning user row: %v", err)
			return nil, err
		}

		if lastNotificationNull.Valid {
			user.LastNotification = &lastNotificationNull.Time
		}

		users = append(users, user)
	}
	return users, nil
}

// Get retrieves a single user by ID
func (r *UserRepository) Get(id int64) (*models.User, error) {
	fmt.Println("Get user by ID", id)
	query := `SELECT id, name, employee_id, department, phone, active, last_notification, created_at, updated_at FROM users WHERE employee_id = ?`

	var user models.User
	var lastNotificationNull sql.NullTime

	err := r.db.QueryRow(query, id).Scan(
		&user.ID,
		&user.Name,
		&user.EmployeeId,
		&user.Department,
		&user.Phone,
		&user.Active,
		&lastNotificationNull,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		log.Errorf("No user found with ID %d", id)
		return nil, nil
	}
	if err != nil {
		log.Errorf("Error in getting user by ID: %v", err)
		return nil, err
	}

	if lastNotificationNull.Valid {
		user.LastNotification = &lastNotificationNull.Time
	}

	return &user, nil
}

// GetByEmployeeID retrieves a single user by employee ID
func (r *UserRepository) GetByEmployeeID(employeeID string) (*models.User, error) {
	query := `SELECT id, name, employee_id, department, phone, active, last_notification, created_at, updated_at FROM users WHERE employee_id = ?`

	var user models.User
	var lastNotificationNull sql.NullTime

	err := r.db.QueryRow(query, employeeID).Scan(
		&user.ID,
		&user.Name,
		&user.EmployeeId,
		&user.Department,
		&user.Phone,
		&user.Active,
		&lastNotificationNull,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		log.Errorf("No user found with employee ID %s", employeeID)
		return nil, nil
	}
	if err != nil {
		log.Errorf("Error in getting user by employee ID: %v", err)
		return nil, err
	}

	if lastNotificationNull.Valid {
		user.LastNotification = &lastNotificationNull.Time
	}

	return &user, nil
}

// Update updates an existing user
func (r *UserRepository) Update(user *models.User) error {
	fmt.Println("Edit user by ID", user)
	query := `
		UPDATE users
		SET name = ?, employee_id = ?, department = ?, phone = ?, active = ?, updated_at = ?
		WHERE id = ?
	`
	now := time.Now()
	_, err := r.db.Exec(
		query,
		user.Name,
		user.EmployeeId,
		user.Department,
		user.Phone,
		user.Active,
		now,
		user.ID,
	)
	if err != nil {
		log.Errorf("Error updating user: %v", err)
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

func (r *UserRepository) UpdateLastNotificationTime(employeeID string) error {
	query := `UPDATE users SET last_notification = ? WHERE employee_id = ?`
	_, err := r.db.Exec(query, time.Now(), employeeID)
	if err != nil {
		log.Errorf("Error updating last notification time for user: %v", err)
	}
	return err
}

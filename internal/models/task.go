package models

import (
	"database/sql"
	"time"
)

type Task struct {
	ID          int64     `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type TaskModel struct {
	DB *sql.DB
}

func NewTaskModel(db *sql.DB) *TaskModel {
	return &TaskModel{DB: db}
}

func (m *TaskModel) InitTable() error {
	query := `
		CREATE TABLE IF NOT EXISTS tasks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			description TEXT,
			status TEXT NOT NULL DEFAULT 'pending',
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL
		)
	`
	_, err := m.DB.Exec(query)
	return err
}

func (m *TaskModel) Create(task *Task) error {
	query := `
		INSERT INTO tasks (title, description, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)
	`
	now := time.Now()
	result, err := m.DB.Exec(query, task.Title, task.Description, task.Status, now, now)
	if err != nil {
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	task.ID = id
	task.CreatedAt = now
	task.UpdatedAt = now
	return nil
}

func (m *TaskModel) GetAll() ([]Task, error) {
	query := `
		SELECT * FROM tasks ORDER BY created_at DESC
	`
	rows, err := m.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var task Task
		err := rows.Scan(
			&task.ID,
			&task.Title,
			&task.Description,
			&task.Status,
			&task.CreatedAt,
			&task.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}
	return tasks, nil
}

func (m *TaskModel) Get(id int64) (*Task, error) {
	query := `
		SELECT * FROM tasks
		WHERE id = ?
	`
	var task Task
	err := m.DB.QueryRow(query, id).Scan(
		&task.ID,
		&task.Title,
		&task.Description,
		&task.Status,
		&task.CreatedAt,
		&task.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &task, nil
}

func (m *TaskModel) Update(task *Task) error {
	query := `
		UPDATE tasks
		SET title = ?, description = ?, status = ?, updated_at = ?
		WHERE id = ?
	`

	now := time.Now()
	_, err := m.DB.Exec(query, task.Title, task.Description, task.Status, now, task.ID)
	if err != nil {
		return err
	}
	task.UpdatedAt = now
	return nil
}

func (m *TaskModel) Delete(id int64) error {
	query := `DELETE FROM tasks WHERE id = ?`
	_, err := m.DB.Exec(query, id)
	return err
}


package repository

import (
	"database/sql"
	"reflect"
)

// Entity represents a database entity with ID, CreatedAt, and UpdatedAt fields
type Entity interface {
	GetID() int64
	SetID(id int64)
	SetCreatedAt(timestamp interface{})
	SetUpdatedAt(timestamp interface{})
}

// GenericRepository provides a generic implementation for common repository operations
type GenericRepository struct {
	db        *sql.DB
	tableName string
}

// NewGenericRepository creates a new generic repository
func NewGenericRepository(db *sql.DB, tableName string) *GenericRepository {
	return &GenericRepository{
		db:        db,
		tableName: tableName,
	}
}

// GetDB returns the database connection
func (r *GenericRepository) GetDB() *sql.DB {
	return r.db
}

// GetTableName returns the table name
func (r *GenericRepository) GetTableName() string {
	return r.tableName
}

// BuildInsertQuery builds an INSERT query for the given columns
func (r *GenericRepository) BuildInsertQuery(columns []string) string {
	query := "INSERT INTO " + r.tableName + " ("

	// Add columns
	for i, col := range columns {
		if i > 0 {
			query += ", "
		}
		query += col
	}

	query += ") VALUES ("

	// Add placeholders
	for i := range columns {
		if i > 0 {
			query += ", "
		}
		query += "?"
	}

	query += ")"

	return query
}

// BuildUpdateQuery builds an UPDATE query for the given columns
func (r *GenericRepository) BuildUpdateQuery(columns []string) string {
	query := "UPDATE " + r.tableName + " SET "

	// Add columns with placeholders
	for i, col := range columns {
		if i > 0 {
			query += ", "
		}
		query += col + " = ?"
	}

	query += " WHERE id = ?"

	return query
}

// BuildSelectQuery builds a SELECT query with optional WHERE clause
func (r *GenericRepository) BuildSelectQuery(where string) string {
	query := "SELECT * FROM " + r.tableName

	if where != "" {
		query += " WHERE " + where
	}

	return query
}

// BuildDeleteQuery builds a DELETE query with optional WHERE clause
func (r *GenericRepository) BuildDeleteQuery(where string) string {
	query := "DELETE FROM " + r.tableName

	if where != "" {
		query += " WHERE " + where
	}

	return query
}

// ScanRows scans rows into a slice of entities using reflection
func ScanRows(rows *sql.Rows, entityType reflect.Type) (interface{}, error) {
	sliceType := reflect.SliceOf(entityType)
	slice := reflect.MakeSlice(sliceType, 0, 0)

	for rows.Next() {
		entity := reflect.New(entityType).Interface()
		err := rows.Scan(entity)
		if err != nil {
			return nil, err
		}
		slice = reflect.Append(slice, reflect.ValueOf(entity).Elem())
	}

	return slice.Interface(), nil
}

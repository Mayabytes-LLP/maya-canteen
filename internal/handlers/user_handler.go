package handlers

import (
	"encoding/csv"
	"fmt"
	"io"
	"maya-canteen/internal/database"
	"maya-canteen/internal/errors"
	"maya-canteen/internal/handlers/common"
	"maya-canteen/internal/models"
	"net/http"

	"github.com/gorilla/mux"
)

// UserHandler handles user-related HTTP requests
type UserHandler struct {
	common.BaseHandler
}

// NewUserHandler creates a new user handler
func NewUserHandler(db database.Service) *UserHandler {
	return &UserHandler{
		BaseHandler: common.NewBaseHandler(db),
	}
}

// CreateUser handles POST /api/users
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var user models.User
	if err := h.DecodeJSON(r, &user); err != nil {
		h.HandleError(w, err)
		return
	}

	if err := h.DB.CreateUser(&user); err != nil {
		h.HandleError(w, errors.Internal(err))
		return
	}

	common.RespondWithSuccess(w, http.StatusCreated, user)
}

// GetAllUsers handles GET /api/users
func (h *UserHandler) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.DB.GetAllUsers()
	if err != nil {
		h.HandleError(w, errors.Internal(err))
		return
	}

	common.RespondWithSuccess(w, http.StatusOK, users)
}

// GetUser handles GET /api/users/{id}
func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := h.ParseID(vars, "id")
	if err != nil {
		h.HandleError(w, err)
		return
	}

	user, err := h.DB.GetUser(id)
	if err != nil {
		h.HandleError(w, errors.Internal(err))
		return
	}

	if user == nil {
		h.HandleError(w, errors.NotFound("User", id))
		return
	}

	common.RespondWithSuccess(w, http.StatusOK, user)
}

// UpdateUser handles PUT /api/users/{id}
func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := h.ParseID(vars, "id")
	if err != nil {
		h.HandleError(w, err)
		return
	}

	var user models.User
	if err := h.DecodeJSON(r, &user); err != nil {
		h.HandleError(w, err)
		return
	}
	user.ID = id

	if err := h.DB.UpdateUser(&user); err != nil {
		h.HandleError(w, errors.Internal(err))
		return
	}

	common.RespondWithSuccess(w, http.StatusOK, user)
}

// DeleteUser handles DELETE /api/users/{id}
func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := h.ParseID(vars, "id")
	if err != nil {
		h.HandleError(w, err)
		return
	}

	if err := h.DB.DeleteUser(id); err != nil {
		h.HandleError(w, errors.Internal(err))
		return
	}

	common.RespondWithSuccess(w, http.StatusNoContent, nil)
}

// CSVUploadResponse represents the response for CSV upload
type CSVUploadResponse struct {
	Success int      `json:"success"`
	Failed  int      `json:"failed"`
	Errors  []string `json:"errors"`
}

// UploadUserCSV handles the CSV upload and creates users from it
func (h *UserHandler) UploadUserCSV(w http.ResponseWriter, r *http.Request) {
	// Parse the multipart form
	err := r.ParseMultipartForm(10 << 20) // 10 MB max
	if err != nil {
		common.RespondWithError(w, http.StatusBadRequest, "Failed to parse form")
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		common.RespondWithError(w, http.StatusBadRequest, "Failed to get file from form")
		return
	}
	defer file.Close()

	// Read CSV
	reader := csv.NewReader(file)

	// Read header
	header, err := reader.Read()
	if err != nil {
		common.RespondWithError(w, http.StatusBadRequest, "Failed to read CSV header")
		return
	}

	// Validate header
	expectedHeaders := []string{"name", "employee_id", "department", "phone"}
	if !validateHeaders(header, expectedHeaders) {
		common.RespondWithError(w, http.StatusBadRequest, "Invalid CSV headers. Expected: name, employee_id, department, phone")
		return
	}

	response := CSVUploadResponse{
		Success: 0,
		Failed:  0,
		Errors:  make([]string, 0),
	}

	// Read and process each row
	lineNum := 1 // Start from 1 as header is line 0
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			response.Failed++
			response.Errors = append(response.Errors, fmt.Sprintf("Line %d: Failed to read row", lineNum))
			continue
		}

		// Create user from CSV record
		user := models.User{
			Name:       record[0],
			EmployeeId: record[1],
			Department: record[2],
			Phone:      record[3],
		}

		// Attempt to create the user
		err = h.DB.CreateUser(&user)
		if err != nil {
			response.Failed++
			response.Errors = append(response.Errors, fmt.Sprintf("Line %d: %s", lineNum, err.Error()))
		} else {
			response.Success++
		}

		lineNum++
	}

	common.RespondWithJSON(w, http.StatusOK, response)
}

// validateHeaders checks if the CSV headers match the expected headers
func validateHeaders(actual, expected []string) bool {
	if len(actual) != len(expected) {
		return false
	}
	for i, header := range actual {
		if header != expected[i] {
			return false
		}
	}
	return true
}

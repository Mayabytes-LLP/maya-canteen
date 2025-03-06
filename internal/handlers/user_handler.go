package handlers

import (
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

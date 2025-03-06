package common

import (
	"encoding/json"
	"maya-canteen/internal/database"
	"maya-canteen/internal/errors"
	"net/http"
	"strconv"
)

// BaseHandler provides common functionality for all handlers
type BaseHandler struct {
	DB database.Service
}

// NewBaseHandler creates a new base handler
func NewBaseHandler(db database.Service) BaseHandler {
	return BaseHandler{DB: db}
}

// ParseID parses an ID from the URL parameters
func (h *BaseHandler) ParseID(vars map[string]string, paramName string) (int64, error) {
	idStr, ok := vars[paramName]
	if !ok {
		return 0, errors.InvalidInput("Missing ID parameter: " + paramName)
	}
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return 0, errors.InvalidInput("Invalid ID: " + idStr)
	}
	return id, nil
}

// DecodeJSON decodes JSON from the request body into the given target
func (h *BaseHandler) DecodeJSON(r *http.Request, target interface{}) error {
	if err := json.NewDecoder(r.Body).Decode(target); err != nil {
		return errors.InvalidInput("Invalid JSON: " + err.Error())
	}
	return nil
}

// HandleError handles an error and sends an appropriate response
func (h *BaseHandler) HandleError(w http.ResponseWriter, err error) {
	var appErr *errors.AppError
	if errors.As(err, &appErr) {
		switch {
		case errors.Is(appErr, errors.ErrNotFound):
			RespondWithError(w, http.StatusNotFound, appErr.Error())
		case errors.Is(appErr, errors.ErrInvalidInput):
			RespondWithError(w, http.StatusBadRequest, appErr.Error())
		case errors.Is(appErr, errors.ErrUnauthorized):
			RespondWithError(w, http.StatusUnauthorized, appErr.Error())
		case errors.Is(appErr, errors.ErrForbidden):
			RespondWithError(w, http.StatusForbidden, appErr.Error())
		default:
			RespondWithError(w, http.StatusInternalServerError, appErr.Error())
		}
		return
	}

	// Handle non-AppError errors
	RespondWithInternalError(w, err)
}

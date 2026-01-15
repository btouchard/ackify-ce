package handlers

import (
	"errors"
	"net/http"

	"github.com/btouchard/ackify-ce/backend/pkg/logger"
	"github.com/btouchard/ackify-ce/backend/pkg/models"
)

// HandleError handles different types of errors and returns appropriate HTTP responses
func HandleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, models.ErrUnauthorized):
		logger.Logger.Warn("Unauthorized access attempt", "error", err.Error())
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
	case errors.Is(err, models.ErrSignatureNotFound):
		logger.Logger.Debug("Signature not found", "error", err.Error())
		http.Error(w, "Signature not found", http.StatusNotFound)
	case errors.Is(err, models.ErrSignatureAlreadyExists):
		logger.Logger.Debug("Duplicate signature attempt", "error", err.Error())
		http.Error(w, "Signature already exists", http.StatusConflict)
	case errors.Is(err, models.ErrInvalidUser):
		logger.Logger.Warn("Invalid user data", "error", err.Error())
		http.Error(w, "Invalid user", http.StatusBadRequest)
	case errors.Is(err, models.ErrInvalidDocument):
		logger.Logger.Warn("Invalid document ID", "error", err.Error())
		http.Error(w, "Invalid document ID", http.StatusBadRequest)
	case errors.Is(err, models.ErrDomainNotAllowed):
		logger.Logger.Warn("Domain not allowed", "error", err.Error())
		http.Error(w, "Domain not allowed", http.StatusForbidden)
	case errors.Is(err, models.ErrDatabaseConnection):
		logger.Logger.Error("Database connection error", "error", err.Error())
		http.Error(w, "Database error", http.StatusInternalServerError)
	default:
		logger.Logger.Error("Unhandled error", "error", err.Error())
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

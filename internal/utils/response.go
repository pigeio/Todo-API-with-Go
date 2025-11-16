package utils

import (
	"encoding/json"
	"net/http"

	"github.com/pigeio/todo-api/internal/models"
)

func RespondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func RespondError(w http.ResponseWriter, status int, message string, details ...string) {
	response := models.ErrorResponse{
		Error:   message,
		Details: details,
	}
	RespondJSON(w, status, response)
}

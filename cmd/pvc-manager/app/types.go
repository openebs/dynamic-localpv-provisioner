package app

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// VolumeRequest represents a request for volume operations
type VolumeRequest struct {
	Name               string            `json:"name"`
	Path               string            `json:"path"`
	NodeAffinityLabels map[string]string `json:"nodeAffinityLabels"`
	FsMode             string            `json:"fsMode,omitempty"`
	Commands           []string          `json:"commands"`
	SoftLimitGrace     string            `json:"softLimitGrace,omitempty"`
	HardLimitGrace     string            `json:"hardLimitGrace,omitempty"`
	PVCStorage         int64             `json:"pvcStorage,omitempty"`
}

// VolumeResponse represents a response for volume operations
type VolumeResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

// HealthResponse represents a health check response
type HealthResponse struct {
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
	Version   string `json:"version"`
}

// APIError represents an API error response
type APIError struct {
	Error   string `json:"error"`
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// WriteJSONResponse writes a JSON response to the http.ResponseWriter
func WriteJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

// WriteError writes an error response to the http.ResponseWriter
func WriteError(w http.ResponseWriter, statusCode int, message string) {
	WriteJSONResponse(w, statusCode, APIError{
		Error:   http.StatusText(statusCode),
		Code:    statusCode,
		Message: message,
	})
}

// ParseJSONRequest parses a JSON request body into the provided interface
func ParseJSONRequest(r *http.Request, v interface{}) error {
	if r.Body == nil {
		return fmt.Errorf("request body is empty")
	}
	defer r.Body.Close()

	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		return fmt.Errorf("failed to parse JSON request: %v", err)
	}

	return nil
}

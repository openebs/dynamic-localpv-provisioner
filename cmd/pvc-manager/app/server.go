package app

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"k8s.io/klog/v2"
)

const (
	// API version
	APIVersion = "v1"

	// Server version
	Version = "1.0.0"
)

// Server represents the PVC Manager HTTP server
type Server struct {
	listenAddr    string
	server        *http.Server
	volumeManager *VolumeManager
	router        *mux.Router
}

// NewServer creates a new PVC Manager server
func NewServer(listenAddr string) *Server {
	s := &Server{
		listenAddr:    listenAddr,
		volumeManager: NewVolumeManager(),
	}

	s.setupRoutes()

	return s
}

// setupRoutes configures the HTTP routes
func (s *Server) setupRoutes() {
	s.router = mux.NewRouter()

	// API v1 routes
	api := s.router.PathPrefix(fmt.Sprintf("/api/%s", APIVersion)).Subrouter()

	// Health check
	api.HandleFunc("/health", s.healthHandler).Methods("GET")

	// Volume operations
	api.HandleFunc("/volumes/create", s.createVolumeHandler).Methods("POST")
	api.HandleFunc("/volumes/delete", s.deleteVolumeHandler).Methods("POST")
	api.HandleFunc("/volumes/quota", s.applyQuotaHandler).Methods("POST")

	// Add middleware
	s.router.Use(s.loggingMiddleware)
	s.router.Use(s.corsMiddleware)
}

// Start starts the HTTP server
func (s *Server) Start() error {
	s.server = &http.Server{
		Addr:         s.listenAddr,
		Handler:      s.router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	klog.Infof("PVC Manager server listening on %s", s.listenAddr)
	return s.server.ListenAndServe()
}

// Stop gracefully shuts down the HTTP server
func (s *Server) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return s.server.Shutdown(ctx)
}

// healthHandler handles health check requests
func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	response := HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now().Format(time.RFC3339),
		Version:   Version,
	}

	WriteJSONResponse(w, http.StatusOK, response)
}

// createVolumeHandler handles volume creation requests
func (s *Server) createVolumeHandler(w http.ResponseWriter, r *http.Request) {
	var req VolumeRequest
	if err := ParseJSONRequest(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := s.validateVolumeRequest(&req); err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := s.volumeManager.CreateVolume(r.Context(), &req); err != nil {
		klog.Errorf("Failed to create volume %s: %v", req.Name, err)
		WriteError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to create volume: %v", err))
		return
	}

	response := VolumeResponse{
		Success: true,
		Message: fmt.Sprintf("Volume %s created successfully", req.Name),
	}

	WriteJSONResponse(w, http.StatusOK, response)
}

// deleteVolumeHandler handles volume deletion requests
func (s *Server) deleteVolumeHandler(w http.ResponseWriter, r *http.Request) {
	var req VolumeRequest
	if err := ParseJSONRequest(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	if req.Name == "" || req.Path == "" {
		WriteError(w, http.StatusBadRequest, "name and path are required for volume deletion")
		return
	}

	if err := s.volumeManager.DeleteVolume(r.Context(), &req); err != nil {
		klog.Errorf("Failed to delete volume %s: %v", req.Name, err)
		WriteError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to delete volume: %v", err))
		return
	}

	response := VolumeResponse{
		Success: true,
		Message: fmt.Sprintf("Volume %s deleted successfully", req.Name),
	}

	WriteJSONResponse(w, http.StatusOK, response)
}

// applyQuotaHandler handles quota application requests
func (s *Server) applyQuotaHandler(w http.ResponseWriter, r *http.Request) {
	var req VolumeRequest
	if err := ParseJSONRequest(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := s.validateQuotaRequest(&req); err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := s.volumeManager.ApplyQuota(r.Context(), &req); err != nil {
		klog.Errorf("Failed to apply quota for volume %s: %v", req.Name, err)
		WriteError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to apply quota: %v", err))
		return
	}

	response := VolumeResponse{
		Success: true,
		Message: fmt.Sprintf("Quota applied successfully for volume %s", req.Name),
	}

	WriteJSONResponse(w, http.StatusOK, response)
}

// validateVolumeRequest validates a volume request
func (s *Server) validateVolumeRequest(req *VolumeRequest) error {
	if req.Name == "" {
		return fmt.Errorf("name is required")
	}
	if req.Path == "" {
		return fmt.Errorf("path is required")
	}
	if len(req.Commands) == 0 {
		return fmt.Errorf("commands are required")
	}
	return nil
}

// validateQuotaRequest validates a quota request
func (s *Server) validateQuotaRequest(req *VolumeRequest) error {
	if req.Name == "" {
		return fmt.Errorf("name is required")
	}
	if req.Path == "" {
		return fmt.Errorf("path is required")
	}
	if req.PVCStorage == 0 {
		return fmt.Errorf("pvcStorage is required for quota operations")
	}
	return nil
}

// loggingMiddleware logs HTTP requests
func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		klog.Infof("%s %s %s", r.Method, r.URL.Path, time.Since(start))
	})
}

// corsMiddleware adds CORS headers
func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

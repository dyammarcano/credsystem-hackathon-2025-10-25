package service

import (
	"fmt"
	"net/http"
	"os"

	"github.com/spf13/cobra"
)

// Intent representa una intención del CSV
type Intent struct {
	Text        string
	ServiceID   int
	ServiceName string
}

// FindServiceRequest representa el request del endpoint
type FindServiceRequest struct {
	Intent string `json:"intent"`
}

// FindServiceResponse representa el response del endpoint
type FindServiceResponse struct {
	Success bool         `json:"success"`
	Data    *ServiceData `json:"data,omitempty"`
	Error   string       `json:"error,omitempty"`
}

// ServiceData contiene los datos del servicio
type ServiceData struct {
	ServiceID   int    `json:"service_id"`
	ServiceName string `json:"service_name"`
}

// HealthResponse representa el response del healthcheck
type HealthResponse struct {
	Status string `json:"status"`
}

func Service(cmd *cobra.Command, args []string) error {
	// Read environment variables
	openRouterKey := os.Getenv("OPENROUTER_API_KEY")
	if openRouterKey == "" {
		return fmt.Errorf("environment variable OPENROUTER_API_KEY is required")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "18020"
	}

	router := http.NewServeMux()

	server := http.Server{
		Addr:    fmt.Sprintf(":%s", port),
		Handler: router,
	}

	router.HandleFunc("GET /api/health", healthHandler)
	router.HandleFunc("POST /api/find-service", findServiceHandler(openRouterKey))

	return server.ListenAndServe()
}

package service

import (
	"fmt"
	"net/http"
	"os"

	"github.com/dyammarcano/crew-das-closures/internal/client/openrouter"
	"github.com/dyammarcano/crew-das-closures/internal/core"
	"github.com/spf13/cobra"
)

func Service(_ *cobra.Command, _ []string) error {
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
		Handler: forceStatusOK(router),
	}

	urlStr := "https://openrouter.ai/api/v1"
	token := fmt.Sprintf("%s", openRouterKey)

	opts := openrouter.WithAuth(token)

	aks, err := core.NewCore(urlStr, opts)
	if err != nil {
		return fmt.Errorf("failed to initialize core: %w", err)
	}

	if err := aks.SetKey(openRouterKey); err != nil {
		return fmt.Errorf("failed to set api key: %w", err)
	}

	router.HandleFunc("GET /api/health", healthHandler)
	router.HandleFunc("POST /api/find-service", findServiceHandler(aks))

	return server.ListenAndServe()
}

package service

import (
	"fmt"
	"net/http"
)

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprint(w, `{"status":"ok"}`)
}

func findServiceHandler(key string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		// Aquí iría la lógica para manejar la solicitud de encontrar un servicio
		// usando el API key proporcionado.
		_, _ = fmt.Fprint(w, `{"status":"sou a AI"}`)
	}
}

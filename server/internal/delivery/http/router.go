package httpdelivery

import "net/http"

func NewRouter(checker HealthChecker) *http.ServeMux {
	mux := http.NewServeMux()
	healthHandler := NewHealthHandler(checker)
	mux.HandleFunc("GET /health", healthHandler.Check)
	return mux
}

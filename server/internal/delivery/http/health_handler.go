package httpdelivery

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/putriindaht/leftnt/server/internal/usecase"
)

// HealthChecker describes the application behavior this handler needs.
type HealthChecker interface {
	Check() usecase.HealthStatus
}

type HealthHandler struct {
	checker HealthChecker
}

func NewHealthHandler(checker HealthChecker) *HealthHandler {
	return &HealthHandler{checker: checker}
}

func (h *HealthHandler) Check(w http.ResponseWriter, r *http.Request) {
	status := h.checker.Check()
	response := struct {
		Status string `json:"status"`
		Game   string `json:"game"`
	}{
		Status: status.Status,
		Game:   status.Game,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("write health response: %v", err)
	}
}

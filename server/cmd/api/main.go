package main

import (
	"log"
	"net/http"
	"time"

	httpdelivery "github.com/putriindaht/leftnt/server/internal/delivery/http"
	"github.com/putriindaht/leftnt/server/internal/storage/memory"
	"github.com/putriindaht/leftnt/server/internal/usecase"
)

func main() {
	healthService := usecase.NewHealthService()
	gameRepository := memory.NewGameRepository()
	gameService := usecase.NewGameService(gameRepository)
	router := httpdelivery.NewRouter(healthService, gameService)

	server := &http.Server{
		Addr:              ":8080",
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Println("LEFTN'T server is listening on http://localhost:8080")
	log.Fatal(server.ListenAndServe())
}

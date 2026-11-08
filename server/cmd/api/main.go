package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	httpdelivery "github.com/putriindaht/leftnt/server/internal/delivery/http"
	"github.com/putriindaht/leftnt/server/internal/protocol"
	"github.com/putriindaht/leftnt/server/internal/storage/memory"
	"github.com/putriindaht/leftnt/server/internal/usecase"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	rooms := usecase.NewRoomService(memory.NewRoomRepository())
	roomDone := make(chan struct{})
	go func() { defer close(roomDone); rooms.Run(ctx) }()
	origins := []string{"http://localhost:3000", "http://127.0.0.1:3000"}
	if configured := os.Getenv("ALLOWED_ORIGINS"); configured != "" {
		origins = nil
		for _, origin := range strings.Split(configured, ",") {
			if value := strings.TrimSpace(origin); value != "" {
				origins = append(origins, value)
			}
		}
	}
	var iceServers []protocol.ICEServer
	if raw := os.Getenv("WEBRTC_ICE_SERVERS"); raw != "" {
		if err := json.Unmarshal([]byte(raw), &iceServers); err != nil {
			log.Fatal("WEBRTC_ICE_SERVERS must be a JSON array of ICE server settings")
		}
		for _, server := range iceServers {
			if len(server.URLs) == 0 {
				log.Fatal("each WebRTC ICE server requires at least one URL")
			}
		}
	}
	router := httpdelivery.NewRouter(usecase.NewHealthService(), usecase.NewGameService(memory.NewGameRepository()), httpdelivery.NewRoomHandler(rooms, origins, iceServers))
	root := http.NewServeMux()
	root.Handle("/api/", http.StripPrefix("/api", router))
	root.Handle("/health", router)
	root.Handle("/rooms", router)
	root.Handle("/rooms/", router)
	root.Handle("/games", router)
	root.Handle("/games/", router)
	if staticDir := os.Getenv("STATIC_DIR"); staticDir != "" {
		root.Handle("/", http.FileServer(http.Dir(staticDir)))
	} else {
		root.Handle("/", router)
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	server := &http.Server{Addr: net.JoinHostPort("", port), Handler: root, ReadHeaderTimeout: 5 * time.Second, BaseContext: func(net.Listener) context.Context { return ctx }}
	shutdownDone := make(chan struct{})
	go func() {
		defer close(shutdownDone)
		<-ctx.Done()
		shutdownContext, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownContext); err != nil {
			log.Printf("shutdown: %v", err)
		}
	}()
	log.Printf("LEFTN'T server is listening on http://localhost:%s", port)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
	stop()
	<-roomDone
	<-shutdownDone
}

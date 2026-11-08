package httpdelivery

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/putriindaht/leftnt/server/internal/protocol"
	"github.com/putriindaht/leftnt/server/internal/usecase"
)

type RoomHandler struct {
	service    *usecase.RoomService
	origins    []string
	iceServers []protocol.ICEServer
}

func NewRoomHandler(service *usecase.RoomService, origins []string, ice ...[]protocol.ICEServer) *RoomHandler {
	handler := &RoomHandler{service: service, origins: origins}
	if len(ice) > 0 {
		handler.iceServers = ice[0]
	}
	return handler
}

func (h *RoomHandler) Create(w http.ResponseWriter, r *http.Request) {
	session, err := h.service.Create()
	if err != nil {
		writeRoomError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, session)
}

func (h *RoomHandler) Join(w http.ResponseWriter, r *http.Request) {
	session, err := h.service.Join(r.PathValue("code"))
	if err != nil {
		writeRoomError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, session)
}

func writeRoomError(w http.ResponseWriter, err error) {
	status := http.StatusBadRequest
	switch {
	case errors.Is(err, usecase.ErrRoomNotFound):
		status = http.StatusNotFound
	case errors.Is(err, usecase.ErrRoomFull):
		status = http.StatusConflict
	case errors.Is(err, usecase.ErrRoomCapacity):
		status = http.StatusServiceUnavailable
	}
	writeJSON(w, status, map[string]string{"error": err.Error()})
}

func (h *RoomHandler) Socket(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{OriginPatterns: h.origins})
	if err != nil {
		return
	}
	defer conn.CloseNow()
	conn.SetReadLimit(1024)
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()
	authContext, authCancel := context.WithTimeout(ctx, 5*time.Second)
	var auth protocol.ClientMessage
	err = wsjson.Read(authContext, conn, &auth)
	authCancel()
	if err != nil || auth.Type != "authenticate" {
		_ = conn.Close(websocket.StatusPolicyViolation, "authenticate before sending commands")
		return
	}
	views, disconnect, err := h.service.Subscribe(r.PathValue("code"), auth.Token)
	if err != nil {
		writeContext, stop := context.WithTimeout(ctx, time.Second)
		_ = wsjson.Write(writeContext, conn, protocol.ServerMessage{Type: "error", Error: err.Error()})
		stop()
		_ = conn.Close(websocket.StatusPolicyViolation, "room session rejected")
		return
	}
	defer disconnect()
	replies := make(chan protocol.ServerMessage, 8)
	writerDone := make(chan struct{})
	go func() {
		defer close(writerDone)
		defer cancel()
		heartbeat := time.NewTicker(20 * time.Second)
		defer heartbeat.Stop()
		for {
			var message protocol.ServerMessage
			select {
			case view, ok := <-views:
				if !ok {
					return
				}
				message = protocol.ServerMessage{Type: "state", State: &view}
			case message = <-replies:
			case <-heartbeat.C:
				pingContext, stop := context.WithTimeout(ctx, 5*time.Second)
				err := conn.Ping(pingContext)
				stop()
				if err != nil {
					return
				}
				continue
			case <-ctx.Done():
				return
			}
			writeContext, stop := context.WithTimeout(ctx, 5*time.Second)
			err := wsjson.Write(writeContext, conn, message)
			stop()
			if err != nil {
				return
			}
		}
	}()
	defer func() { cancel(); <-writerDone }()
	for {
		var message protocol.ClientMessage
		if err := wsjson.Read(ctx, conn, &message); err != nil {
			return
		}
		moved, err := h.service.Act(r.PathValue("code"), auth.Token, message, time.Now())
		reply := protocol.ServerMessage{Type: "result"}
		if err != nil {
			reply.Type, reply.Error = "error", err.Error()
		} else if message.Type == "move" {
			reply.Moved = &moved
		}
		select {
		case replies <- reply:
		case <-ctx.Done():
			return
		default:
			return
		}
	}
}

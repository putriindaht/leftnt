package httpdelivery

import (
	"context"
	"net/http"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/putriindaht/leftnt/server/internal/protocol"
)

func (h *RoomHandler) Voice(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{OriginPatterns: h.origins})
	if err != nil {
		return
	}
	defer conn.CloseNow()
	conn.SetReadLimit(1024)
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()
	authCtx, stop := context.WithTimeout(ctx, 5*time.Second)
	var auth protocol.VoiceMessage
	err = wsjson.Read(authCtx, conn, &auth)
	stop()
	if err != nil || auth.Type != "authenticate" {
		_ = conn.Close(websocket.StatusPolicyViolation, "authenticate before joining voice")
		return
	}
	stream, leave, err := h.service.SubscribeVoice(r.PathValue("code"), auth.Token)
	if err != nil {
		writeCtx, stop := context.WithTimeout(ctx, time.Second)
		_ = wsjson.Write(writeCtx, conn, protocol.VoiceMessage{Type: "error", Error: err.Error()})
		stop()
		_ = conn.Close(websocket.StatusPolicyViolation, "voice session rejected")
		return
	}
	defer leave()
	conn.SetReadLimit(32768)
	replies := make(chan protocol.VoiceMessage, 4)
	done := make(chan struct{})
	go func() {
		defer close(done)
		defer cancel()
		ticker := time.NewTicker(20 * time.Second)
		defer ticker.Stop()
		for {
			var message protocol.VoiceMessage
			select {
			case message = <-replies:
			case next, ok := <-stream:
				if !ok {
					return
				}
				message = next
			case <-ticker.C:
				pingCtx, stop := context.WithTimeout(ctx, 5*time.Second)
				err := conn.Ping(pingCtx)
				stop()
				if err != nil {
					return
				}
				continue
			case <-ctx.Done():
				return
			}
			if message.Type == "ready" {
				message.ICEServers = h.iceServers
			}
			writeCtx, stop := context.WithTimeout(ctx, 5*time.Second)
			err := wsjson.Write(writeCtx, conn, message)
			stop()
			if err != nil {
				return
			}
		}
	}()
	defer func() { cancel(); <-done }()
	for {
		var signal protocol.VoiceMessage
		if err := wsjson.Read(ctx, conn, &signal); err != nil {
			return
		}
		if err := h.service.RelayVoice(r.PathValue("code"), auth.Token, signal); err != nil {
			select {
			case replies <- protocol.VoiceMessage{Type: "error", Error: err.Error()}:
			case <-ctx.Done():
				return
			default:
				return
			}
		}
	}
}

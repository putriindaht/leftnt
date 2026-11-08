package client

import (
	"context"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/putriindaht/leftnt/server/internal/protocol"
)

type RoomEvent struct {
	Message    protocol.ServerMessage
	Connection string
}

// RoomClient performs network work in the background; rendering only polls events.
type RoomClient struct {
	Events   chan RoomEvent
	commands chan protocol.ClientMessage
	cancel   context.CancelFunc
}

func NewRoomClient(endpoint, token string) *RoomClient {
	ctx, cancel := context.WithCancel(context.Background())
	client := &RoomClient{Events: make(chan RoomEvent, 16), commands: make(chan protocol.ClientMessage, 4), cancel: cancel}
	go client.run(ctx, endpoint, token)
	return client
}

func (c *RoomClient) Close() { c.cancel() }

func (c *RoomClient) Send(message protocol.ClientMessage) bool {
	select {
	case c.commands <- message:
		return true
	default:
		return false
	}
}

func (c *RoomClient) emit(event RoomEvent) {
	select {
	case c.Events <- event:
	default:
		select {
		case <-c.Events:
		default:
		}
		select {
		case c.Events <- event:
		default:
		}
	}
}

func (c *RoomClient) run(ctx context.Context, endpoint, token string) {
	for ctx.Err() == nil {
		c.emit(RoomEvent{Connection: "Connecting to room..."})
		c.connect(ctx, endpoint, token)
		if ctx.Err() != nil {
			return
		}
		c.emit(RoomEvent{Connection: "Connection lost. Reconnecting..."})
		// Discard commands entered before a disconnect, rather than replaying them.
		for len(c.commands) > 0 {
			<-c.commands
		}
		timer := time.NewTimer(time.Second)
		select {
		case <-timer.C:
		case <-ctx.Done():
			timer.Stop()
			return
		}
	}
}

func (c *RoomClient) connect(ctx context.Context, endpoint, token string) {
	dialContext, stop := context.WithTimeout(ctx, 8*time.Second)
	conn, _, err := websocket.Dial(dialContext, endpoint, nil)
	stop()
	if err != nil {
		return
	}
	defer conn.CloseNow()
	connectionContext, cancel := context.WithCancel(ctx)
	defer cancel()
	writeContext, stop := context.WithTimeout(connectionContext, 5*time.Second)
	err = wsjson.Write(writeContext, conn, protocol.ClientMessage{Type: "authenticate", Token: token})
	stop()
	if err != nil {
		return
	}
	readerDone := make(chan struct{})
	go func() {
		defer close(readerDone)
		defer cancel()
		for {
			var message protocol.ServerMessage
			if err := wsjson.Read(connectionContext, conn, &message); err != nil {
				return
			}
			c.emit(RoomEvent{Message: message})
		}
	}()
	defer func() { cancel(); conn.CloseNow(); <-readerDone }()
	for {
		select {
		case message := <-c.commands:
			writeContext, stop := context.WithTimeout(connectionContext, 5*time.Second)
			err := wsjson.Write(writeContext, conn, message)
			stop()
			if err != nil {
				return
			}
		case <-connectionContext.Done():
			return
		}
	}
}

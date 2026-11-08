package usecase_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/putriindaht/leftnt/server/internal/domain"
	"github.com/putriindaht/leftnt/server/internal/protocol"
	"github.com/putriindaht/leftnt/server/internal/storage/memory"
	"github.com/putriindaht/leftnt/server/internal/usecase"
)

func voiceEvent(t *testing.T, stream <-chan protocol.VoiceMessage) protocol.VoiceMessage {
	t.Helper()
	select {
	case message := <-stream:
		return message
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for voice signaling")
		return protocol.VoiceMessage{}
	}
}

func TestVoiceAuthenticatedRelayRejoinAndGameIsolation(t *testing.T) {
	repository := memory.NewRoomRepository()
	service := usecase.NewRoomService(repository)
	a, err := service.Create()
	if err != nil {
		t.Fatal(err)
	}
	b, err := service.Join(a.Code)
	if err != nil {
		t.Fatal(err)
	}
	gameA, leaveGameA, err := service.Subscribe(a.Code, a.Token)
	if err != nil {
		t.Fatal(err)
	}
	defer leaveGameA()
	_, leaveGameB, err := service.Subscribe(a.Code, b.Token)
	if err != nil {
		t.Fatal(err)
	}
	defer leaveGameB()
	if (<-gameA).Phase != "playing" {
		t.Fatal("both game seats should be connected")
	}
	if _, _, err := service.SubscribeVoice(a.Code, "wrong"); !errors.Is(err, usecase.ErrInvalidSession) {
		t.Fatal("voice must authenticate the existing seat")
	}
	streamA, leaveA, err := service.SubscribeVoice(a.Code, a.Token)
	if err != nil {
		t.Fatal(err)
	}
	defer leaveA()
	waiting := voiceEvent(t, streamA)
	if waiting.Type != "ready" || waiting.PeerReady || waiting.Role != domain.Explorer {
		t.Fatal("first caller should wait for their partner")
	}
	if _, _, err := service.SubscribeVoice(a.Code, a.Token); !errors.Is(err, usecase.ErrVoiceConnected) {
		t.Fatal("a seat may join voice once")
	}
	streamB, leaveB, err := service.SubscribeVoice(a.Code, b.Token)
	if err != nil {
		t.Fatal(err)
	}
	defer leaveB()
	readyA, readyB := voiceEvent(t, streamA), voiceEvent(t, streamB)
	if !readyA.PeerReady || !readyB.PeerReady || readyA.CallID == "" || readyA.CallID != readyB.CallID || readyB.Role != domain.Navigator {
		t.Fatal("both peers need the same call generation and their own roles")
	}
	offer := protocol.VoiceMessage{Type: "description", CallID: readyA.CallID, Token: "must not be relayed", Role: domain.Navigator, Description: &protocol.SessionDescription{Type: "offer", SDP: "v=0\r\na=audio-test\r\n"}}
	if err := service.RelayVoice(a.Code, a.Token, offer); err != nil {
		t.Fatal(err)
	}
	delivered := voiceEvent(t, streamB)
	if delivered.Description == nil || delivered.Description.SDP != offer.Description.SDP || delivered.Token != "" || delivered.Role != "" {
		t.Fatal("relay should forward only setup data to the other peer")
	}
	if err := service.RelayVoice(a.Code, b.Token, offer); !errors.Is(err, usecase.ErrInvalidSignal) {
		t.Fatal("only Explorer may originate an offer")
	}
	answer := protocol.VoiceMessage{Type: "description", CallID: readyA.CallID, Description: &protocol.SessionDescription{Type: "answer", SDP: "v=0\r\n"}}
	if err := service.RelayVoice(a.Code, b.Token, answer); err != nil {
		t.Fatal(err)
	}
	if voiceEvent(t, streamA).Description.Type != "answer" {
		t.Fatal("answer should reach Explorer")
	}
	candidate := protocol.VoiceMessage{Type: "candidate", CallID: readyA.CallID, Candidate: &protocol.ICECandidate{Candidate: "candidate:test"}}
	if err := service.RelayVoice(a.Code, a.Token, candidate); err != nil {
		t.Fatal(err)
	}
	if voiceEvent(t, streamB).Candidate.Candidate != "candidate:test" {
		t.Fatal("ICE should reach the other seat")
	}
	tooLarge := offer
	tooLarge.Description = &protocol.SessionDescription{Type: "offer", SDP: strings.Repeat("x", 24577)}
	if err := service.RelayVoice(a.Code, a.Token, tooLarge); !errors.Is(err, usecase.ErrInvalidSignal) {
		t.Fatal("large descriptions must be rejected")
	}
	other, err := service.Create()
	if err != nil {
		t.Fatal(err)
	}
	if err := service.RelayVoice(other.Code, a.Token, offer); !errors.Is(err, usecase.ErrInvalidSession) {
		t.Fatal("signals must never cross room boundaries")
	}
	leaveB()
	leaveB()
	if voiceEvent(t, streamA).Type != "peer-left" {
		t.Fatal("leaving must notify the remaining caller")
	}
	err = repository.Update(a.Code, func(record *usecase.RoomRecord) error {
		if !record.Room.Ready() {
			t.Fatal("voice leave must not pause game sockets")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	newB, leaveNewB, err := service.SubscribeVoice(a.Code, b.Token)
	if err != nil {
		t.Fatal(err)
	}
	defer leaveNewB()
	nextA, nextB := voiceEvent(t, streamA), voiceEvent(t, newB)
	if nextA.CallID == readyA.CallID || nextA.CallID != nextB.CallID {
		t.Fatal("rejoin must use a fresh call generation")
	}
	if err := service.RelayVoice(a.Code, a.Token, candidate); err != nil {
		t.Fatal("late candidates should be safely ignored")
	}
	select {
	case <-newB:
		t.Fatal("stale ICE must not reach the new call")
	default:
	}
}

func TestVoiceQueueIsBoundedAndLeaveCannotBlock(t *testing.T) {
	service := usecase.NewRoomService(memory.NewRoomRepository())
	a, err := service.Create()
	if err != nil {
		t.Fatal(err)
	}
	b, err := service.Join(a.Code)
	if err != nil {
		t.Fatal(err)
	}
	streamA, leaveA, err := service.SubscribeVoice(a.Code, a.Token)
	if err != nil {
		t.Fatal(err)
	}
	voiceEvent(t, streamA)
	streamB, leaveB, err := service.SubscribeVoice(a.Code, b.Token)
	if err != nil {
		t.Fatal(err)
	}
	defer leaveB()
	ready := voiceEvent(t, streamA)
	voiceEvent(t, streamB)
	signal := protocol.VoiceMessage{Type: "candidate", CallID: ready.CallID, Candidate: &protocol.ICECandidate{Candidate: "candidate:test"}}
	for i := 0; i < 64; i++ {
		if err := service.RelayVoice(a.Code, a.Token, signal); err != nil {
			t.Fatal(err)
		}
	}
	if err := service.RelayVoice(a.Code, a.Token, signal); !errors.Is(err, usecase.ErrVoiceBusy) {
		t.Fatal("a slow peer must not create an unbounded queue")
	}
	done := make(chan struct{})
	go func() { leaveA(); close(done) }()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("leaving must not block on a full queue")
	}
	if voiceEvent(t, streamB).Type != "peer-left" {
		t.Fatal("stale signaling should be cleared before the leave notice")
	}
}

package usecase

import (
	"crypto/rand"
	"errors"
	"time"

	"github.com/putriindaht/leftnt/server/internal/domain"
	"github.com/putriindaht/leftnt/server/internal/protocol"
)

var (
	ErrVoiceConnected = errors.New("voice is already joined in another window")
	ErrVoiceWaiting   = errors.New("your partner has not joined voice yet")
	ErrVoiceBusy      = errors.New("voice connection is busy; leave and join again")
	ErrInvalidSignal  = errors.New("invalid voice connection message")
)

// Voice uses a separate socket so joining or muting cannot claim the game seat.
func (s *RoomService) SubscribeVoice(code, token string) (<-chan protocol.VoiceMessage, func(), error) {
	code = normalizeCode(code)
	var stream chan protocol.VoiceMessage
	var role domain.Role
	err := s.repository.Update(code, func(r *RoomRecord) error {
		var err error
		role, err = roleFor(r, token)
		if err != nil {
			return err
		}
		if r.VoiceStreams == nil {
			r.VoiceStreams = map[domain.Role]chan protocol.VoiceMessage{}
		}
		if r.VoiceStreams[role] != nil {
			return ErrVoiceConnected
		}
		stream = make(chan protocol.VoiceMessage, 64)
		r.VoiceStreams[role] = stream
		r.LastActive = time.Now()
		ready := len(r.VoiceStreams) == 2
		if ready {
			r.VoiceCallID = rand.Text()
		}
		for member, target := range r.VoiceStreams {
			target <- protocol.VoiceMessage{Type: "ready", Role: member, PeerReady: ready, CallID: r.VoiceCallID}
		}
		return nil
	})
	if err != nil {
		return nil, nil, err
	}
	return stream, func() {
		_ = s.repository.Update(code, func(r *RoomRecord) error {
			if r.VoiceStreams[role] != stream {
				return nil
			}
			delete(r.VoiceStreams, role)
			close(stream)
			oldCall := r.VoiceCallID
			r.VoiceCallID = ""
			r.LastActive = time.Now()
			for _, target := range r.VoiceStreams {
				// Clear stale negotiation data before notifying the remaining peer.
				draining := true
				for draining {
					select {
					case <-target:
					default:
						draining = false
					}
				}
				target <- protocol.VoiceMessage{Type: "peer-left", CallID: oldCall}
			}
			return nil
		})
	}, nil
}

func (s *RoomService) RelayVoice(code, token string, signal protocol.VoiceMessage) error {
	return s.repository.Update(normalizeCode(code), func(r *RoomRecord) error {
		role, err := roleFor(r, token)
		if err != nil {
			return err
		}
		if r.VoiceStreams[role] == nil {
			return ErrInvalidSession
		}
		// Delayed ICE from a previous call must not break a new negotiation.
		if signal.CallID != r.VoiceCallID || signal.CallID == "" {
			return nil
		}
		peer := domain.Explorer
		if role == domain.Explorer {
			peer = domain.Navigator
		}
		target := r.VoiceStreams[peer]
		if target == nil {
			return ErrVoiceWaiting
		}
		message := protocol.VoiceMessage{Type: signal.Type, CallID: r.VoiceCallID}
		switch signal.Type {
		case "description":
			if signal.Description == nil || len(signal.Description.SDP) == 0 || len(signal.Description.SDP) > 24576 {
				return ErrInvalidSignal
			}
			if !((role == domain.Explorer && signal.Description.Type == "offer") || (role == domain.Navigator && signal.Description.Type == "answer")) {
				return ErrInvalidSignal
			}
			message.Description = signal.Description
		case "candidate":
			if signal.Candidate == nil || len(signal.Candidate.Candidate) > 2048 {
				return ErrInvalidSignal
			}
			message.Candidate = signal.Candidate
		default:
			return ErrInvalidSignal
		}
		select {
		case target <- message:
			return nil
		default:
			return ErrVoiceBusy
		}
	})
}

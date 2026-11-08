package protocol

import "github.com/putriindaht/leftnt/server/internal/domain"

type SessionDescription struct {
	Type string `json:"type"`
	SDP  string `json:"sdp"`
}

type ICECandidate struct {
	Candidate        string  `json:"candidate"`
	SDPMid           *string `json:"sdpMid"`
	SDPMLineIndex    *uint16 `json:"sdpMLineIndex"`
	UsernameFragment *string `json:"usernameFragment,omitempty"`
}

type ICEServer struct {
	URLs       []string `json:"urls"`
	Username   string   `json:"username,omitempty"`
	Credential string   `json:"credential,omitempty"`
}

// VoiceMessage carries connection setup only. Audio never travels through Go.
type VoiceMessage struct {
	Type        string              `json:"type"`
	Token       string              `json:"token,omitempty"`
	CallID      string              `json:"callId,omitempty"`
	Role        domain.Role         `json:"role,omitempty"`
	PeerReady   bool                `json:"peerReady,omitempty"`
	ICEServers  []ICEServer         `json:"iceServers,omitempty"`
	Description *SessionDescription `json:"description,omitempty"`
	Candidate   *ICECandidate       `json:"candidate,omitempty"`
	Error       string              `json:"error,omitempty"`
}

package codec

import (
	"fmt"

	"github.com/gofrs/uuid/v5"
)

// LobbyPendingSessionCancel represents a message from client to the server, indicating intent to cancel pending matchmaker operations.
type LobbyPendingSessionCancel struct {
	Session uuid.UUID // The user's session token.
}

// ToString returns a string representation of the LobbyPendingSessionCancel message.
func (m *LobbyPendingSessionCancel) String() string {
	return fmt.Sprintf("%T(session=%v)", m, m.Session)
}

// NewLobbyPendingSessionCancelWithSession initializes a new LobbyPendingSessionCancel message with the provided session token.
func NewLobbyPendingSessionCancel(session uuid.UUID) *LobbyPendingSessionCancel {
	return &LobbyPendingSessionCancel{
		Session: session,
	}
}

func (m *LobbyPendingSessionCancel) Stream(s *EasyStream) error {
	return RunErrorFunctions([]func() error{
		func() error { return s.StreamGUID(&m.Session) },
	})
}

func (m *LobbyPendingSessionCancel) SessionID() uuid.UUID {
	return m.Session
}

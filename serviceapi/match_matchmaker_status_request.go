package codec

import (
	"fmt"
)

// LobbyMatchmakerStatusRequest is a message from client to server, requesting the status
// of a pending matchmaking operation.
type LobbyMatchmakerStatusRequest struct {
	Unk0 byte
	Unk1 byte
	Unk2 byte
}

// NewLobbyMatchmakerStatusRequest initializes a new LobbyMatchmakerStatusRequest message.
func NewLobbyMatchmakerStatusRequest() *LobbyMatchmakerStatusRequest {
	return &LobbyMatchmakerStatusRequest{}
}

// Stream streams the message data in/out based on the streaming mode set.
func (m *LobbyMatchmakerStatusRequest) Stream(s *EasyStream) error {
	return RunErrorFunctions([]func() error{
		func() error { return s.StreamByte(&m.Unk0) },
	})
}

func (m *LobbyMatchmakerStatusRequest) String() string {
	return fmt.Sprintf("%T(unk0=%d, unk1=%d, unk2=%d)", m, m.Unk0, m.Unk1, m.Unk2)
}

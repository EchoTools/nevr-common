package codec

import (
	"encoding/binary"
	"fmt"
	"strings"

	"github.com/gofrs/uuid/v5"
)

// LobbyPlayerSessionsRequest is a message from client to server, asking it to obtain game server sessions for a given list of user identifiers.
type LobbyPlayerSessionsRequest struct {
	LoginSessionID uuid.UUID
	XPID           XPID
	LobbyID        uuid.UUID
	Platform       Symbol
	PlayerXPIDs    []XPID
}

func (m *LobbyPlayerSessionsRequest) Stream(s *EasyStream) error {
	playerCount := uint64(len(m.PlayerXPIDs))
	return RunErrorFunctions([]func() error{
		func() error { return s.StreamGUID(&m.LoginSessionID) },
		func() error { return s.StreamNumber(binary.LittleEndian, &m.XPID.PlatformCode) },
		func() error { return s.StreamNumber(binary.LittleEndian, &m.XPID.AccountId) },
		func() error { return s.StreamGUID(&m.LobbyID) },
		func() error { return s.StreamNumber(binary.LittleEndian, &m.Platform) },
		func() error { return s.StreamNumber(binary.LittleEndian, &playerCount) },
		func() error {
			if s.Mode == DecodeMode {
				m.PlayerXPIDs = make([]XPID, playerCount)
			}
			for i := range m.PlayerXPIDs {
				if err := s.StreamStruct(&m.PlayerXPIDs[i]); err != nil {
					return err
				}
			}
			return nil
		},
	})
}

func (m *LobbyPlayerSessionsRequest) String() string {
	xpidStrs := make([]string, len(m.PlayerXPIDs))
	for i, id := range m.PlayerXPIDs {
		xpidStrs[i] = fmt.Sprintf("%s", id.String())
	}
	xpidstrs := strings.Join(xpidStrs, ", ")
	return fmt.Sprintf("%T(login_session_id=%s, evr_id=%s, lobby_id=%s, evr_ids=%s)", m, m.LoginSessionID, m.XPID, m.LobbyID, xpidstrs)
}

func (m *LobbyPlayerSessionsRequest) SessionID() uuid.UUID {
	return m.LoginSessionID
}

func (m *LobbyPlayerSessionsRequest) GetXPID() XPID {
	return m.XPID
}

func (m *LobbyPlayerSessionsRequest) LobbySessionID() uuid.UUID {
	return m.LobbyID
}

package codec

import (
	"encoding/binary"
	"fmt"

	"github.com/gofrs/uuid/v5"
)

type LoginSuccess struct {
	Session uuid.UUID
	XPID    XPID
}

func NewLoginSuccess(session uuid.UUID, xpid XPID) *LoginSuccess {
	return &LoginSuccess{
		Session: session,
		XPID:    xpid,
	}
}

func (m LoginSuccess) String() string {
	return fmt.Sprintf("%Tsession=%v, user_id=%s)", m, m.Session, m.XPID.String())
}

func (m *LoginSuccess) Stream(s *EasyStream) error {
	return RunErrorFunctions([]func() error{
		func() error { return s.StreamGUID(&m.Session) },
		func() error { return s.StreamNumber(binary.LittleEndian, &m.XPID.PlatformCode) },
		func() error { return s.StreamNumber(binary.LittleEndian, &m.XPID.AccountId) },
	})
}

func (m *LoginSuccess) SessionID() uuid.UUID {
	return m.Session
}

func (m *LoginSuccess) GetXPID() XPID {
	return m.XPID
}

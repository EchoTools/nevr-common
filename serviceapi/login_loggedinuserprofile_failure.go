package codec

import (
	"encoding/binary"
	"fmt"
	"net/http"
)

// nakama -> client: failure response to LoggedInUserProfileFailure.
type LoggedInUserProfileFailure struct {
	XPID         XPID
	StatusCode   uint64 // HTTP status code
	ErrorMessage string
}

func (m *LoggedInUserProfileFailure) String() string {
	return fmt.Sprintf("%T(user_id=%v, status=%v, msg=\"%s\")", m, m.XPID, http.StatusText(int(m.StatusCode)), m.ErrorMessage)
}

func (m *LoggedInUserProfileFailure) Stream(s *EasyStream) error {
	return RunErrorFunctions([]func() error{
		func() error { return s.StreamNumber(binary.LittleEndian, &m.XPID.PlatformCode) },
		func() error { return s.StreamNumber(binary.LittleEndian, &m.XPID.AccountId) },
		func() error { return s.StreamNumber(binary.LittleEndian, &m.StatusCode) },
		func() error { return s.StreamNullTerminatedString(&m.ErrorMessage) },
	})
}

func NewLoggedInUserProfileFailure(xpid XPID, statusCode int, message string) *LoggedInUserProfileFailure {
	return &LoggedInUserProfileFailure{
		XPID:         xpid,
		StatusCode:   uint64(statusCode),
		ErrorMessage: message,
	}
}

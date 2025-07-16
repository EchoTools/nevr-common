package codec

import (
	"encoding/binary"
	"fmt"
	"net/http"
)

// SNSLobbySmiteEntrant represents a message from server to client indicating a failure in OtherUserProfileRequest.
type SNSLobbySmiteEntrant struct {
	XPID       XPID   // The identifier of the associated user.
	StatusCode uint64 // The status code returned with the failure. (These are http status codes)
	Message    string // The message returned with the failure.
}

func NewSNSLobbySmiteEntrant(xpid XPID, statusCode uint64, message string) *SNSLobbySmiteEntrant {
	return &SNSLobbySmiteEntrant{
		XPID:       xpid,
		StatusCode: statusCode,
		Message:    message,
	}
}

func (m *SNSLobbySmiteEntrant) Stream(s *EasyStream) error {
	return RunErrorFunctions([]func() error{
		func() error { return s.StreamStruct(&m.XPID) },
		func() error { return s.StreamNumber(binary.LittleEndian, &m.StatusCode) },
		func() error { return s.StreamNullTerminatedString(&m.Message) },
	})
}

func (m *SNSLobbySmiteEntrant) String() string {
	return fmt.Sprintf("%T(user_id=%v, status=%v, msg=\"%s\")", m, m.XPID, http.StatusText(int(m.StatusCode)), m.Message)
}

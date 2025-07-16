package codec

import (
	"encoding/binary"
	"fmt"
)

type UpdateProfileFailure struct {
	XPID       XPID
	statusCode uint64 // HTTP Status Code
	Message    string
}

func (lr *UpdateProfileFailure) String() string {
	return fmt.Sprintf("%T(user_id=%s, status_code=%d, msg='%s')", lr, lr.XPID.String(), lr.statusCode, lr.Message)
}

func (m *UpdateProfileFailure) Stream(s *EasyStream) error {
	return RunErrorFunctions([]func() error{
		func() error { return s.StreamNumber(binary.LittleEndian, &m.XPID) },
		func() error { return s.StreamNumber(binary.LittleEndian, &m.statusCode) },
		func() error { return s.StreamNullTerminatedString(&m.Message) },
	})
}

func NewUpdateProfileFailure(xpid XPID, statusCode uint64, message string) *UpdateProfileFailure {
	return &UpdateProfileFailure{
		XPID:       xpid,
		statusCode: statusCode,
		Message:    message,
	}
}

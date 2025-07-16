package codec

import (
	"encoding/binary"
	"fmt"
)

// DocumentFailure represents a message from server to client indicating a failed DocumentFailurev2.
type DocumentFailure struct {
	Unk0    uint64 // TODO: Figure this out
	Unk1    uint64 // TODO: Figure this out
	Message string // The message to return with the failure.
}

func NewDocumentFailureWithArgs(message string) *DocumentFailure {
	return &DocumentFailure{
		Unk0:    0,
		Unk1:    0,
		Message: message,
	}
}

func (m *DocumentFailure) Stream(s *EasyStream) error {
	return RunErrorFunctions([]func() error{
		func() error { return s.StreamNumber(binary.LittleEndian, &m.Unk0) },
		func() error { return s.StreamNumber(binary.LittleEndian, &m.Unk1) },
		func() error { return s.StreamNullTerminatedString(&m.Message) },
	})
}
func (m *DocumentFailure) String() string {
	return fmt.Sprintf("%T(unk0=%d, unk1=%d, msg='%s')", m, m.Unk0, m.Unk1, m.Message)
}

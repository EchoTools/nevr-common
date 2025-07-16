package codec

import (
	"fmt"
)

type DocumentRequest struct {
	Language string
	Type     string
}

func (m DocumentRequest) String() string {
	return fmt.Sprintf("%T(lang=%v, t=%v)", m, m.Language, m.Type)
}

func (m *DocumentRequest) Stream(s *EasyStream) error {
	return RunErrorFunctions([]func() error{
		func() error { return s.StreamNullTerminatedString(&m.Language) },
		func() error { return s.StreamNullTerminatedString(&m.Type) },
	})
}

func NewDocumentRequest(t, language string) *DocumentRequest {
	return &DocumentRequest{
		Language: string(language),
		Type:     string(t),
	}
}

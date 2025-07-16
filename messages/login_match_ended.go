package codec

import (
	"fmt"
)

type MatchEnded struct {
	// TODO
}

func (m *MatchEnded) Stream(s *EasyStream) error {
	return RunErrorFunctions([]func() error{
		func() error { return nil },
	})
}

func (m MatchEnded) String() string {
	return fmt.Sprintf("%T()", m)
}

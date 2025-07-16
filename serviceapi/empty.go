package codec

import "fmt"

type EmptyMessage struct{}

func (m *EmptyMessage) Stream(s *EasyStream) error {
	return s.StreamNull(1)
}
func (m *EmptyMessage) String() string {
	return fmt.Sprintf("%T()", m)
}
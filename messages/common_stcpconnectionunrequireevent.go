package codec

var UnrequireMessagePayload, _ = Marshal(&TCPConnectionUnrequireEvent{})
var RequireMessagePayload, _ = Marshal(&TCPConnectionUnrequireEvent{})

type TCPConnectionEvent struct{}

func (m *TCPConnectionEvent) Stream(s *EasyStream) error {
	return s.StreamNull(1)
}

type TCPConnectionUnrequireEvent struct {
	TCPConnectionEvent
}
type TCPConnectionRequireEvent struct {
	TCPConnectionEvent
}

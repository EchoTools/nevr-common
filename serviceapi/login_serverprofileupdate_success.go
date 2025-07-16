package codec

import (
	"fmt"
)

type UserServerProfileUpdateSuccess struct {
	XPID XPID
}

func (lr *UserServerProfileUpdateSuccess) String() string {
	return fmt.Sprintf("%T(user_id=%s)", lr, lr.XPID.String())
}
func (m *UserServerProfileUpdateSuccess) Stream(s *EasyStream) error {
	return s.StreamStruct(&m.XPID)
}
func NewUserServerProfileUpdateSuccess(userId XPID) *UserServerProfileUpdateSuccess {
	return &UserServerProfileUpdateSuccess{
		XPID: userId,
	}
}

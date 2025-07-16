package codec

import (
	"encoding/binary"
	"fmt"

	"github.com/gofrs/uuid/v5"
)

// client -> nakama: request the user profile for their logged-in account.
type LoggedInUserProfileRequest struct {
	Session            uuid.UUID
	XPID               XPID
	ProfileRequestData ProfileRequestData
}

func (r LoggedInUserProfileRequest) String() string {
	return fmt.Sprintf("%T(session=%v, user_id=%v, profile_request=%v)", r, r.Session, r.XPID, r.ProfileRequestData)
}

func (m *LoggedInUserProfileRequest) Stream(s *EasyStream) error {
	return RunErrorFunctions([]func() error{
		func() error { return s.StreamGUID(&m.Session) },
		func() error { return s.StreamNumber(binary.LittleEndian, &m.XPID.PlatformCode) },
		func() error { return s.StreamNumber(binary.LittleEndian, &m.XPID.AccountId) },
		func() error { return s.StreamJson(&m.ProfileRequestData, true, NoCompression) },
	})
}

func NewLoggedInUserProfileRequest(session uuid.UUID, xpid XPID, profileRequestData ProfileRequestData) LoggedInUserProfileRequest {
	return LoggedInUserProfileRequest{
		Session:            session,
		XPID:               xpid,
		ProfileRequestData: profileRequestData,
	}
}
func (m *LoggedInUserProfileRequest) SessionID() uuid.UUID {
	return m.Session
}

func (m *LoggedInUserProfileRequest) GetXPID() XPID {
	return m.XPID
}

type ProfileRequestData struct {
	Defaultclientprofileid string       `json:"defaultclientprofileid"`
	Defaultserverprofileid string       `json:"defaultserverprofileid"`
	Unlocksetids           Unlocksetids `json:"unlocksetids"`
	Statgroupids           Statgroupids `json:"statgroupids"`
}

type Statgroupids struct {
	Arena           map[string]interface{} `json:"arena"`
	ArenaPracticeAI map[string]interface{} `json:"arena_practice_ai"`
	ArenaPublicAI   map[string]interface{} `json:"arena_public_ai"`
	Combat          map[string]interface{} `json:"combat"`
}

type Unlocksetids struct {
	All map[string]interface{} `json:"all"`
}

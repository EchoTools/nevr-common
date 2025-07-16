package codec

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math/rand"

	"github.com/gofrs/uuid/v5"
)

type GameServerSessionStart struct {
	MatchID     uuid.UUID            // The identifier for the game server session to start.
	GroupID     uuid.UUID            // TODO: Unverified, suspected to be channel UUID.
	PlayerLimit byte                 // The maximum amount of players allowed to join the lobby.
	LobbyType   byte                 // The type of lobby
	Settings    LobbySessionSettings // The JSON settings associated with the session.
	Entrants    []EntrantDescriptor  // Information regarding entrants (e.g. including offline/local player ids, or AI bot platform ids).
}

func (s *GameServerSessionStart) String() string {
	return fmt.Sprintf("%T(session_id=%s, player_limit=%d, mode=%s, level=%s)",
		s, s.MatchID, s.PlayerLimit, ToSymbol(s.Settings.Mode).Token(), ToSymbol(s.Settings.Level).Token())
}

func NewGameServerSessionStart(sessionID uuid.UUID, channel uuid.UUID, playerLimit uint8, lobbyType uint8, appID string, mode Symbol, level Symbol, features []string, entrants []XPID) *GameServerSessionStart {
	descriptors := make([]EntrantDescriptor, len(entrants))
	for i, entrant := range entrants {
		descriptors[i] = *NewEntrantDescriptor(entrant)
	}

	return &GameServerSessionStart{
		MatchID:     sessionID,
		GroupID:     channel,
		PlayerLimit: byte(playerLimit),
		LobbyType:   byte(lobbyType),
		Settings:    *NewSessionSettings(appID, mode, level, features),
		Entrants:    descriptors,
	}
}

type LobbySessionSettings struct {
	AppID    string   `json:"appid"`
	Mode     int64    `json:"gametype"`
	Level    int64    `json:"level"`
	Features []string `json:"features,omitempty"`
}

func (s *LobbySessionSettings) MarshalJSON() ([]byte, error) {
	if s.Level == 0 {
		s.Level = int64(LevelUnspecified)
	}
	type Alias LobbySessionSettings
	return json.Marshal(&struct {
		Level int64 `json:"level"`
		*Alias
	}{
		Level: s.Level,
		Alias: (*Alias)(s),
	})
}

func (s *LobbySessionSettings) UnmarshalJSON(data []byte) error {
	type Alias LobbySessionSettings
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(s),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	s.Level = aux.Level
	s.Features = aux.Features
	s.AppID = aux.AppID
	s.Mode = aux.Mode
	return nil
}

func NewSessionSettings(appID string, mode Symbol, level Symbol, features []string) *LobbySessionSettings {
	if level == 0 {
		level = LevelUnspecified
	}
	settings := LobbySessionSettings{
		AppID:    appID,
		Mode:     int64(mode),
		Level:    int64(level),
		Features: features,
	}
	if level != 0 {
		l := int64(level)
		settings.Level = l
	}
	return &settings
}

func (s *LobbySessionSettings) String() string {
	b, err := json.Marshal(s)
	if err != nil {
		panic(err)
	}
	return string(b)
}

type EntrantDescriptor struct {
	Unk0  uuid.UUID
	XPID  XPID
	Flags uint64
}

func (m *EntrantDescriptor) String() string {
	return fmt.Sprintf("EREntrantDescriptor(unk0=%s, player_id=%s, flags=%d)", m.Unk0, m.XPID.String(), m.Flags)
}

func NewEntrantDescriptor(playerId XPID) *EntrantDescriptor {
	return &EntrantDescriptor{
		Unk0:  uuid.Must(uuid.NewV4()),
		XPID:  playerId,
		Flags: 0x0044BB8000,
	}
}

func RandomBotEntrantDescriptor() EntrantDescriptor {
	botuuid, _ := uuid.NewV4()
	return EntrantDescriptor{
		Unk0:  botuuid,
		XPID:  XPID{PlatformCode: BOT, AccountId: rand.Uint64()},
		Flags: 0x0044BB8000,
	}
}

func (m *GameServerSessionStart) Stream(s *EasyStream) error {
	finalStructCount := byte(len(m.Entrants))
	pad1 := byte(0)
	return RunErrorFunctions([]func() error{
		func() error { return s.StreamGUID(&m.MatchID) },
		func() error { return s.StreamGUID(&m.GroupID) },
		func() error { return s.StreamByte(&m.PlayerLimit) },
		func() error { return s.StreamNumber(binary.LittleEndian, &finalStructCount) },
		func() error { return s.StreamByte(&m.LobbyType) },
		func() error { return s.StreamByte(&pad1) },
		func() error { return s.StreamJson(&m.Settings, true, NoCompression) },
		func() error {
			if s.Mode == DecodeMode {
				m.Entrants = make([]EntrantDescriptor, finalStructCount)
			}
			for _, entrant := range m.Entrants {
				err := RunErrorFunctions([]func() error{
					func() error { return s.StreamGUID(&entrant.Unk0) },
					func() error { return s.StreamStruct(&entrant.XPID) },
					func() error { return s.StreamNumber(binary.LittleEndian, &entrant.Flags) },
				})
				if err != nil {
					return err
				}
			}
			return nil
		},
	})

}

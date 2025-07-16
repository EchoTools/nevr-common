package codec

import "math/rand"

type LobbyType uint8

const (
	PublicLobby     LobbyType = iota // An active public lobby
	PrivateLobby                     // An active private lobby
	UnassignedLobby                  // An unloaded lobby
)

const (
	TeamUnassigned int = iota - 1
	TeamBlue
	TeamOrange
	TeamSpectator
	TeamSocial
	TeamModerator
)

var (
	ModeUnloaded        Symbol = ToSymbol("")                      // Unloaded Lobby
	ModeSocialPublic    Symbol = ToSymbol("social_2.0")            // Public Social Lobby
	ModeSocialPrivate   Symbol = ToSymbol("social_2.0_private")    // Private Social Lobby
	ModeSocialNPE       Symbol = ToSymbol("social_2.0_npe")        // Social Lobby NPE
	ModeArenaPublic     Symbol = ToSymbol("echo_arena")            // Public Echo Arena
	ModeArenaPrivate    Symbol = ToSymbol("echo_arena_private")    // Private Echo Arena
	ModeArenaTournment  Symbol = ToSymbol("echo_arena_tournament") // Echo Arena Tournament
	ModeArenaPublicAI   Symbol = ToSymbol("echo_arena_public_ai")  // Public Echo Arena AI
	ModeArenaPracticeAI Symbol = ToSymbol("echo_arena_practice")   // Echo Arena Practice

	ModeEchoCombatTournament Symbol = ToSymbol("echo_combat_tournament") // Echo Combat Tournament
	ModeCombatPublic         Symbol = ToSymbol("echo_combat")            // Echo Combat
	ModeCombatPrivate        Symbol = ToSymbol("echo_combat_private")    // Private Echo Combat

	LevelUnloaded     Symbol = Symbol(0)                          // Unloaded Lobby
	LevelSocial       Symbol = ToSymbol("mpl_lobby_b2")           // Social Lobby
	LevelUnspecified  Symbol = Symbol(0xffffffffffffffff)         // Unspecified Level
	LevelArena        Symbol = ToSymbol("mpl_arena_a")            // Echo Arena
	ModeArenaTutorial Symbol = ToSymbol("mpl_tutorial_arena")     // Echo Arena Tutorial
	LevelFission      Symbol = ToSymbol("mpl_combat_fission")     // Echo Combat
	LevelCombustion   Symbol = ToSymbol("mpl_combat_combustion")  // Echo Combat
	LevelDyson        Symbol = ToSymbol("mpl_combat_dyson")       // Echo Combat
	LevelGauss        Symbol = ToSymbol("mpl_combat_gauss")       // Echo Combat
	LevelPebbles      Symbol = ToSymbol("mpl_combat_pebbles")     // Echo Combat
	LevelPtyPebbles   Symbol = ToSymbol("pty_mpl_combat_pebbles") // Echo Combat

	// Valid levels by game mode
	LevelsByMode = map[Symbol][]Symbol{
		ModeArenaPublic:          {LevelArena},
		ModeArenaPrivate:         {LevelArena},
		ModeArenaTournment:       {LevelArena},
		ModeArenaPublicAI:        {LevelArena},
		ModeArenaTutorial:        {LevelArena},
		ModeSocialPublic:         {LevelSocial},
		ModeSocialPrivate:        {LevelSocial},
		ModeSocialNPE:            {LevelSocial},
		ModeCombatPublic:         {LevelCombustion, LevelDyson, LevelFission, LevelGauss},
		ModeCombatPrivate:        {LevelCombustion, LevelDyson, LevelFission, LevelGauss},
		ModeEchoCombatTournament: {LevelCombustion, LevelDyson, LevelFission, LevelGauss},
	}

	// Valid roles for the game mode
	RolesByMode = map[Symbol][]int{
		ModeCombatPublic:  {TeamBlue, TeamOrange, TeamSpectator, TeamModerator},
		ModeArenaPublic:   {TeamBlue, TeamOrange, TeamSpectator, TeamModerator},
		ModeCombatPrivate: {TeamBlue, TeamOrange, TeamSpectator, TeamModerator},
		ModeArenaPrivate:  {TeamBlue, TeamOrange, TeamSpectator, TeamModerator},
		ModeSocialPublic:  {TeamSocial, TeamModerator},
		ModeSocialPrivate: {TeamSocial, TeamModerator},
	}

	// Roles that may be specified by the player when finding/joining a lobby session.
	AlignmentsByMode = map[Symbol][]int{
		ModeCombatPublic:  {TeamUnassigned, TeamSpectator},
		ModeArenaPublic:   {TeamUnassigned, TeamSpectator},
		ModeCombatPrivate: {TeamUnassigned, TeamSpectator, TeamBlue, TeamOrange},
		ModeArenaPrivate:  {TeamUnassigned, TeamSpectator, TeamBlue, TeamOrange},
		ModeSocialPublic:  {TeamUnassigned, TeamModerator, TeamSocial},
		ModeSocialPrivate: {TeamUnassigned, TeamModerator, TeamSocial},
	}
)

func RandomLevelByMode(mode Symbol) Symbol {
	levels, ok := LevelsByMode[mode]
	if !ok {
		return LevelUnspecified
	}
	return levels[rand.Intn(len(levels))]
}

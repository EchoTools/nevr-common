package nevrcap

import (
	"github.com/echotools/nevr-common/gen/go/apigame"
	"github.com/echotools/nevr-common/gen/go/rtapi"
)

// EventDetector efficiently detects events between consecutive frames
type EventDetector struct {
	// Cache for player states to avoid repeated lookups
	prevPlayersBySlot map[int32]*apigame.TeamMember
	prevScoreboard    *ScoreboardState
	prevDiscState     *DiscState
}

// ScoreboardState represents the scoring state
type ScoreboardState struct {
	BluePoints       int32
	OrangePoints     int32
	BlueRoundScore   int32
	OrangeRoundScore int32
	GameClock        string
}

// DiscState represents disc possession state
type DiscState struct {
	HasPossession bool
	PlayerSlot    int32 // -1 if no player has possession
}

// NewEventDetector creates a new event detector
func NewEventDetector() *EventDetector {
	return &EventDetector{
		prevPlayersBySlot: make(map[int32]*apigame.TeamMember),
	}
}

// DetectEvents analyzes two consecutive frames and returns detected events
func (ed *EventDetector) DetectEvents(prevFrame, currentFrame *rtapi.LobbySessionStateFrame) []*rtapi.LobbySessionEvent {
	var events []*rtapi.LobbySessionEvent

	// Update player tracking
	currentPlayersBySlot := ed.buildPlayerSlotMap(currentFrame)

	// Detect player events
	events = append(events, ed.detectPlayerEvents(currentPlayersBySlot)...)

	// Detect scoreboard events
	events = append(events, ed.detectScoreboardEvents(prevFrame.Session, currentFrame.Session)...)

	// Detect disc events
	events = append(events, ed.detectDiscEvents(prevFrame.Session, currentFrame.Session)...)

	// Detect stat-based events
	events = append(events, ed.detectStatEvents(currentPlayersBySlot)...)

	// Update cached state for next comparison
	ed.prevPlayersBySlot = currentPlayersBySlot
	ed.updateCachedState(currentFrame)

	return events
}

// buildPlayerSlotMap creates a map of player slot to player for efficient lookup
func (ed *EventDetector) buildPlayerSlotMap(frame *rtapi.LobbySessionStateFrame) map[int32]*apigame.TeamMember {
	playersBySlot := make(map[int32]*apigame.TeamMember)

	for _, team := range frame.Session.Teams {
		for _, player := range team.Players {
			playersBySlot[player.SlotNumber] = player
		}
	}

	return playersBySlot
}

// detectPlayerEvents detects player join/leave/team switch events
func (ed *EventDetector) detectPlayerEvents(currentPlayers map[int32]*apigame.TeamMember) []*rtapi.LobbySessionEvent {
	var events []*rtapi.LobbySessionEvent

	// Detect new players (joined)
	for slot, player := range currentPlayers {
		if _, exists := ed.prevPlayersBySlot[slot]; !exists {
			events = append(events, &rtapi.LobbySessionEvent{
				Payload: &rtapi.LobbySessionEvent_PlayerJoined{
					PlayerJoined: &rtapi.PlayerJoined{
						Player: player,
						Role:   ed.determinePlayerRole(player),
					},
				},
			})
		}
	}

	// Detect missing players (left)
	for slot, prevPlayer := range ed.prevPlayersBySlot {
		if _, exists := currentPlayers[slot]; !exists {
			events = append(events, &rtapi.LobbySessionEvent{
				Payload: &rtapi.LobbySessionEvent_PlayerLeft{
					PlayerLeft: &rtapi.PlayerLeft{
						PlayerSlot:  slot,
						DisplayName: prevPlayer.DisplayName,
					},
				},
			})
		}
	}

	return events
}

// detectScoreboardEvents detects scoring and round changes
func (ed *EventDetector) detectScoreboardEvents(prevSession, currentSession *apigame.SessionResponse) []*rtapi.LobbySessionEvent {
	var events []*rtapi.LobbySessionEvent

	currentScoreboard := &ScoreboardState{
		BluePoints:       currentSession.BluePoints,
		OrangePoints:     currentSession.OrangePoints,
		BlueRoundScore:   currentSession.BlueRoundScore,
		OrangeRoundScore: currentSession.OrangeRoundScore,
		GameClock:        currentSession.GameClockDisplay,
	}

	if ed.prevScoreboard != nil {
		// Check for score changes
		if currentScoreboard.BluePoints != ed.prevScoreboard.BluePoints ||
			currentScoreboard.OrangePoints != ed.prevScoreboard.OrangePoints ||
			currentScoreboard.BlueRoundScore != ed.prevScoreboard.BlueRoundScore ||
			currentScoreboard.OrangeRoundScore != ed.prevScoreboard.OrangeRoundScore {

			events = append(events, &rtapi.LobbySessionEvent{
				Payload: &rtapi.LobbySessionEvent_ScoreboardUpdated{
					ScoreboardUpdated: &rtapi.ScoreboardUpdated{
						BluePoints:       currentScoreboard.BluePoints,
						OrangePoints:     currentScoreboard.OrangePoints,
						BlueRoundScore:   currentScoreboard.BlueRoundScore,
						OrangeRoundScore: currentScoreboard.OrangeRoundScore,
						GameClockDisplay: currentScoreboard.GameClock,
					},
				},
			})
		}

		// Check for goal scored
		if currentSession.LastScore != nil {
			// This is a simple heuristic - in practice, you might want more sophisticated detection
			events = append(events, &rtapi.LobbySessionEvent{
				Payload: &rtapi.LobbySessionEvent_GoalScored{
					GoalScored: &rtapi.GoalScored{
						ScoreDetails: currentSession.LastScore,
					},
				},
			})
		}
	}

	return events
}

// detectDiscEvents detects disc possession changes and throws
func (ed *EventDetector) detectDiscEvents(prevSession, currentSession *apigame.SessionResponse) []*rtapi.LobbySessionEvent {
	var events []*rtapi.LobbySessionEvent

	// Find current disc possession
	currentDiscState := ed.getDiscState(currentSession)

	if ed.prevDiscState != nil {
		// Check for possession change
		if currentDiscState.PlayerSlot != ed.prevDiscState.PlayerSlot {
			events = append(events, &rtapi.LobbySessionEvent{
				Payload: &rtapi.LobbySessionEvent_DiscPossessionChanged{
					DiscPossessionChanged: &rtapi.DiscPossessionChanged{
						PlayerSlot:   currentDiscState.PlayerSlot,
						PreviousSlot: ed.prevDiscState.PlayerSlot,
					},
				},
			})
		}
	}

	// Check for disc thrown (if last throw info is present)
	if currentSession.LastThrow != nil {
		// Find the player who threw
		for _, team := range currentSession.Teams {
			for _, player := range team.Players {
				if player.HasPossession {
					events = append(events, &rtapi.LobbySessionEvent{
						Payload: &rtapi.LobbySessionEvent_DiscThrown{
							DiscThrown: &rtapi.DiscThrown{
								PlayerSlot:   player.SlotNumber,
								ThrowDetails: currentSession.LastThrow,
							},
						},
					})
					break
				}
			}
		}
	}

	return events
}

// detectStatEvents detects changes in player statistics
func (ed *EventDetector) detectStatEvents(currentPlayers map[int32]*apigame.TeamMember) []*rtapi.LobbySessionEvent {
	var events []*rtapi.LobbySessionEvent

	for slot, player := range currentPlayers {
		if prevPlayer, exists := ed.prevPlayersBySlot[slot]; exists {
			// Check each stat type for increments
			if player.Stats.Saves > prevPlayer.Stats.Saves {
				events = append(events, &rtapi.LobbySessionEvent{
					Payload: &rtapi.LobbySessionEvent_PlayerSave{
						PlayerSave: &rtapi.PlayerSave{
							PlayerSlot: slot,
							TotalSaves: player.Stats.Saves,
						},
					},
				})
			}

			if player.Stats.Stuns > prevPlayer.Stats.Stuns {
				events = append(events, &rtapi.LobbySessionEvent{
					Payload: &rtapi.LobbySessionEvent_PlayerStun{
						PlayerStun: &rtapi.PlayerStun{
							PlayerSlot: slot,
							TotalStuns: player.Stats.Stuns,
						},
					},
				})
			}

			if player.Stats.Passes > prevPlayer.Stats.Passes {
				events = append(events, &rtapi.LobbySessionEvent{
					Payload: &rtapi.LobbySessionEvent_PlayerPass{
						PlayerPass: &rtapi.PlayerPass{
							PlayerSlot:  slot,
							TotalPasses: player.Stats.Passes,
						},
					},
				})
			}

			if player.Stats.Steals > prevPlayer.Stats.Steals {
				events = append(events, &rtapi.LobbySessionEvent{
					Payload: &rtapi.LobbySessionEvent_PlayerSteal{
						PlayerSteal: &rtapi.PlayerSteal{
							PlayerSlot:  slot,
							TotalSteals: player.Stats.Steals,
						},
					},
				})
			}

			if player.Stats.Blocks > prevPlayer.Stats.Blocks {
				events = append(events, &rtapi.LobbySessionEvent{
					Payload: &rtapi.LobbySessionEvent_PlayerBlock{
						PlayerBlock: &rtapi.PlayerBlock{
							PlayerSlot:  slot,
							TotalBlocks: player.Stats.Blocks,
						},
					},
				})
			}

			if player.Stats.Interceptions > prevPlayer.Stats.Interceptions {
				events = append(events, &rtapi.LobbySessionEvent{
					Payload: &rtapi.LobbySessionEvent_PlayerInterception{
						PlayerInterception: &rtapi.PlayerInterception{
							PlayerSlot:         slot,
							TotalInterceptions: player.Stats.Interceptions,
						},
					},
				})
			}

			if player.Stats.Assists > prevPlayer.Stats.Assists {
				events = append(events, &rtapi.LobbySessionEvent{
					Payload: &rtapi.LobbySessionEvent_PlayerAssist{
						PlayerAssist: &rtapi.PlayerAssist{
							PlayerSlot:   slot,
							TotalAssists: player.Stats.Assists,
						},
					},
				})
			}

			if player.Stats.ShotsTaken > prevPlayer.Stats.ShotsTaken {
				events = append(events, &rtapi.LobbySessionEvent{
					Payload: &rtapi.LobbySessionEvent_PlayerShotTaken{
						PlayerShotTaken: &rtapi.PlayerShotTaken{
							PlayerSlot: slot,
							TotalShots: player.Stats.ShotsTaken,
						},
					},
				})
			}
		}
	}

	return events
}

// getDiscState determines current disc possession state
func (ed *EventDetector) getDiscState(session *apigame.SessionResponse) *DiscState {
	for _, team := range session.Teams {
		for _, player := range team.Players {
			if player.HasPossession {
				return &DiscState{
					HasPossession: true,
					PlayerSlot:    player.SlotNumber,
				}
			}
		}
	}
	return &DiscState{
		HasPossession: false,
		PlayerSlot:    -1,
	}
}

// determinePlayerRole maps a player to their role
func (ed *EventDetector) determinePlayerRole(player *apigame.TeamMember) rtapi.Role {
	// This is a simplified mapping - you might need more sophisticated logic
	switch player.JerseyNumber {
	case -1:
		return rtapi.Role_SPECTATOR
	default:
		// Determine team based on some logic (this is simplified)
		if player.SlotNumber%2 == 0 {
			return rtapi.Role_BLUE_TEAM
		}
		return rtapi.Role_ORANGE_TEAM
	}
}

// updateCachedState updates the cached state for next comparison
func (ed *EventDetector) updateCachedState(frame *rtapi.LobbySessionStateFrame) {
	ed.prevScoreboard = &ScoreboardState{
		BluePoints:       frame.Session.BluePoints,
		OrangePoints:     frame.Session.OrangePoints,
		BlueRoundScore:   frame.Session.BlueRoundScore,
		OrangeRoundScore: frame.Session.OrangeRoundScore,
		GameClock:        frame.Session.GameClockDisplay,
	}

	ed.prevDiscState = ed.getDiscState(frame.Session)
}

// Reset clears the event detector state
func (ed *EventDetector) Reset() {
	ed.prevPlayersBySlot = make(map[int32]*apigame.TeamMember)
	ed.prevScoreboard = nil
	ed.prevDiscState = nil
}

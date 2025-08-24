package telemetry

import (
	"github.com/echotools/nevr-common/v3/gameapi"
)

// EventDetector efficiently detects events between consecutive frames
type EventDetector struct {
	// Cache for player states to avoid repeated lookups
	prevPlayersBySlot map[int32]*gameapi.TeamMember
	prevScoreboard    *ScoreboardState
	prevDiscState     *DiscState
}

// ScoreboardState represents the scoring state
type ScoreboardState struct {
	BluePoints      int32
	OrangePoints    int32
	BlueRoundScore  int32
	OrangeRoundScore int32
	GameClock       string
}

// DiscState represents disc possession state
type DiscState struct {
	HasPossession bool
	PlayerSlot    int32 // -1 if no player has possession
}

// NewEventDetector creates a new event detector
func NewEventDetector() *EventDetector {
	return &EventDetector{
		prevPlayersBySlot: make(map[int32]*gameapi.TeamMember),
	}
}

// DetectEvents analyzes two consecutive frames and returns detected events
func (ed *EventDetector) DetectEvents(prevFrame, currentFrame *LobbySessionStateFrame) []*LobbySessionEvent {
	var events []*LobbySessionEvent
	
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
func (ed *EventDetector) buildPlayerSlotMap(frame *LobbySessionStateFrame) map[int32]*gameapi.TeamMember {
	playersBySlot := make(map[int32]*gameapi.TeamMember)
	
	for _, team := range frame.Session.Teams {
		for _, player := range team.Players {
			playersBySlot[player.SlotNumber] = player
		}
	}
	
	return playersBySlot
}

// detectPlayerEvents detects player join/leave/team switch events
func (ed *EventDetector) detectPlayerEvents(currentPlayers map[int32]*gameapi.TeamMember) []*LobbySessionEvent {
	var events []*LobbySessionEvent
	
	// Detect new players (joined)
	for slot, player := range currentPlayers {
		if _, exists := ed.prevPlayersBySlot[slot]; !exists {
			events = append(events, &LobbySessionEvent{
				Payload: &LobbySessionEvent_PlayerJoined{
					PlayerJoined: &PlayerJoined{
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
			events = append(events, &LobbySessionEvent{
				Payload: &LobbySessionEvent_PlayerLeft{
					PlayerLeft: &PlayerLeft{
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
func (ed *EventDetector) detectScoreboardEvents(prevSession, currentSession *gameapi.SessionResponse) []*LobbySessionEvent {
	var events []*LobbySessionEvent
	
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
			
			events = append(events, &LobbySessionEvent{
				Payload: &LobbySessionEvent_ScoreboardUpdated{
					ScoreboardUpdated: &ScoreboardUpdated{
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
			events = append(events, &LobbySessionEvent{
				Payload: &LobbySessionEvent_GoalScored{
					GoalScored: &GoalScored{
						ScoreDetails: currentSession.LastScore,
					},
				},
			})
		}
	}
	
	return events
}

// detectDiscEvents detects disc possession changes and throws
func (ed *EventDetector) detectDiscEvents(prevSession, currentSession *gameapi.SessionResponse) []*LobbySessionEvent {
	var events []*LobbySessionEvent
	
	// Find current disc possession
	currentDiscState := ed.getDiscState(currentSession)
	
	if ed.prevDiscState != nil {
		// Check for possession change
		if currentDiscState.PlayerSlot != ed.prevDiscState.PlayerSlot {
			events = append(events, &LobbySessionEvent{
				Payload: &LobbySessionEvent_DiscPossessionChanged{
					DiscPossessionChanged: &DiscPossessionChanged{
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
					events = append(events, &LobbySessionEvent{
						Payload: &LobbySessionEvent_DiscThrown{
							DiscThrown: &DiscThrown{
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
func (ed *EventDetector) detectStatEvents(currentPlayers map[int32]*gameapi.TeamMember) []*LobbySessionEvent {
	var events []*LobbySessionEvent
	
	for slot, player := range currentPlayers {
		if prevPlayer, exists := ed.prevPlayersBySlot[slot]; exists {
			// Check each stat type for increments
			if player.Stats.Saves > prevPlayer.Stats.Saves {
				events = append(events, &LobbySessionEvent{
					Payload: &LobbySessionEvent_PlayerSave{
						PlayerSave: &PlayerSave{
							PlayerSlot: slot,
							TotalSaves: player.Stats.Saves,
						},
					},
				})
			}
			
			if player.Stats.Stuns > prevPlayer.Stats.Stuns {
				events = append(events, &LobbySessionEvent{
					Payload: &LobbySessionEvent_PlayerStun{
						PlayerStun: &PlayerStun{
							PlayerSlot: slot,
							TotalStuns: player.Stats.Stuns,
						},
					},
				})
			}
			
			if player.Stats.Passes > prevPlayer.Stats.Passes {
				events = append(events, &LobbySessionEvent{
					Payload: &LobbySessionEvent_PlayerPass{
						PlayerPass: &PlayerPass{
							PlayerSlot:  slot,
							TotalPasses: player.Stats.Passes,
						},
					},
				})
			}
			
			if player.Stats.Steals > prevPlayer.Stats.Steals {
				events = append(events, &LobbySessionEvent{
					Payload: &LobbySessionEvent_PlayerSteal{
						PlayerSteal: &PlayerSteal{
							PlayerSlot:  slot,
							TotalSteals: player.Stats.Steals,
						},
					},
				})
			}
			
			if player.Stats.Blocks > prevPlayer.Stats.Blocks {
				events = append(events, &LobbySessionEvent{
					Payload: &LobbySessionEvent_PlayerBlock{
						PlayerBlock: &PlayerBlock{
							PlayerSlot:  slot,
							TotalBlocks: player.Stats.Blocks,
						},
					},
				})
			}
			
			if player.Stats.Interceptions > prevPlayer.Stats.Interceptions {
				events = append(events, &LobbySessionEvent{
					Payload: &LobbySessionEvent_PlayerInterception{
						PlayerInterception: &PlayerInterception{
							PlayerSlot:         slot,
							TotalInterceptions: player.Stats.Interceptions,
						},
					},
				})
			}
			
			if player.Stats.Assists > prevPlayer.Stats.Assists {
				events = append(events, &LobbySessionEvent{
					Payload: &LobbySessionEvent_PlayerAssist{
						PlayerAssist: &PlayerAssist{
							PlayerSlot:   slot,
							TotalAssists: player.Stats.Assists,
						},
					},
				})
			}
			
			if player.Stats.ShotsTaken > prevPlayer.Stats.ShotsTaken {
				events = append(events, &LobbySessionEvent{
					Payload: &LobbySessionEvent_PlayerShotTaken{
						PlayerShotTaken: &PlayerShotTaken{
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
func (ed *EventDetector) getDiscState(session *gameapi.SessionResponse) *DiscState {
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
func (ed *EventDetector) determinePlayerRole(player *gameapi.TeamMember) Role {
	// This is a simplified mapping - you might need more sophisticated logic
	switch player.JerseyNumber {
	case -1:
		return Role_SPECTATOR
	default:
		// Determine team based on some logic (this is simplified)
		if player.SlotNumber%2 == 0 {
			return Role_BLUE_TEAM
		}
		return Role_ORANGE_TEAM
	}
}

// updateCachedState updates the cached state for next comparison
func (ed *EventDetector) updateCachedState(frame *LobbySessionStateFrame) {
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
	ed.prevPlayersBySlot = make(map[int32]*gameapi.TeamMember)
	ed.prevScoreboard = nil
	ed.prevDiscState = nil
}

// Package main demonstrates the wire size improvements of telemetry v2 over v1.
package main

import (
"fmt"
"time"

apigame "github.com/echotools/nevr-common/v4/gen/go/apigame"
spatial "github.com/echotools/nevr-common/v4/gen/go/spatial/v1"
telemetryv1 "github.com/echotools/nevr-common/v4/gen/go/telemetry/v1"
telemetryv2 "github.com/echotools/nevr-common/v4/gen/go/telemetry/v2"
"google.golang.org/protobuf/proto"
"google.golang.org/protobuf/types/known/timestamppb"
)

func main() {
fmt.Println("=== Telemetry V2 Wire Size Comparison ===\n")

// Test v2 frame with 2 players
v2Frame2 := createV2Frame(2)
v2Data2, _ := proto.Marshal(v2Frame2)
fmt.Printf("V2 Frame with 2 players: %d bytes\n", len(v2Data2))

// Test v2 frame with 10 players
v2Frame10 := createV2Frame(10)
v2Data10, _ := proto.Marshal(v2Frame10)
fmt.Printf("V2 Frame with 10 players: %d bytes\n", len(v2Data10))

// Test v1 frame for comparison
v1Frame2 := createV1Frame(2)
v1Data2, _ := proto.Marshal(v1Frame2)
fmt.Printf("\nV1 Frame with 2 players: %d bytes\n", len(v1Data2))

v1Frame10 := createV1Frame(10)
v1Data10, _ := proto.Marshal(v1Frame10)
fmt.Printf("V1 Frame with 10 players: %d bytes\n", len(v1Data10))

// Calculate savings
savings2 := float64(len(v1Data2)-len(v2Data2)) / float64(len(v1Data2)) * 100
savings10 := float64(len(v1Data10)-len(v2Data10)) / float64(len(v1Data10)) * 100

fmt.Printf("\n=== Wire Size Reduction ===\n")
fmt.Printf("2 players: %.1f%% reduction (%d → %d bytes)\n", savings2, len(v1Data2), len(v2Data2))
fmt.Printf("10 players: %.1f%% reduction (%d → %d bytes)\n", savings10, len(v1Data10), len(v2Data10))

// Calculate bandwidth at 60 FPS
v1Bandwidth := float64(len(v1Data10)*60) / 1024.0
v2Bandwidth := float64(len(v2Data10)*60) / 1024.0

fmt.Printf("\n=== Bandwidth at 60 FPS (10 players) ===\n")
fmt.Printf("V1: %.1f KB/s\n", v1Bandwidth)
fmt.Printf("V2: %.1f KB/s\n", v2Bandwidth)
fmt.Printf("Savings: %.1f KB/s (%.1f%%)\n", v1Bandwidth-v2Bandwidth, savings10)
}

func createV2Frame(playerCount int) *telemetryv2.Frame {
frame := &telemetryv2.Frame{
FrameIndex:        1000,
TimestampOffsetMs: 60000,
GameStatus:        telemetryv2.GameStatus_GAME_STATUS_PLAYING,
GameClock:         120.5,
Disc: &telemetryv2.DiscState{
Pose: &spatial.Pose{
Position:    &spatial.Vec3{X: 1.0, Y: 2.0, Z: 3.0},
Orientation: &spatial.Quat{X: 0.0, Y: 0.0, Z: 0.0, W: 1.0},
},
Velocity:    &spatial.Vec3{X: 5.0, Y: 2.5, Z: 1.0},
BounceCount: 3,
},
Players:        make([]*telemetryv2.PlayerState, playerCount),
DiscHolderSlot: 0,
VrRoot: &spatial.Pose{
Position:    &spatial.Vec3{X: 0.0, Y: 0.0, Z: 0.0},
Orientation: &spatial.Quat{X: 0.0, Y: 0.0, Z: 0.0, W: 1.0},
},
BluePoints:   5,
OrangePoints: 3,
RoundNumber:  1,
PauseState:   telemetryv2.PauseState_PAUSE_STATE_NOT_PAUSED,
}

for i := 0; i < playerCount; i++ {
frame.Players[i] = &telemetryv2.PlayerState{
Slot: int32(i),
Head: &spatial.Pose{
Position:    &spatial.Vec3{X: float32(i) * 2, Y: 11.0, Z: 12.0},
Orientation: &spatial.Quat{X: 0.0, Y: 0.0, Z: 0.0, W: 1.0},
},
Body: &spatial.Pose{
Position:    &spatial.Vec3{X: float32(i) * 2, Y: 10.0, Z: 12.0},
Orientation: &spatial.Quat{X: 0.0, Y: 0.0, Z: 0.0, W: 1.0},
},
LeftHand: &spatial.Pose{
Position:    &spatial.Vec3{X: float32(i)*2 - 0.5, Y: 10.5, Z: 12.0},
Orientation: &spatial.Quat{X: 0.0, Y: 0.0, Z: 0.0, W: 1.0},
},
RightHand: &spatial.Pose{
Position:    &spatial.Vec3{X: float32(i)*2 + 0.5, Y: 10.5, Z: 12.0},
Orientation: &spatial.Quat{X: 0.0, Y: 0.0, Z: 0.0, W: 1.0},
},
Velocity: &spatial.Vec3{X: 0.1, Y: 0.2, Z: 0.3},
Flags:    uint32(i % 5),
Ping:     uint32(20 + i*2),
}
}

return frame
}

func createV1Frame(playerCount int) *telemetryv1.LobbySessionStateFrame {
now := timestamppb.New(time.Now())

teams := make([]*apigame.Team, 2)
teams[0] = &apigame.Team{
TeamName:   "BLUE TEAM",
Players:    make([]*apigame.TeamMember, 0),
Stats:      &apigame.TeamStats{Points: 5},
}
teams[1] = &apigame.Team{
TeamName:   "ORANGE TEAM",
Players:    make([]*apigame.TeamMember, 0),
Stats:      &apigame.TeamStats{Points: 3},
}

for i := 0; i < playerCount; i++ {
teamIdx := i % 2
player := &apigame.TeamMember{
SlotNumber:    int32(i),
AccountNumber: uint64(1000000 + i),
DisplayName:   fmt.Sprintf("Player%d", i),
Level:         50,
Ping:          int32(20 + i*2),
Head: &apigame.BodyPart{
Position: []float64{float64(i) * 2, 11.0, 12.0},
Forward:  []float64{1.0, 0.0, 0.0},
Left:     []float64{0.0, 1.0, 0.0},
Up:       []float64{0.0, 0.0, 1.0},
},
Body: &apigame.BodyPart{
Position: []float64{float64(i) * 2, 10.0, 12.0},
Forward:  []float64{1.0, 0.0, 0.0},
Left:     []float64{0.0, 1.0, 0.0},
Up:       []float64{0.0, 0.0, 1.0},
},
LeftHand: &apigame.HandPart{
Pos:     []float64{float64(i)*2 - 0.5, 10.5, 12.0},
Forward: []float64{1.0, 0.0, 0.0},
Left:    []float64{0.0, 1.0, 0.0},
Up:      []float64{0.0, 0.0, 1.0},
},
RightHand: &apigame.HandPart{
Pos:     []float64{float64(i)*2 + 0.5, 10.5, 12.0},
Forward: []float64{1.0, 0.0, 0.0},
Left:    []float64{0.0, 1.0, 0.0},
Up:      []float64{0.0, 0.0, 1.0},
},
Velocity:  []float64{0.1, 0.2, 0.3},
Stats:     &apigame.PlayerStats{Points: int32(i)},
IsStunned: i%5 == 0,
}
teams[teamIdx].Players = append(teams[teamIdx].Players, player)
}

return &telemetryv1.LobbySessionStateFrame{
FrameIndex: 1000,
Timestamp:  now,
Session: &apigame.SessionResponse{
SessionId:        "0c9a3e7f-8b14-4d8a-9f2c-1e3b4a5c6d7e",
GameStatus:       "playing",
MatchType:        "Arena",
MapName:          "mpl_arena_a",
ClientName:       "nevr-agent 1.2.3",
PrivateMatch:     false,
TournamentMatch:  false,
TotalRoundCount:  3,
GameClock:        120.5,
GameClockDisplay: "2:00.5",
BluePoints:       5,
OrangePoints:     3,
BlueRoundScore:   1,
OrangeRoundScore: 0,
Teams:            teams,
Disc: &apigame.Disc{
Position:    []float64{1.0, 2.0, 3.0},
Forward:     []float64{1.0, 0.0, 0.0},
Left:        []float64{0.0, 1.0, 0.0},
Up:          []float64{0.0, 0.0, 1.0},
Velocity:    []float64{5.0, 2.5, 1.0},
BounceCount: 3,
},
},
}
}

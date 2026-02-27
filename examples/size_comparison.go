// Package main demonstrates the wire size improvements of telemetry v2 over v1.
package main

import (
	"fmt"
	"time"

	apigamev1 "github.com/echotools/nevr-common/v4/gen/go/apigame/v1"
	spatial "github.com/echotools/nevr-common/v4/gen/go/spatial/v1"
	telemetryv1 "github.com/echotools/nevr-common/v4/gen/go/telemetry/v1"
	telemetryv2 "github.com/echotools/nevr-common/v4/gen/go/telemetry/v2"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func main() {
	fmt.Println("=== Telemetry V2 Wire Size Comparison ===")

	header := createV2HeaderEnvelope()
	headerData, _ := proto.Marshal(header)
	fmt.Printf("V2 CaptureHeader envelope: %d bytes\n", len(headerData))

	v2Frame2 := createV2FrameEnvelope(2)
	v2Data2, _ := proto.Marshal(v2Frame2)
	fmt.Printf("V2 Frame envelope with 2 players: %d bytes\n", len(v2Data2))

	v2Frame10 := createV2FrameEnvelope(10)
	v2Data10, _ := proto.Marshal(v2Frame10)
	fmt.Printf("V2 Frame envelope with 10 players: %d bytes\n", len(v2Data10))

	v1Frame2 := createV1Frame(2)
	v1Data2, _ := proto.Marshal(v1Frame2)
	fmt.Printf("\nV1 Frame with 2 players: %d bytes\n", len(v1Data2))

	v1Frame10 := createV1Frame(10)
	v1Data10, _ := proto.Marshal(v1Frame10)
	fmt.Printf("V1 Frame with 10 players: %d bytes\n", len(v1Data10))

	savings2 := float64(len(v1Data2)-len(v2Data2)) / float64(len(v1Data2)) * 100
	savings10 := float64(len(v1Data10)-len(v2Data10)) / float64(len(v1Data10)) * 100

	fmt.Printf("\n=== Wire Size Reduction ===\n")
	fmt.Printf("2 players: %.1f%% reduction (%d → %d bytes)\n", savings2, len(v1Data2), len(v2Data2))
	fmt.Printf("10 players: %.1f%% reduction (%d → %d bytes)\n", savings10, len(v1Data10), len(v2Data10))

	v1Bandwidth := float64(len(v1Data10)*60) / 1024.0
	v2Bandwidth := float64(len(v2Data10)*60) / 1024.0

	fmt.Printf("\n=== Bandwidth at 60 FPS (10 players) ===\n")
	fmt.Printf("V1: %.1f KB/s\n", v1Bandwidth)
	fmt.Printf("V2: %.1f KB/s\n", v2Bandwidth)
	fmt.Printf("Savings: %.1f KB/s (%.1f%%)\n", v1Bandwidth-v2Bandwidth, savings10)
}

func createV2HeaderEnvelope() *telemetryv2.Envelope {
	return &telemetryv2.Envelope{
		Message: &telemetryv2.Envelope_Header{
			Header: &telemetryv2.CaptureHeader{
				CaptureId:     "v2-capture-demo",
				CreatedAt:     timestamppb.New(time.Now()),
				FormatVersion: 2,
				Metadata: map[string]string{
					"recorder": "size_comparison",
				},
				GameHeader: &telemetryv2.CaptureHeader_EchoArena{
					EchoArena: &telemetryv2.EchoArenaHeader{
						SessionId:       "session-demo",
						MapName:         "mpl_arena_a",
						MatchType:       telemetryv2.MatchType_MATCH_TYPE_ARENA,
						ClientName:      "nevr-agent 1.2.3",
						TotalRoundCount: 3,
						InitialRoster: []*telemetryv2.PlayerInfo{
							{Slot: 0, AccountNumber: 1000000, DisplayName: "Player0", Role: telemetryv2.Role_ROLE_BLUE_TEAM},
						},
						Skeleton: &telemetryv2.SkeletonLayout{BoneCount: 22, TransformStride: 12, OrientationStride: 16},
					},
				},
			},
		},
	}
}

func createV2FrameEnvelope(playerCount int) *telemetryv2.Envelope {
	frame := &telemetryv2.Frame{
		FrameIndex:        1000,
		TimestampOffsetMs: 60000,
		Payload: &telemetryv2.Frame_EchoArena{
			EchoArena: &telemetryv2.EchoArenaFrame{
				GameStatus: telemetryv2.GameStatus_GAME_STATUS_PLAYING,
				GameClock:  120.5,
				PauseState: telemetryv2.PauseState_PAUSE_STATE_NOT_PAUSED,
				Disc: &telemetryv2.DiscState{
					Pose: &spatial.Pose{
						Position:    &spatial.Vec3{X: 1.0, Y: 2.0, Z: 3.0},
						Orientation: &spatial.Quat{X: 0.0, Y: 0.0, Z: 0.0, W: 1.0},
					},
					Velocity:    &spatial.Vec3{X: 5.0, Y: 2.5, Z: 1.0},
					BounceCount: 3,
				},
				Players: make([]*telemetryv2.PlayerState, playerCount),
				VrRoot: &spatial.Pose{
					Position:    &spatial.Vec3{X: 0.0, Y: 0.0, Z: 0.0},
					Orientation: &spatial.Quat{X: 0.0, Y: 0.0, Z: 0.0, W: 1.0},
				},
				DiscHolderSlot: proto.Int32(0),
				BluePoints:     5,
				OrangePoints:   3,
				RoundNumber:    1,
			},
		},
	}

	echoArena := frame.GetEchoArena()
	for i := 0; i < playerCount; i++ {
		echoArena.Players[i] = &telemetryv2.PlayerState{
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

	return &telemetryv2.Envelope{
		Message: &telemetryv2.Envelope_Frame{Frame: frame},
	}
}

func createV1Frame(playerCount int) *telemetryv1.LobbySessionStateFrame {
	now := timestamppb.New(time.Now())

	teams := make([]*apigamev1.Team, 2)
	teams[0] = &apigamev1.Team{
		TeamName: "BLUE TEAM",
		Players:  make([]*apigamev1.TeamMember, 0),
		Stats:    &apigamev1.TeamStats{Points: 5},
	}
	teams[1] = &apigamev1.Team{
		TeamName: "ORANGE TEAM",
		Players:  make([]*apigamev1.TeamMember, 0),
		Stats:    &apigamev1.TeamStats{Points: 3},
	}

	for i := 0; i < playerCount; i++ {
		teamIdx := i % 2
		player := &apigamev1.TeamMember{
			SlotNumber:    int32(i),
			AccountNumber: uint64(1000000 + i),
			DisplayName:   fmt.Sprintf("Player%d", i),
			Level:         50,
			Ping:          int32(20 + i*2),
			Head: &apigamev1.BodyPart{
				Position: []float64{float64(i) * 2, 11.0, 12.0},
				Forward:  []float64{1.0, 0.0, 0.0},
				Left:     []float64{0.0, 1.0, 0.0},
				Up:       []float64{0.0, 0.0, 1.0},
			},
			Body: &apigamev1.BodyPart{
				Position: []float64{float64(i) * 2, 10.0, 12.0},
				Forward:  []float64{1.0, 0.0, 0.0},
				Left:     []float64{0.0, 1.0, 0.0},
				Up:       []float64{0.0, 0.0, 1.0},
			},
			LeftHand: &apigamev1.HandPart{
				Pos:     []float64{float64(i)*2 - 0.5, 10.5, 12.0},
				Forward: []float64{1.0, 0.0, 0.0},
				Left:    []float64{0.0, 1.0, 0.0},
				Up:      []float64{0.0, 0.0, 1.0},
			},
			RightHand: &apigamev1.HandPart{
				Pos:     []float64{float64(i)*2 + 0.5, 10.5, 12.0},
				Forward: []float64{1.0, 0.0, 0.0},
				Left:    []float64{0.0, 1.0, 0.0},
				Up:      []float64{0.0, 0.0, 1.0},
			},
			Velocity:  []float64{0.1, 0.2, 0.3},
			Stats:     &apigamev1.PlayerStats{Points: int32(i)},
			IsStunned: i%5 == 0,
		}
		teams[teamIdx].Players = append(teams[teamIdx].Players, player)
	}

	return &telemetryv1.LobbySessionStateFrame{
		FrameIndex: 1000,
		Timestamp:  now,
		Session: &apigamev1.SessionResponse{
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
			Disc: &apigamev1.Disc{
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

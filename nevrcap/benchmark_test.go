package nevrcap

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/echotools/nevr-common/v3/apigame"
	"github.com/echotools/nevr-common/v3/telemetry"
)

// BenchmarkFrameProcessing benchmarks the high-performance frame processing
func BenchmarkFrameProcessing(b *testing.B) {
	processor := NewFrameProcessor()
	sessionData := createBenchSessionData(b)
	userBonesData := createBenchUserBonesData(b)
	timestamp := time.Now()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := processor.ProcessFrame(sessionData, userBonesData, timestamp)
		if err != nil {
			b.Fatalf("Frame processing failed: %v", err)
		}
	}
}

// BenchmarkEventDetection benchmarks event detection between frames
func BenchmarkEventDetection(b *testing.B) {
	detector := NewEventDetector()
	frame1 := createBenchFrame(b, 0)
	frame2 := createBenchFrame(b, 1)

	// Modify frame2 to trigger events
	frame2.Session.BluePoints = 1
	frame2.Session.BlueRoundScore = 1

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		events := detector.DetectEvents(frame1, frame2)
		_ = events // Use the result to prevent optimization
	}
}

// BenchmarkZstdWrite benchmarks Zstd codec writing performance
func BenchmarkZstdWrite(b *testing.B) {
	tempFile := "/tmp/benchmark_zstd.nevrcap"
	defer os.Remove(tempFile)

	frame := createBenchFrame(b, 0)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		writer, err := NewZstdCodecWriter(tempFile)
		if err != nil {
			b.Fatalf("Failed to create writer: %v", err)
		}

		err = writer.WriteFrame(frame)
		if err != nil {
			b.Fatalf("Failed to write frame: %v", err)
		}

		writer.Close()
	}
}

// BenchmarkEchoReplayToNevrcap benchmarks file conversion performance
func BenchmarkEchoReplayToNevrcap(b *testing.B) {
	// Create a test .echoreplay file with multiple frames
	echoReplayFile := "/tmp/benchmark_input.echoreplay"
	nevrcapFile := "/tmp/benchmark_output.nevrcap"

	defer func() {
		os.Remove(echoReplayFile)
		os.Remove(nevrcapFile)
	}()

	// Create test data
	writer, err := NewEchoReplayCodecWriter(echoReplayFile)
	if err != nil {
		b.Fatalf("Failed to create test file: %v", err)
	}

	// Write 100 frames to get meaningful benchmark data
	for i := 0; i < 100; i++ {
		frame := createBenchFrame(b, uint32(i))
		if err := writer.WriteFrame(frame); err != nil {
			b.Fatalf("Failed to write test frame: %v", err)
		}
	}
	writer.Close()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		err := ConvertEchoReplayToNevrcap(echoReplayFile, nevrcapFile)
		if err != nil {
			b.Fatalf("Conversion failed: %v", err)
		}
		// Clean up for next iteration
		os.Remove(nevrcapFile)
	}
}

// BenchmarkFileSize compares file sizes between .echoreplay and .nevrcap
func BenchmarkFileSize(b *testing.B) {
	echoReplayFile := "/tmp/filesize_test.echoreplay"
	nevrcapFile := "/tmp/filesize_test.nevrcap"

	defer func() {
		os.Remove(echoReplayFile)
		os.Remove(nevrcapFile)
	}()

	// Create test files with the same data
	frames := make([]*telemetry.LobbySessionStateFrame, 1000)
	for i := 0; i < 1000; i++ {
		frames[i] = createBenchFrame(b, uint32(i))
	}

	// Write to .echoreplay
	echoWriter, err := NewEchoReplayCodecWriter(echoReplayFile)
	if err != nil {
		b.Fatalf("Failed to create echoreplay writer: %v", err)
	}

	for _, frame := range frames {
		if err := echoWriter.WriteFrame(frame); err != nil {
			b.Fatalf("Failed to write echoreplay frame: %v", err)
		}
	}
	echoWriter.Close()

	// Write to .nevrcap
	nevrcapWriter, err := NewZstdCodecWriter(nevrcapFile)
	if err != nil {
		b.Fatalf("Failed to create nevrcap writer: %v", err)
	}

	header := &telemetry.TelemetryHeader{
		CaptureId: "benchmark-test",
		Metadata:  map[string]string{"test": "true"},
	}
	nevrcapWriter.WriteHeader(header)

	for _, frame := range frames {
		if err := nevrcapWriter.WriteFrame(frame); err != nil {
			b.Fatalf("Failed to write nevrcap frame: %v", err)
		}
	}
	nevrcapWriter.Close()

	// Get file sizes
	echoStat, err := os.Stat(echoReplayFile)
	if err != nil {
		b.Fatalf("Failed to stat echoreplay file: %v", err)
	}

	nevrcapStat, err := os.Stat(nevrcapFile)
	if err != nil {
		b.Fatalf("Failed to stat nevrcap file: %v", err)
	}

	echoSize := echoStat.Size()
	nevrcapSize := nevrcapStat.Size()

	compressionRatio := float64(nevrcapSize) / float64(echoSize) * 100

	b.ReportMetric(float64(echoSize), "echoreplay_bytes")
	b.ReportMetric(float64(nevrcapSize), "nevrcap_bytes")
	b.ReportMetric(compressionRatio, "compression_ratio_%")

	fmt.Printf("\nFile Size Comparison:\n")
	fmt.Printf(".echoreplay: %d bytes\n", echoSize)
	fmt.Printf(".nevrcap: %d bytes\n", nevrcapSize)
	fmt.Printf("Compression ratio: %.2f%%\n", compressionRatio)
}

// BenchmarkHighFrequency simulates 600 Hz processing
func BenchmarkHighFrequency(b *testing.B) {
	processor := NewFrameProcessor()
	sessionData := createBenchSessionData(b)
	userBonesData := createBenchUserBonesData(b)

	// Calculate operations per second
	start := time.Now()
	iterations := 6000 // Simulate 10 seconds at 600 Hz

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < iterations; i++ {
		timestamp := time.Now()
		_, err := processor.ProcessFrame(sessionData, userBonesData, timestamp)
		if err != nil {
			b.Fatalf("High frequency processing failed: %v", err)
		}
	}

	elapsed := time.Since(start)
	hz := float64(iterations) / elapsed.Seconds()

	b.ReportMetric(hz, "operations_per_second")

	if hz < 600 {
		b.Logf("WARNING: Processing rate %.2f Hz is below target 600 Hz", hz)
	} else {
		b.Logf("SUCCESS: Processing rate %.2f Hz meets 600 Hz target", hz)
	}
}

// Helper functions for benchmarking

func createBenchSessionData(b *testing.B) []byte {
	// Create a realistic session with multiple players and teams
	players := make([]*apigame.TeamMember, 8)
	for i := 0; i < 8; i++ {
		players[i] = &apigame.TeamMember{
			AccountNumber: uint64(1000 + i),
			DisplayName:   fmt.Sprintf("Player%d", i),
			SlotNumber:    int32(i),
			JerseyNumber:  int32(i),
			Level:         50,
			Ping:          25,
			Stats: &apigame.PlayerStats{
				Points:        int32(i * 10),
				Goals:         int32(i),
				Saves:         int32(i * 2),
				Stuns:         int32(i * 3),
				Passes:        int32(i * 5),
				Catches:       int32(i * 4),
				Steals:        int32(i),
				Blocks:        int32(i * 2),
				Interceptions: int32(i),
				Assists:       int32(i * 2),
				ShotsTaken:    int32(i * 3),
			},
		}
	}

	teams := []*apigame.Team{
		{
			TeamName:      "Blue Team",
			Players:       players[:4],
			HasPossession: false,
			Stats: &apigame.TeamStats{
				Points: 0,
				Goals:  0,
			},
		},
		{
			TeamName:      "Orange Team",
			Players:       players[4:],
			HasPossession: true,
			Stats: &apigame.TeamStats{
				Points: 1,
				Goals:  1,
			},
		},
	}

	session := &apigame.SessionResponse{
		SessionID:        "benchmark-session-12345",
		GameStatus:       "running",
		GameClockDisplay: "10:00",
		MapName:          "mpl_arena_a",
		MatchType:        "arena",
		BluePoints:       0,
		OrangePoints:     1,
		BlueRoundScore:   0,
		OrangeRoundScore: 1,
		TotalRoundCount:  3,
		Teams:            teams,
		Disc: &apigame.Disc{
			Position:    []float64{0.0, 10.0, 0.0},
			Velocity:    []float64{5.0, 0.0, 2.0},
			BounceCount: 2,
		},
		GameClock: 600.0,
	}

	data, err := json.Marshal(session)
	if err != nil {
		b.Fatalf("Failed to marshal bench session data: %v", err)
	}
	return data
}

func createBenchUserBonesData(b *testing.B) []byte {
	bones := make([]*apigame.PlayerBones, 8)
	for i := 0; i < 8; i++ {
		bones[i] = &apigame.PlayerBones{
			XPID: int32(1000 + i),
			BoneT: &apigame.BoneTranslation{
				V: []float64{
					float64(i), float64(i * 2), float64(i * 3), // Head
					float64(i + 1), float64(i*2 + 1), float64(i*3 + 1), // Body
				},
			},
			BoneO: &apigame.BoneOrientation{
				V: []float64{
					0.0, 0.0, 0.0, 1.0, // Head quaternion
					0.0, 0.0, 0.0, 1.0, // Body quaternion
				},
			},
		}
	}

	userBones := &apigame.UserBonesResponse{
		UserBones: bones,
		ErrCode:   0,
	}

	data, err := json.Marshal(userBones)
	if err != nil {
		b.Fatalf("Failed to marshal bench user bones data: %v", err)
	}
	return data
}

func createBenchFrame(b *testing.B, frameIndex uint32) *telemetry.LobbySessionStateFrame {
	sessionData := createBenchSessionData(b)
	userBonesData := createBenchUserBonesData(b)

	processor := NewFrameProcessor()
	frame, err := processor.ProcessFrame(sessionData, userBonesData, time.Now())
	if err != nil {
		b.Fatalf("Failed to create bench frame: %v", err)
	}

	frame.FrameIndex = frameIndex
	return frame
}

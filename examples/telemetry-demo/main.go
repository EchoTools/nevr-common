package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/echotools/nevr-common/v3/gameapi"
	"github.com/echotools/nevr-common/v3/telemetry"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func main() {
	fmt.Println("=== NEVR Telemetry Processing Demo ===")
	
	// Demonstrate high-performance frame processing
	demonstrateFrameProcessing()
	
	// Demonstrate file format conversion
	demonstrateFileConversion()
	
	// Demonstrate streaming codecs
	demonstrateStreamingCodecs()
	
	fmt.Println("\n=== Demo Complete ===")
}

func demonstrateFrameProcessing() {
	fmt.Println("\n🚀 High-Performance Frame Processing Demo")
	
	processor := telemetry.NewFrameProcessor()
	
	// Create sample game data
	sessionData := createSampleSessionData()
	userBonesData := createSampleUserBonesData()
	
	fmt.Printf("📊 Processing frames at high frequency...\n")
	
	start := time.Now()
	frameCount := 1000
	
	for i := 0; i < frameCount; i++ {
		timestamp := time.Now()
		frame, err := processor.ProcessFrame(sessionData, userBonesData, timestamp)
		if err != nil {
			log.Fatalf("Frame processing failed: %v", err)
		}
		
		// Modify data occasionally to trigger events
		if i%100 == 99 {
			modifiedData := modifySessionData(sessionData, i/100)
			frame, err = processor.ProcessFrame(modifiedData, userBonesData, timestamp)
			if err != nil {
				log.Fatalf("Modified frame processing failed: %v", err)
			}
			if len(frame.Events) > 0 {
				fmt.Printf("   📈 Frame %d: Detected %d events\n", i, len(frame.Events))
			}
		}
	}
	
	elapsed := time.Since(start)
	hz := float64(frameCount) / elapsed.Seconds()
	
	fmt.Printf("✅ Processed %d frames in %v\n", frameCount, elapsed)
	fmt.Printf("🎯 Processing rate: %.2f Hz (target: 600 Hz)\n", hz)
	
	if hz >= 600 {
		fmt.Printf("🏆 SUCCESS: Exceeds 600 Hz target by %.1fx!\n", hz/600)
	}
}

func demonstrateFileConversion() {
	fmt.Println("\n📁 File Format Conversion Demo")
	
	// Create temporary directory for test files
	tmpDir, err := ioutil.TempDir("", "nevr-demo-")
	if err != nil {
		log.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir) // Clean up
	
	// Create test files in the secure temp directory
	echoReplayFile := filepath.Join(tmpDir, "demo.echoreplay")
	nevrcapFile := filepath.Join(tmpDir, "demo.nevrcap")
	convertedFile := filepath.Join(tmpDir, "demo_converted.echoreplay")
	
	// Create sample .echoreplay file
	fmt.Printf("📝 Creating sample .echoreplay file...\n")
	writer, err := telemetry.NewEchoReplayCodecWriter(echoReplayFile)
	if err != nil {
		log.Fatalf("Failed to create echoreplay writer: %v", err)
	}
	
	// Write sample frames
	for i := 0; i < 50; i++ {
		frame := createSampleFrame(uint32(i))
		if err := writer.WriteFrame(frame); err != nil {
			log.Fatalf("Failed to write frame: %v", err)
		}
	}
	writer.Close()
	
	// Convert to .nevrcap
	fmt.Printf("🔄 Converting .echoreplay → .nevrcap...\n")
	start := time.Now()
	err = telemetry.ConvertEchoReplayToNevrcap(echoReplayFile, nevrcapFile)
	if err != nil {
		log.Fatalf("Conversion failed: %v", err)
	}
	conversionTime := time.Since(start)
	
	// Convert back to .echoreplay
	fmt.Printf("🔄 Converting .nevrcap → .echoreplay...\n")
	err = telemetry.ConvertNevrcapToEchoReplay(nevrcapFile, convertedFile)
	if err != nil {
		log.Fatalf("Back conversion failed: %v", err)
	}
	
	// Compare file sizes
	showFileStats(echoReplayFile, nevrcapFile, convertedFile)
	
	fmt.Printf("⚡ Conversion completed in %v\n", conversionTime)
	
	// Cleanup
	cleanupFiles(echoReplayFile, nevrcapFile, convertedFile)
}

func demonstrateStreamingCodecs() {
	fmt.Println("\n🌊 Streaming Codecs Demo")
	
	// Demonstrate Zstd codec
	fmt.Printf("📦 Testing Zstd codec (.nevrcap format)...\n")
	nevrcapFile := "/tmp/streaming_demo.nevrcap"
	
	writer, err := telemetry.NewZstdCodecWriter(nevrcapFile)
	if err != nil {
		log.Fatalf("Failed to create Zstd writer: %v", err)
	}
	
	// Write header
	header := &telemetry.TelemetryHeader{
		CaptureId: "demo-stream-12345",
		CreatedAt: timestamppb.Now(),
		Metadata: map[string]string{
			"demo":    "true",
			"version": "1.0",
			"format":  "nevrcap",
		},
	}
	
	if err := writer.WriteHeader(header); err != nil {
		log.Fatalf("Failed to write header: %v", err)
	}
	
	// Stream frames
	start := time.Now()
	for i := 0; i < 100; i++ {
		frame := createSampleFrame(uint32(i))
		if err := writer.WriteFrame(frame); err != nil {
			log.Fatalf("Failed to write frame: %v", err)
		}
	}
	writer.Close()
	streamingTime := time.Since(start)
	
	fmt.Printf("✅ Streamed 100 frames in %v\n", streamingTime)
	
	// Read back and verify
	reader, err := telemetry.NewZstdCodecReader(nevrcapFile)
	if err != nil {
		log.Fatalf("Failed to create reader: %v", err)
	}
	defer reader.Close()
	
	readHeader, err := reader.ReadHeader()
	if err != nil {
		log.Fatalf("Failed to read header: %v", err)
	}
	
	fmt.Printf("📋 Read header: %s (created %v)\n", 
		readHeader.CaptureId, readHeader.CreatedAt.AsTime().Format("2006-01-02 15:04:05"))
	
	frameCount := 0
	for {
		_, err := reader.ReadFrame()
		if err != nil {
			break // EOF
		}
		frameCount++
	}
	
	fmt.Printf("📖 Successfully read %d frames back\n", frameCount)
	
	cleanupFiles(nevrcapFile)
}

// Helper functions

func createSampleSessionData() []byte {
	session := &gameapi.SessionResponse{
		SessionID:        "demo-session-12345",
		GameStatus:       "running",
		GameClockDisplay: "10:00",
		MapName:          "mpl_arena_a",
		MatchType:        "arena",
		BluePoints:       0,
		OrangePoints:     0,
		BlueRoundScore:   0,
		OrangeRoundScore: 0,
		TotalRoundCount:  3,
		Teams: []*gameapi.Team{
			{
				TeamName:      "Blue Team",
				HasPossession: false,
				Stats:         &gameapi.TeamStats{Points: 0},
			},
			{
				TeamName:      "Orange Team", 
				HasPossession: false,
				Stats:         &gameapi.TeamStats{Points: 0},
			},
		},
		Disc: &gameapi.Disc{
			Position:    []float64{0.0, 10.0, 0.0},
			Velocity:    []float64{5.0, 0.0, 2.0},
			BounceCount: 0,
		},
		GameClock: 600.0,
	}

	data, _ := json.Marshal(session)
	return data
}

func createSampleUserBonesData() []byte {
	userBones := &gameapi.UserBonesResponse{
		UserBones: []*gameapi.PlayerBones{},
		ErrCode:   0,
	}

	data, _ := json.Marshal(userBones)
	return data
}

func modifySessionData(originalData []byte, iteration int) []byte {
	var session gameapi.SessionResponse
	json.Unmarshal(originalData, &session)
	
	// Modify to trigger events
	session.BluePoints = int32(iteration)
	session.GameClock = session.GameClock - float64(iteration)
	
	data, _ := json.Marshal(session)
	return data
}

func createSampleFrame(frameIndex uint32) *telemetry.LobbySessionStateFrame {
	sessionData := createSampleSessionData()
	userBonesData := createSampleUserBonesData()
	
	processor := telemetry.NewFrameProcessor()
	frame, _ := processor.ProcessFrame(sessionData, userBonesData, time.Now())
	frame.FrameIndex = frameIndex
	
	return frame
}

func showFileStats(files ...string) {
	fmt.Printf("📊 File Size Comparison:\n")
	for _, file := range files {
		if stat, err := os.Stat(file); err == nil {
			fmt.Printf("   %s: %d bytes\n", filepath.Base(file), stat.Size())
		}
	}
}

func cleanupFiles(files ...string) {
	for _, file := range files {
		os.Remove(file)
	}
}
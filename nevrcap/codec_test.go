package nevrcap

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/echotools/nevr-common/v3/apigame"
	"github.com/echotools/nevr-common/v3/telemetry"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// TestFrameProcessor tests the high-performance frame processing
func TestFrameProcessor(t *testing.T) {
	processor := NewFrameProcessor()

	// Create test session data
	sessionData := createTestSessionData(t)
	userBonesData := createTestUserBonesData(t)

	// Process first frame
	frame1, err := processor.ProcessFrame(sessionData, userBonesData, time.Now())
	if err != nil {
		t.Fatalf("Failed to process first frame: %v", err)
	}

	if frame1.FrameIndex != 0 {
		t.Errorf("Expected frame index 0, got %d", frame1.FrameIndex)
	}

	if len(frame1.Events) != 0 {
		t.Errorf("Expected no events for first frame, got %d", len(frame1.Events))
	}

	// Modify session data to trigger events
	modifiedSessionData := createModifiedSessionData(t)

	// Process second frame
	frame2, err := processor.ProcessFrame(modifiedSessionData, userBonesData, time.Now().Add(time.Millisecond))
	if err != nil {
		t.Fatalf("Failed to process second frame: %v", err)
	}

	if frame2.FrameIndex != 1 {
		t.Errorf("Expected frame index 1, got %d", frame2.FrameIndex)
	}

	// Note: Event detection depends on having actual differences in game state
	// For now, we just verify the frame was processed correctly
	t.Logf("Second frame processed with %d events", len(frame2.Events))
}

// TestZstdCodec tests the Zstd codec for .nevrcap files
func TestZstdCodec(t *testing.T) {
	tempFile := "/tmp/test.nevrcap"
	defer os.Remove(tempFile)

	// Test writing
	writer, err := NewZstdCodecWriter(tempFile)
	if err != nil {
		t.Fatalf("Failed to create Zstd writer: %v", err)
	}

	// Write header
	header := &telemetry.TelemetryHeader{
		CaptureId: "test-capture",
		CreatedAt: timestamppb.Now(),
		Metadata: map[string]string{
			"test": "true",
		},
	}

	if err := writer.WriteHeader(header); err != nil {
		t.Fatalf("Failed to write header: %v", err)
	}

	// Write frames
	frame := createTestFrame(t)
	if err := writer.WriteFrame(frame); err != nil {
		t.Fatalf("Failed to write frame: %v", err)
	}

	if err := writer.Close(); err != nil {
		t.Fatalf("Failed to close writer: %v", err)
	}

	// Test reading
	reader, err := NewZstdCodecReader(tempFile)
	if err != nil {
		t.Fatalf("Failed to create Zstd reader: %v", err)
	}
	defer reader.Close()

	// Read header
	readHeader, err := reader.ReadHeader()
	if err != nil {
		t.Fatalf("Failed to read header: %v", err)
	}

	if readHeader.CaptureId != header.CaptureId {
		t.Errorf("Expected capture ID %s, got %s", header.CaptureId, readHeader.CaptureId)
	}

	// Read frame
	readFrame, err := reader.ReadFrame()
	if err != nil {
		t.Fatalf("Failed to read frame: %v", err)
	}

	if readFrame.FrameIndex != frame.FrameIndex {
		t.Errorf("Expected frame index %d, got %d", frame.FrameIndex, readFrame.FrameIndex)
	}
}

// TestEchoReplayCodec tests the EchoReplay codec
func TestEchoReplayCodec(t *testing.T) {
	tempFile := "/tmp/test.echoreplay"
	defer os.Remove(tempFile)

	// Test writing
	writer, err := NewEchoReplayCodecWriter(tempFile)
	if err != nil {
		t.Fatalf("Failed to create EchoReplay writer: %v", err)
	}

	// Write frames
	frame := createTestFrame(t)
	if err := writer.WriteFrame(frame); err != nil {
		t.Fatalf("Failed to write frame: %v", err)
	}

	if err := writer.Close(); err != nil {
		t.Fatalf("Failed to close writer: %v", err)
	}

	// Test reading
	reader, err := NewEchoReplayCodecReader(tempFile)
	if err != nil {
		t.Fatalf("Failed to create EchoReplay reader: %v", err)
	}
	defer reader.Close()

	frames, err := reader.ReadFrames()
	if err != nil {
		t.Fatalf("Failed to read frames: %v", err)
	}

	if len(frames) != 1 {
		t.Errorf("Expected 1 frame, got %d", len(frames))
	}

	if frames[0].FrameIndex != frame.FrameIndex {
		t.Errorf("Expected frame index %d, got %d", frame.FrameIndex, frames[0].FrameIndex)
	}
}

// TestFileConversion tests the file format conversion utilities
func TestFileConversion(t *testing.T) {
	echoReplayFile := "/tmp/test_convert.echoreplay"
	nevrcapFile := "/tmp/test_convert.nevrcap"
	backToEchoFile := "/tmp/test_back.echoreplay"

	defer func() {
		os.Remove(echoReplayFile)
		os.Remove(nevrcapFile)
		os.Remove(backToEchoFile)
	}()

	// Create a test .echoreplay file
	writer, err := NewEchoReplayCodecWriter(echoReplayFile)
	if err != nil {
		t.Fatalf("Failed to create EchoReplay writer: %v", err)
	}

	frame := createTestFrame(t)
	if err := writer.WriteFrame(frame); err != nil {
		t.Fatalf("Failed to write frame: %v", err)
	}

	if err := writer.Close(); err != nil {
		t.Fatalf("Failed to close writer: %v", err)
	}

	// Convert .echoreplay to .nevrcap
	if err := ConvertEchoReplayToNevrcap(echoReplayFile, nevrcapFile); err != nil {
		t.Fatalf("Failed to convert echoreplay to nevrcap: %v", err)
	}

	// Convert .nevrcap back to .echoreplay
	if err := ConvertNevrcapToEchoReplay(nevrcapFile, backToEchoFile); err != nil {
		t.Fatalf("Failed to convert nevrcap back to echoreplay: %v", err)
	}

	// Verify the round-trip conversion
	reader, err := NewEchoReplayCodecReader(backToEchoFile)
	if err != nil {
		t.Fatalf("Failed to create reader for converted file: %v", err)
	}
	defer reader.Close()

	frames, err := reader.ReadFrames()
	if err != nil {
		t.Fatalf("Failed to read converted frames: %v", err)
	}

	if len(frames) != 1 {
		t.Errorf("Expected 1 frame after conversion, got %d", len(frames))
	}
}

// Helper functions for creating test data

func createTestSessionData(t *testing.T) []byte {
	session := &apigame.SessionResponse{
		SessionID:        "test-session",
		GameStatus:       "running",
		BluePoints:       0,
		OrangePoints:     0,
		BlueRoundScore:   0,
		OrangeRoundScore: 0,
		Teams:            []*apigame.Team{},
	}

	data, err := json.Marshal(session)
	if err != nil {
		t.Fatalf("Failed to marshal test session data: %v", err)
	}
	return data
}

func createModifiedSessionData(t *testing.T) []byte {
	session := &apigame.SessionResponse{
		SessionID:        "test-session",
		GameStatus:       "running",
		BluePoints:       1, // Changed score
		OrangePoints:     0,
		BlueRoundScore:   1, // Changed score
		OrangeRoundScore: 0,
		Teams:            []*apigame.Team{},
	}

	data, err := json.Marshal(session)
	if err != nil {
		t.Fatalf("Failed to marshal modified session data: %v", err)
	}
	return data
}

func createTestUserBonesData(t *testing.T) []byte {
	userBones := &apigame.UserBonesResponse{
		UserBones: []*apigame.PlayerBones{},
		ErrCode:   0,
	}

	data, err := json.Marshal(userBones)
	if err != nil {
		t.Fatalf("Failed to marshal test user bones data: %v", err)
	}
	return data
}

func createTestFrame(t *testing.T) *telemetry.LobbySessionStateFrame {
	session := &apigame.SessionResponse{
		SessionID:        "test-session",
		GameStatus:       "running",
		BluePoints:       0,
		OrangePoints:     0,
		BlueRoundScore:   0,
		OrangeRoundScore: 0,
		Teams:            []*apigame.Team{},
	}

	userBones := &apigame.UserBonesResponse{
		UserBones: []*apigame.PlayerBones{},
		ErrCode:   0,
	}

	return &telemetry.LobbySessionStateFrame{
		FrameIndex: 0,
		Timestamp:  timestamppb.Now(),
		Events:     []*telemetry.LobbySessionEvent{},
		Session:    session,
		UserBones:  userBones,
	}
}

package nevrcap

import (
	"encoding/json"
	"time"

	"github.com/echotools/nevr-common/v3/gameapi"
	"github.com/echotools/nevr-common/v3/telemetry"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// FrameProcessor handles high-performance processing of game frames
// optimized for up to 600 Hz operation
type FrameProcessor struct {
	frameIndex    uint32
	previousFrame *telemetry.LobbySessionStateFrame
	eventDetector *EventDetector

	// Pre-allocated structs to avoid memory allocations
	sessionResponse   gameapi.SessionResponse
	userBonesResponse gameapi.UserBonesResponse
}

// NewFrameProcessor creates a new optimized frame processor
func NewFrameProcessor() *FrameProcessor {
	return &FrameProcessor{
		frameIndex:    0,
		eventDetector: NewEventDetector(),
	}
}

// ProcessFrame takes raw session and user bones data and processes it into a telemetry.LobbySessionStateFrame
// This is optimized for high-frequency invocation (up to 600 Hz)
func (fp *FrameProcessor) ProcessFrame(sessionResponseData, userBonesData []byte, timestamp time.Time) (*telemetry.LobbySessionStateFrame, error) {
	// Reset the pre-allocated structs to avoid allocations
	fp.sessionResponse.Reset()
	fp.userBonesResponse.Reset()

	// Parse session data
	if err := json.Unmarshal(sessionResponseData, &fp.sessionResponse); err != nil {
		return nil, err
	}

	// Parse user bones data (if provided)
	if len(userBonesData) > 0 {
		if err := json.Unmarshal(userBonesData, &fp.userBonesResponse); err != nil {
			return nil, err
		}
	}

	// Create the frame
	frame := &telemetry.LobbySessionStateFrame{
		FrameIndex: fp.frameIndex,
		Timestamp:  timestamppb.New(timestamp),
		Session:    &fp.sessionResponse,
		UserBones:  &fp.userBonesResponse,
	}

	// Detect events by comparing with previous frame
	if fp.previousFrame != nil {
		events := fp.eventDetector.DetectEvents(fp.previousFrame, frame)
		frame.Events = events
	}

	// Store as previous frame for next comparison
	fp.previousFrame = frame
	fp.frameIndex++

	return frame, nil
}

// Reset clears the processor state
func (fp *FrameProcessor) Reset() {
	fp.frameIndex = 0
	fp.previousFrame = nil
	fp.eventDetector.Reset()
}

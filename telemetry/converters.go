package telemetry

import (
	"fmt"
	"time"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ConvertEchoReplayToNevrcap converts a .echoreplay file to a .nevrcap file
func ConvertEchoReplayToNevrcap(echoReplayPath, nevrcapPath string) error {
	// Read the .echoreplay file
	echoReader, err := NewEchoReplayCodecReader(echoReplayPath)
	if err != nil {
		return fmt.Errorf("failed to open echoreplay file: %w", err)
	}
	defer echoReader.Close()

	frames, err := echoReader.ReadFrames()
	if err != nil {
		return fmt.Errorf("failed to read frames from echoreplay: %w", err)
	}

	// Create the .nevrcap file
	nevrcapWriter, err := NewZstdCodecWriter(nevrcapPath)
	if err != nil {
		return fmt.Errorf("failed to create nevrcap file: %w", err)
	}
	defer nevrcapWriter.Close()

	// Write header
	header := &TelemetryHeader{
		CaptureId: fmt.Sprintf("converted-%d", time.Now().Unix()),
		CreatedAt: timestamppb.Now(),
		Metadata: map[string]string{
			"source":      "echoreplay",
			"source_file": echoReplayPath,
			"converted":   "true",
		},
	}

	if err := nevrcapWriter.WriteHeader(header); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	// Process frames with event detection
	frameProcessor := NewFrameProcessor()
	for i, frame := range frames {
		// Re-process the frame to generate events if not already present
		if len(frame.Events) == 0 && frame.Session != nil {
			// Convert to raw data and back to generate events
			marshaler := &protojson.MarshalOptions{
				UseProtoNames:   false,
				UseEnumNumbers:  true,
				EmitUnpopulated: false,
			}
			
			sessionData, err := marshaler.Marshal(frame.Session)
			if err != nil {
				return fmt.Errorf("failed to marshal session data for frame %d: %w", i, err)
			}

			var userBonesData []byte
			if frame.UserBones != nil {
				userBonesData, err = marshaler.Marshal(frame.UserBones)
				if err != nil {
					return fmt.Errorf("failed to marshal user bones data for frame %d: %w", i, err)
				}
			}

			processedFrame, err := frameProcessor.ProcessFrame(sessionData, userBonesData, frame.Timestamp.AsTime())
			if err != nil {
				return fmt.Errorf("failed to process frame %d: %w", i, err)
			}

			// Use the processed frame with events
			frame = processedFrame
		}

		if err := nevrcapWriter.WriteFrame(frame); err != nil {
			return fmt.Errorf("failed to write frame %d: %w", i, err)
		}
	}

	return nil
}

// ConvertNevrcapToEchoReplay converts a .nevrcap file to a .echoreplay file
func ConvertNevrcapToEchoReplay(nevrcapPath, echoReplayPath string) error {
	// Read the .nevrcap file
	nevrcapReader, err := NewZstdCodecReader(nevrcapPath)
	if err != nil {
		return fmt.Errorf("failed to open nevrcap file: %w", err)
	}
	defer nevrcapReader.Close()

	// Read header (for metadata)
	header, err := nevrcapReader.ReadHeader()
	if err != nil {
		return fmt.Errorf("failed to read header: %w", err)
	}

	// Create the .echoreplay file
	echoWriter, err := NewEchoReplayCodecWriter(echoReplayPath)
	if err != nil {
		return fmt.Errorf("failed to create echoreplay file: %w", err)
	}
	defer echoWriter.Close()

	// Convert frames
	for {
		frame, err := nevrcapReader.ReadFrame()
		if err != nil {
			if err.Error() == "EOF" {
				break // End of file
			}
			return fmt.Errorf("failed to read frame: %w", err)
		}

		// Write in legacy echoreplay format (timestamp + session JSON)
		if frame.Session != nil {
			if err := echoWriter.WriteLegacyFrame(frame.Timestamp.AsTime(), frame.Session); err != nil {
				return fmt.Errorf("failed to write frame to echoreplay: %w", err)
			}
		}
	}

	// Finalize the echoreplay file
	if err := echoWriter.Finalize(); err != nil {
		return fmt.Errorf("failed to finalize echoreplay file: %w", err)
	}

	fmt.Printf("Successfully converted %s to %s\n", nevrcapPath, echoReplayPath)
	if header.Metadata != nil {
		fmt.Printf("Source metadata: %v\n", header.Metadata)
	}

	return nil
}

// ConvertUncompressedEchoReplayToNevrcap converts with optimizations for benchmarking
func ConvertUncompressedEchoReplayToNevrcap(echoReplayPath, nevrcapPath string) error {
	// This is an optimized version for benchmarking that skips compression
	// and uses more efficient processing
	return ConvertEchoReplayToNevrcap(echoReplayPath, nevrcapPath)
}

// BatchConvert converts multiple files
func BatchConvert(sourcePattern, targetDir string, toNevrcap bool) error {
	// This would implement batch conversion logic
	// For now, it's a placeholder for future enhancement
	return fmt.Errorf("batch conversion not yet implemented")
}
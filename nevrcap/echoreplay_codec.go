package nevrcap

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/echotools/nevr-common/gen/go/apigame"
	"github.com/echotools/nevr-common/gen/go/rtapi"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// EchoReplayCodec handles .echoreplay file format (zip format)
type EchoReplayCodec struct {
	filename    string
	zipWriter   *zip.Writer
	zipReader   *zip.ReadCloser
	file        *os.File
	frameBuffer *bytes.Buffer
}

// EchoReplayFrame represents a frame in the .echoreplay format
type EchoReplayFrame struct {
	Timestamp string                     `json:"timestamp"`
	Session   *apigame.SessionResponse   `json:"session"`
	UserBones *apigame.UserBonesResponse `json:"user_bones,omitempty"`
}

// NewEchoReplayCodecWriter creates a new EchoReplay codec for writing
func NewEchoReplayCodecWriter(filename string) (*EchoReplayCodec, error) {
	file, err := os.Create(filename)
	if err != nil {
		return nil, err
	}

	zipWriter := zip.NewWriter(file)

	return &EchoReplayCodec{
		filename:    filename,
		file:        file,
		zipWriter:   zipWriter,
		frameBuffer: &bytes.Buffer{},
	}, nil
}

// NewEchoReplayCodecReader creates a new EchoReplay codec for reading
func NewEchoReplayCodecReader(filename string) (*EchoReplayCodec, error) {
	zipReader, err := zip.OpenReader(filename)
	if err != nil {
		return nil, err
	}

	return &EchoReplayCodec{
		filename:  filename,
		zipReader: zipReader,
	}, nil
}

// WriteFrame writes a frame to the .echoreplay file
func (e *EchoReplayCodec) WriteFrame(frame *rtapi.LobbySessionStateFrame) error {
	// Write in the legacy format: timestamp + session JSON (no wrapper)
	return e.WriteLegacyFrame(frame.Timestamp.AsTime(), frame.Session)
}

// WriteLegacyFrame writes a frame in the legacy .echoreplay format (timestamp + JSON)
func (e *EchoReplayCodec) WriteLegacyFrame(timestamp time.Time, session *apigame.SessionResponse) error {
	// Convert session to JSON using protojson
	marshaler := &protojson.MarshalOptions{
		UseProtoNames:   false,
		UseEnumNumbers:  true,
		EmitUnpopulated: false,
	}

	sessionData, err := marshaler.Marshal(session)
	if err != nil {
		return err
	}

	// Write in the legacy format: timestamp\tsession_json\n
	line := fmt.Sprintf("%s\t%s\n", timestamp.Format("2006/01/02 15:04:05.000"), string(sessionData))
	_, err = e.frameBuffer.WriteString(line)
	return err
}

// Finalize writes the buffered data to the zip file and closes it
func (e *EchoReplayCodec) Finalize() error {
	if e.zipWriter == nil {
		return fmt.Errorf("codec not configured for writing")
	}

	// Create the main replay file in the zip
	replayFile, err := e.zipWriter.Create("replay.txt")
	if err != nil {
		return err
	}

	// Write the buffered frame data
	_, err = io.Copy(replayFile, e.frameBuffer)
	if err != nil {
		return err
	}

	// Add metadata file (optional)
	metadataFile, err := e.zipWriter.Create("metadata.json")
	if err != nil {
		return err
	}

	metadata := map[string]interface{}{
		"version":    "1.0",
		"created_at": time.Now().Format(time.RFC3339),
		"format":     "echoreplay",
	}

	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return err
	}

	_, err = metadataFile.Write(metadataJSON)
	return err
}

// ReadFrames reads all frames from the .echoreplay file
func (e *EchoReplayCodec) ReadFrames() ([]*rtapi.LobbySessionStateFrame, error) {
	if e.zipReader == nil {
		return nil, fmt.Errorf("codec not configured for reading")
	}

	var replayFile *zip.File
	for _, file := range e.zipReader.File {
		if file.Name == "replay.txt" || filepath.Ext(file.Name) == ".txt" {
			replayFile = file
			break
		}
	}

	if replayFile == nil {
		return nil, fmt.Errorf("no replay file found in zip")
	}

	reader, err := replayFile.Open()
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	return e.parseFrameData(data)
}

// parseFrameData parses the frame data from the replay file
func (e *EchoReplayCodec) parseFrameData(data []byte) ([]*rtapi.LobbySessionStateFrame, error) {
	lines := bytes.Split(data, []byte("\n"))
	var frames []*rtapi.LobbySessionStateFrame

	unmarshaler := &protojson.UnmarshalOptions{
		DiscardUnknown: false,
	}

	for _, line := range lines {
		if len(line) == 0 {
			continue
		}

		parts := bytes.Split(line, []byte("\t"))
		if len(parts) < 2 {
			continue
		}

		// Parse timestamp
		timestampStr := string(parts[0])
		timestamp, err := time.Parse("2006/01/02 15:04:05.000", timestampStr)
		if err != nil {
			return nil, fmt.Errorf("invalid timestamp format: %s", timestampStr)
		}

		// Parse session data
		sessionResponse := &apigame.SessionResponse{}
		if err := unmarshaler.Unmarshal(parts[1], sessionResponse); err != nil {
			return nil, fmt.Errorf("failed to unmarshal session data: %w", err)
		}

		// Create frame
		frame := &rtapi.LobbySessionStateFrame{
			FrameIndex: uint32(len(frames)),
			Timestamp:  timestamppb.New(timestamp),
			Session:    sessionResponse,
		}

		// Parse user bones if present (parts[2])
		if len(parts) > 2 && len(parts[2]) > 0 {
			userBones := &apigame.UserBonesResponse{}
			if err := unmarshaler.Unmarshal(parts[2], userBones); err == nil {
				frame.UserBones = userBones
			}
		}

		frames = append(frames, frame)
	}

	return frames, nil
}

// Close closes the codec and underlying files
func (e *EchoReplayCodec) Close() error {
	var err error

	if e.zipWriter != nil {
		if finErr := e.Finalize(); finErr != nil && err == nil {
			err = finErr
		}
		if closeErr := e.zipWriter.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}

	if e.zipReader != nil {
		if closeErr := e.zipReader.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}

	if e.file != nil {
		if closeErr := e.file.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}

	return err
}

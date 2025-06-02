package main

import (
	"archive/zip"
	"io"
	"os"
	"testing"

	"github.com/echotools/nevr-common/rtapi"
	"google.golang.org/protobuf/proto"
)

func TestGameAPISessionCompression(t *testing.T) {
	// Open the echoreplay file as a zip
	fn := "rec_2025-05-29_18-14-10.echoreplay"

	file, err := os.Open(fn)
	if err != nil {
		t.Fatalf("failed to open file %s: %v", fn, err)
	}
	defer file.Close()
	stat, err := file.Stat()
	if err != nil {
		t.Fatalf("failed to stat file %s: %v", fn, err)
	}
	zipReader, err := zip.NewReader(file, stat.Size())
	if err != nil {
		t.Fatalf("failed to create zip reader for file %s: %v", fn, err)
	}
	// Find the GameAPISession file in the zip
	var gameAPISessionFile *zip.File
	for _, f := range zipReader.File {
		if f.Name == fn {
			gameAPISessionFile = f
			break
		}
	}
	if gameAPISessionFile == nil {
		t.Fatalf("not found in zip file %s", fn)
	}
	// Open the GameAPISession file
	fileReader, err := gameAPISessionFile.Open()
	if err != nil {
		t.Fatalf("failed to open GameAPISession file %s: %v", gameAPISessionFile.Name, err)
	}
	// Read the file content
	data, err := io.ReadAll(fileReader)
	if err != nil {
		t.Fatalf("failed to read GameAPISession file %s: %v", gameAPISessionFile.Name, err)
	}
	// Close the file reader
	if err := fileReader.Close(); err != nil {
		t.Fatalf("failed to close GameAPISession file %s: %v", gameAPISessionFile.Name, err)
	}

	frames := &[]*rtapi.SessionUpdateMessage{}
	// Unmarshal the line into a GameAPISession
	if err := UnmarshalReplay(data, frames); err != nil {
		t.Fatalf("failed to unmarshal GameAPISession line: %v", err)
	}

	outputFile, err := os.Create("GameAPISession_protobuf.bin")
	if err != nil {
		t.Fatalf("failed to create output file GameAPISession_protobuf.bin: %v", err)
	}
	defer outputFile.Close()
	// Create a protobuf encoder
	t.Logf("Encoding %d frames to protobuf", len(*frames))
	for _, frame := range *frames {
		// Encode the GameAPISession to protobuf
		protbufEncodedData, err := proto.Marshal(frame)
		if err != nil {
			t.Fatalf("failed to marshal GameAPISession to protobuf: %v", err)
		}
		t.Logf("Encoded GameAPISession to protobuf, size: %d bytes", len(protbufEncodedData))
		// Write the protobuf encoded data to the output file
		if _, err := outputFile.Write(protbufEncodedData); err != nil {
			t.Fatalf("failed to write GameAPISession_protobuf.bin: %v", err)
		}
	}
	// Close the output file
	if err := outputFile.Close(); err != nil {
		t.Fatalf("failed to close output file GameAPISession_protobuf.bin: %v", err)
	}
	// Read the protobuf encoded data back from the file
	protbufEncodedData, err := os.ReadFile("GameAPISession_protobuf.bin")
	if err != nil {
		t.Fatalf("failed to read GameAPISession_protobuf.bin: %v", err)
	}

	// Verify the size of the protobuf encoded data
	if len(protbufEncodedData) == 0 {
		t.Fatalf("no data was written to GameAPISession_protobuf.bin")
	}
	// Print the size of the protobuf encoded data
	t.Logf("Protobuf encoded data size: %d bytes", len(protbufEncodedData))
	// Verify the size of the original data
	if len(data) == 0 {
		t.Fatalf("no data was read from GameAPISession.json")
	}
	t.Logf("Original data size: %d bytes", len(data))
	// Print the compression ratio
	compressionRatio := float64(len(protbufEncodedData)) / float64(len(data))
	t.Logf("Compression ratio: %.2f", compressionRatio)
	// Check if the compression ratio is less than 0.5
	if compressionRatio >= 0.5 {
		t.Errorf("compression ratio is too high: %.2f, expected less than 0.5", compressionRatio)
	} else {
		t.Logf("compression ratio is acceptable: %.2f", compressionRatio)
	}
	// Check if the protobuf encoded data is smaller than the original data
	if len(protbufEncodedData) >= len(data) {
		t.Errorf("protobuf encoded data is not smaller than original data: %d >= %d", len(protbufEncodedData), len(data))
	} else {
		t.Logf("protobuf encoded data is smaller than original data: %d < %d", len(protbufEncodedData), len(data))
	}
}

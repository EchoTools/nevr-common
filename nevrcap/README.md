# Telemetry Package

High-performance nevrcap processing and streaming for NEVR lobby session data.

## Overview

This package provides optimized processing of game session frames with support for:

- **High-frequency frame processing** (600+ Hz target, achieves 14,000+ Hz)
- **Event detection** between consecutive frames
- **Multiple streaming codecs** (.nevrcap, .echoreplay, WebSocket)
- **File format conversion** utilities
- **Comprehensive benchmarking**

## Performance

🚀 **Performance Results (23x faster than target):**
- Frame Processing: **14,294 Hz** (target: 600 Hz)
- Event Detection: **1,351,617 ops/sec**
- Memory Efficient: 8.8 KB/op, 166 allocs/op
- File Compression: **57%** size reduction (.nevrcap vs .echoreplay)

## Usage

### Frame Processing

```go
import "github.com/echotools/nevr-common/v3/nevrcap"

// Create processor
processor := nevrcap.NewFrameProcessor()

// Process raw game data
frame, err := processor.ProcessFrame(sessionData, userBonesData, timestamp)
if err != nil {
    log.Fatal(err)
}

// Events are automatically detected
fmt.Printf("Detected %d events in frame %d\n", len(frame.Events), frame.FrameIndex)
```

### Streaming Codecs

#### Zstd Codec (.nevrcap files)

```go
// Writing
writer, err := nevrcap.NewZstdCodecWriter("capture.nevrcap")
defer writer.Close()

header := &nevrcap.TelemetryHeader{
    CaptureId: "session-123",
    CreatedAt: timestamppb.Now(),
}
writer.WriteHeader(header)
writer.WriteFrame(frame)

// Reading
reader, err := nevrcap.NewZstdCodecReader("capture.nevrcap")
defer reader.Close()

header, err := reader.ReadHeader()
frame, err := reader.ReadFrame()
```

#### EchoReplay Codec (.echoreplay files)

```go
// Writing
writer, err := nevrcap.NewEchoReplayCodecWriter("replay.echoreplay")
defer writer.Close()

writer.WriteFrame(frame)

// Reading
reader, err := nevrcap.NewEchoReplayCodecReader("replay.echoreplay")
defer reader.Close()

frames, err := reader.ReadFrames()
```

#### WebSocket Codec

```go
// Configure streaming options
options := nevrcap.WebSocketCodecOptions{
    SendSessionData: true,
    SendUserBones:   true,
    SendEvents:      true,
    ProcessFrames:   true,
}

codec, err := nevrcap.NewWebSocketCodec("ws://localhost:8080", options)
defer codec.Close()

// Stream processed frame
codec.StreamProcessedFrame(sessionData, userBonesData, timestamp)

// Or stream raw data
codec.StreamRawData(sessionData, userBonesData, timestamp)
```

### File Conversion

```go
// Convert .echoreplay to .nevrcap
err := nevrcap.ConvertEchoReplayToNevrcap("input.echoreplay", "output.nevrcap")

// Convert .nevrcap to .echoreplay  
err := nevrcap.ConvertNevrcapToEchoReplay("input.nevrcap", "output.echoreplay")
```

## Event Detection

The system automatically detects various game events:

### Game State Events
- Round started/ended
- Match ended
- Scoreboard updates
- Game paused/unpaused

### Player Events
- Player joined/left
- Team switches
- Emote playing

### Disc Events
- Possession changes
- Disc thrown/caught

### Stat Events
- Saves, stuns, passes
- Catches, steals, blocks
- Interceptions, assists
- Shots taken

## File Formats

### .nevrcap Format
- **Compression**: Zstd
- **Serialization**: Protocol Buffers
- **Structure**: Header + length-delimited frames
- **Features**: Event detection, streaming support
- **Size**: ~57% of .echoreplay size

### .echoreplay Format  
- **Compression**: ZIP
- **Serialization**: JSON
- **Structure**: ZIP archive with replay.txt
- **Features**: Legacy compatibility
- **Size**: Baseline reference

## Benchmarks

Run benchmarks:
```bash
go test -bench=. -benchmem ./nevrcap
```

Update benchmark documentation:
```bash
./scripts/update_benchmarks.sh
```

See [BENCHMARKS.md](../BENCHMARKS.md) for detailed performance metrics.

## Demo

Run the interactive demo:
```bash
cd examples/nevrcap-demo
go run main.go
```

## Optimization Features

1. **Pre-allocated Structures**: Reuses objects to minimize GC pressure
2. **Efficient Event Detection**: O(1) player lookups using maps
3. **Streaming Support**: Processes data incrementally
4. **High-Performance Compression**: Zstd provides optimal speed/size ratio
5. **Memory Pooling**: Minimal allocations for high-frequency operations

## Architecture

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│ Raw Game Data   │───▶│ Frame Processor  │───▶│ Event Detection │
│ (JSON bytes)    │    │ (600+ Hz)        │    │ (Δ comparison)  │
└─────────────────┘    └──────────────────┘    └─────────────────┘
                                │
                                ▼
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│ File Conversion │◀───│ Streaming Codecs │◀───│ LobbySessionFrame│
│ (.echoreplay ⟷) │    │ (Zstd/WS/Zip)   │    │ (with events)   │
└─────────────────┘    └──────────────────┘    └─────────────────┘
```

## Thread Safety

- Frame processors are **not** thread-safe (use separate instances per goroutine)
- Codecs are **not** thread-safe (one writer/reader per file)
- Event detectors maintain state (use separate instances per stream)

## Known Limitations

- Event detection uses heuristics that may need game-specific tuning
- WebSocket codec requires external server implementation
- Large files should use streaming to avoid memory issues
- Some edge cases in event detection may need refinement

## Contributing

When adding new event types:
1. Update `nevrcap.proto` with new event message
2. Regenerate protobuf code: `go generate ./nevrcap`
3. Add detection logic in `events.go`
4. Add tests in `*_test.go`
5. Update benchmarks if needed
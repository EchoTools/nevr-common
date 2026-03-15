# Telemetry V2 — Layered Capture Format

Game-agnostic capture envelope with game-specific payload extensions.

## Architecture

```
telemetry/v2/capture.proto      # Game-agnostic: Envelope, CaptureHeader, Frame, CaptureFooter
telemetry/v2/echo_arena.proto   # Echo Arena: EchoArenaHeader, EchoArenaFrame, events, enums
spatial/v1/types.proto           # Shared: Vec3, Quat, Pose primitives
```

### Layering

The core capture format (`capture.proto`) knows nothing about any specific game.
Game-specific payloads are injected through `oneof` extension points:

- `CaptureHeader.game_header` — session metadata (roster, map, match type)
- `Frame.payload` — per-frame game state and events

To add a new game, create a new `<game>.proto` in this package and add a
field to each `oneof`. No changes to existing game protos are needed.

## Envelope Model

A capture stream is a sequence of length-delimited `Envelope` messages:

1. **One `Envelope{header}`** — `CaptureHeader` with format version, timestamp, metadata, and game-specific session info.
2. **N `Envelope{frame}`** — `Frame` messages at capture rate (typically 60/s). Each frame carries a sequential index, a millisecond offset from the header timestamp, and a game-specific payload.
3. **One `Envelope{footer}`** — `CaptureFooter` with frame count, duration, and seek indexes (keyframe + event type).

### Footer and Seeking

The `CaptureFooter` contains:

- **`footer_offset`** — byte offset from file start; readers can seek directly to the footer without scanning.
- **`keyframe_index`** — maps frame indices to byte offsets for random access.
- **`event_index`** — maps `EventType` enum values to frame indices for type-based queries (e.g., "find all goals").

## Consuming

```go
import telemetryv2 "buf.build/gen/go/echotools/nevr-api/protocolbuffers/go/telemetry/v2"

// Read envelope stream
for {
    env := &telemetryv2.Envelope{}
    if err := readDelimited(reader, env); err != nil {
        break
    }
    switch msg := env.Message.(type) {
    case *telemetryv2.Envelope_Header:
        header := msg.Header
        if ea := header.GetEchoArena(); ea != nil {
            // Echo Arena session: ea.SessionId, ea.MapName, etc.
        }
    case *telemetryv2.Envelope_Frame:
        frame := msg.Frame
        if ea := frame.GetEchoArena(); ea != nil {
            // Echo Arena frame: ea.GameStatus, ea.Players, ea.Events, etc.
        }
    case *telemetryv2.Envelope_Footer:
        footer := msg.Footer
        // footer.FrameCount, footer.KeyframeIndex, footer.EventIndex
    }
}
```

## Clean V2 Break

V2 imports nothing from `telemetry/v1` or `apigame/v1`. All types
(enums, events, player state) are redefined in `echo_arena.proto` to
allow independent evolution.

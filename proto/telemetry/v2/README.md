# Telemetry V2 Format — nevrcap

Layered, extensible protocol buffer format for streaming VR capture data.

## Overview

Telemetry V2 is a **game-agnostic capture format** with a core timing envelope and game-specific payloads. The core provides file structure, timestamping, and seeking indexes. Game payloads (starting with Echo Arena) carry all game state, players, and events.

### Architecture

```
Envelope (oneof message)
├── CaptureHeader
│   ├── capture_id, created_at, format_version, metadata
│   └── oneof game_header
│       └── EchoArenaHeader (field 10)  — roster, map, match type, skeleton
├── Frame
│   ├── frame_index (uint32, 0-based sequential)
│   ├── timestamp_offset_ms (uint32, ms since created_at)
│   └── oneof payload
│       └── EchoArenaFrame (field 10)   — game state, players, disc, events
└── CaptureFooter
    ├── frame_count, duration_ms, total_bytes
    ├── repeated KeyframeEntry keyframe_index
    └── repeated EventIndexEntry event_index
```

### Key Improvements over V1

- **Layered architecture** — game-agnostic core separates timing from game data
- **Session-scoped constants** moved to `CaptureHeader` (written once)
- **Timestamp deltas** (`uint32 timestamp_offset_ms`) instead of full `google.protobuf.Timestamp`
- **float32** instead of float64 for all spatial data
- **Quaternions** instead of 3×3 rotation matrices (36 bytes → 16 bytes per orientation)
- **Packed bytes** for bone data enabling zero-copy serialization from C++
- **Proto enums** instead of string enums
- **Bitmask** for boolean flags (5 bools × 2 bytes → 2 bytes)
- **V2-native events** — no v1 or apigame.v1 dependencies
- **Trailing index** — CaptureFooter with keyframe and event indexes for file seeking
- **Dynamic skeleton** — bone count and stride in header, per-player override
- **Removed per-frame stats** (derivable from event stream)
- **Removed redundant fields** (duplicate scores, derived display strings)
- **Clean break from V1** — zero v1/apigame imports

## Wire Size Comparison

| Scenario | V1 (bytes) | V2 (bytes) | Reduction |
|----------|-----------|-----------|-----------|
| 2 players per frame | 1,267 | 329 | 74.0% |
| 10 players per frame | 5,109 | 1,359 | 73.4% |
| **60 FPS, 10 players** | **299.4 KB/s** | **79.6 KB/s** | **73.4%** |

*Note: V1 estimates include SessionResponse + PlayerBonesResponse embedded per frame*

## Package Structure

```
proto/
├── spatial/v1/types.proto              # Vec3, Quat, Pose primitives (shared)
└── telemetry/v2/
    ├── capture.proto                   # Core: Envelope, CaptureHeader, Frame, CaptureFooter
    └── echo_arena.proto                # Echo Arena: header, frame, events, enums
```

## Core Types (`capture.proto`)

### Envelope

Top-level wrapper for all messages in a capture stream or file.

```protobuf
message Envelope {
  oneof message {
    CaptureHeader header = 1;
    Frame frame = 2;
    CaptureFooter footer = 3;
  }
}
```

### CaptureHeader

Session metadata, written once at the start. Contains:
- `capture_id` — UUID for this capture
- `created_at` — Base timestamp for frame deltas
- `format_version` — Protocol version (2)
- `metadata` — Arbitrary key-value pairs
- `oneof game_header` — Game-specific session data (field 10 = Echo Arena)

### Frame

Pure timing envelope — all game data lives in the payload:
- `frame_index` — Sequential, 0-based
- `timestamp_offset_ms` — Milliseconds since `CaptureHeader.created_at`
- `oneof payload` — Game-specific frame data (field 10 = Echo Arena)

Core frame overhead: ≤ 10 bytes.

### CaptureFooter

File-level metadata and seeking indexes, written at stream close:
- `frame_count` — Total frames
- `duration_ms` — Total duration
- `total_bytes` — File size for integrity
- `keyframe_index` — `KeyframeEntry` list for time-based seeking
- `event_index` — `EventIndexEntry` list for event-based seeking

Truncated files (missing footer) are still readable — the footer is for seeking only.

## Extension Model

New games plug in via the `oneof` fields in `CaptureHeader` and `Frame`:

```protobuf
// In CaptureHeader:
oneof game_header {
  EchoArenaHeader echo_arena = 10;
  // FutureGameHeader future_game = 11;  // Add new games here
}

// In Frame:
oneof payload {
  EchoArenaFrame echo_arena = 10;
  // FutureGameFrame future_game = 11;   // Add new games here
}
```

Field number ranges:
- **10-99**: Game-specific payloads (one per game)
- **100-199**: Non-game payloads (annotations, debug — reserved for future use)

## Echo Arena Types (`echo_arena.proto`)

### EchoArenaHeader

Session metadata for Echo Arena matches:
- `session_id`, `map_name`, `match_type`, `client_name`
- `private_match`, `tournament_match`, `total_round_count`
- `initial_roster` — `repeated PlayerInfo` at capture start
- `skeleton` — Default `SkeletonLayout` for bone data

### SkeletonLayout

Dynamic bone layout (stored in header, overridable per-player):
- `bone_count` — Number of bones (e.g., 22 for Echo VR)
- `transform_stride` — Bytes per bone transform (e.g., 12 = 3 × float32)
- `orientation_stride` — Bytes per bone orientation (e.g., 16 = 4 × float32)

### EchoArenaFrame

Per-frame game state including:
- `game_status`, `game_clock`, `pause_state`
- `disc` — `DiscState` (pose + velocity + bounce count)
- `players` — `repeated PlayerState` (head, body, hands, velocity, flags, ping)
- `player_bones` — `repeated PlayerBones` (zero-copy bone data)
- `disc_holder_slot` — `optional int32` (-1 = free disc, uses proto3 field presence)
- `vr_root` — Recorder's VR tracking origin
- `blue_points`, `orange_points`, `round_number`
- `events` — `repeated EchoEvent` (v2-native events)

### Events (20+ types)

All events are wrapped in `EchoEvent` with a `oneof event`:

| Category | Events |
|----------|--------|
| Game State (10-19) | RoundStarted, RoundPaused, RoundUnpaused, RoundEnded, MatchEnded, ScoreboardUpdated |
| Player (20-29) | PlayerJoined, PlayerLeft, PlayerSwitchedTeam, EmotePlayed |
| Disc (30-39) | DiscPossessionChanged, DiscThrown, DiscCaught |
| Scoring (40-49) | GoalScored, PlayerGoal |
| Stats (50-59) | PlayerSave, PlayerStun, PlayerPass, PlayerSteal, PlayerBlock, PlayerInterception, PlayerAssist, PlayerShotTaken |
| Misc (60-69) | GenericEvent |

Key differences from V1 events:
- **`GoalScored`** uses `Role` enum and `GoalType` enum instead of strings
- **`PlayerJoined`** has flat fields (slot, account_number, display_name, role) instead of embedded `TeamMember`
- **`DiscThrown`** uses `ThrowDetails` with all float32 fields (replacing `LastThrowInfo` float64)
- **`RoundPaused`/`RoundUnpaused`** use `PauseState` enum instead of `apigame.v1.PauseState` message

### Enums

- **`GameStatus`**: 10 values (UNSPECIFIED through POST_SUDDEN_DEATH)
- **`MatchType`**: 9 values (UNSPECIFIED through TOURNAMENT)
- **`PauseState`**: 5 values (UNSPECIFIED through AUTOPAUSE_REPLAY)
- **`Role`**: 6 values (UNSPECIFIED, BLUE_TEAM, ORANGE_TEAM, SPECTATOR, SOCIAL_PARTICIPANT, MODERATOR)
- **`GoalType`**: 6 values (UNSPECIFIED, INSIDE_SHOT, LONG_SHOT, BOUNCE_SHOT, LONG_BOUNCE_SHOT, SELF_GOAL)

## Usage

### Go

```go
import (
    spatial "github.com/echotools/nevr-common/v4/gen/go/spatial/v1"
    telemetryv2 "github.com/echotools/nevr-common/v4/gen/go/telemetry/v2"
    "google.golang.org/protobuf/proto"
)

// Create capture header (once per session)
header := &telemetryv2.CaptureHeader{
    CaptureId:     uuid.New().String(),
    CreatedAt:     timestamppb.Now(),
    FormatVersion: 2,
    GameHeader: &telemetryv2.CaptureHeader_EchoArena{
        EchoArena: &telemetryv2.EchoArenaHeader{
            SessionId:       "abc123",
            MapName:         "mpl_arena_a",
            MatchType:       telemetryv2.MatchType_MATCH_TYPE_ARENA,
            TotalRoundCount: 3,
            Skeleton: &telemetryv2.SkeletonLayout{
                BoneCount:        22,
                TransformStride:  12,
                OrientationStride: 16,
            },
        },
    },
}

// Create frame (60 times per second)
frame := &telemetryv2.Frame{
    FrameIndex:        100,
    TimestampOffsetMs: 1667, // milliseconds since header.CreatedAt
    Payload: &telemetryv2.Frame_EchoArena{
        EchoArena: &telemetryv2.EchoArenaFrame{
            GameStatus: telemetryv2.GameStatus_GAME_STATUS_PLAYING,
            GameClock:  120.5,
            Disc: &telemetryv2.DiscState{
                Pose: &spatial.Pose{
                    Position:    &spatial.Vec3{X: 1.0, Y: 2.0, Z: 3.0},
                    Orientation: &spatial.Quat{X: 0.0, Y: 0.0, Z: 0.0, W: 1.0},
                },
                Velocity:    &spatial.Vec3{X: 5.0, Y: 2.5, Z: 1.0},
                BounceCount: 3,
            },
            DiscHolderSlot: proto.Int32(0),
            Players: []*telemetryv2.PlayerState{
                {
                    Slot: 0,
                    Head: &spatial.Pose{
                        Position:    &spatial.Vec3{X: 10.0, Y: 11.0, Z: 12.0},
                        Orientation: &spatial.Quat{X: 0.0, Y: 0.0, Z: 0.0, W: 1.0},
                    },
                    // ... body, left_hand, right_hand ...
                    Velocity: &spatial.Vec3{X: 0.1, Y: 0.2, Z: 0.3},
                    Flags:    0b00001, // stunned
                    Ping:     25,
                },
            },
        },
    },
}

// Wrap in Envelope for streaming
env := &telemetryv2.Envelope{
    Message: &telemetryv2.Envelope_Frame{Frame: frame},
}
```

### C++ (Zero-Copy Bone Data)

```cpp
#include "telemetry/v2/echo_arena.pb.h"

// Player bone transforms (22 bones × 3 floats = 264 bytes)
float bone_positions[22][3];
// ... populate from game engine memory ...

// Player bone orientations (22 bones × 4 floats = 352 bytes)
float bone_orientations[22][4];
// ... populate from game engine memory ...

telemetry::v2::PlayerBones bones;
bones.set_slot(0);
bones.set_transforms(bone_positions, 264);      // zero-copy memcpy
bones.set_orientations(bone_orientations, 352); // zero-copy memcpy
```

### C#

```csharp
using Nevr.Spatial.V1;
using Nevr.Telemetry.V2;

var frame = new Frame
{
    FrameIndex = 100,
    TimestampOffsetMs = 1667,
    EchoArena = new EchoArenaFrame
    {
        GameStatus = GameStatus.Playing,
        GameClock = 120.5f,
        Disc = new DiscState
        {
            Pose = new Pose
            {
                Position = new Vec3 { X = 1.0f, Y = 2.0f, Z = 3.0f },
                Orientation = new Quat { X = 0.0f, Y = 0.0f, Z = 0.0f, W = 1.0f }
            },
            Velocity = new Vec3 { X = 5.0f, Y = 2.5f, Z = 1.0f },
            BounceCount = 3
        }
    }
};
```

## Data Types

### Spatial Primitives (`spatial/v1`)

- **`Vec3`**: 12 bytes (3 × float32) — position, velocity, direction
- **`Quat`**: 16 bytes (4 × float32) — orientation as unit quaternion
- **`Pose`**: 28 bytes — rigid body position + orientation

### Player Flags Bitmask

| Bit | Meaning |
|-----|---------|
| 0   | stunned |
| 1   | invulnerable |
| 2   | blocking |
| 3   | possession |
| 4   | is_emote_playing |
| 5-31 | reserved |

Example: `flags = 0b00001` = stunned, `flags = 0b01000` = has possession

## Streaming Format

A V2 capture stream consists of length-delimited `Envelope` messages:

1. **One `Envelope`** containing `CaptureHeader`
2. **N `Envelope`** messages containing `Frame` (e.g., 60 per second)
3. **One `Envelope`** containing `CaptureFooter` (optional, written at close)

```go
// Write header
writeDelimited(&telemetryv2.Envelope{
    Message: &telemetryv2.Envelope_Header{Header: header},
})

// Write frames
for frame := range frames {
    writeDelimited(&telemetryv2.Envelope{
        Message: &telemetryv2.Envelope_Frame{Frame: frame},
    })
}

// Write footer (optional)
writeDelimited(&telemetryv2.Envelope{
    Message: &telemetryv2.Envelope_Footer{Footer: footer},
})
```

### Trailing Index

The `CaptureFooter` contains two index types for seeking:

- **`KeyframeEntry`**: Maps `frame_index` → `byte_offset` for time-based seeking
- **`EventIndexEntry`**: Maps `event_type` → `frame_indices` for event-based seeking

Truncated files (missing footer) are still fully readable — just without seek capability.

## Design Rationale

### Why a Layered Architecture?

The core format (Envelope, Frame, CaptureHeader, CaptureFooter) is game-agnostic. This enables:
- **Multi-game support** via `oneof` extension points
- **Shared tooling** for capture/replay regardless of game
- **Independent evolution** of core format and game payloads

### Why a Game-Agnostic Core?

The `Frame` message is a pure timing envelope with only `frame_index`, `timestamp_offset_ms`, and a `oneof payload`. This ensures:
- Core frame overhead ≤ 10 bytes
- Game-specific tools only need to understand their own payload type
- File format tools (seeking, splitting, concatenation) work without game knowledge

### Why Quaternions?

Rotation matrices (3×3 = 9 floats = 36 bytes) are redundant for storing orientation. Quaternions (4 floats = 16 bytes) provide:
- 56% smaller wire size
- No gimbal lock
- Efficient interpolation (slerp)
- Direct compatibility with game engine math libraries

### Why Bytes for Bone Data?

The `PlayerBones.transforms` and `PlayerBones.orientations` fields use raw `bytes` instead of `repeated Vec3/Quat` because:
- **Zero-copy serialization**: Direct `memcpy` from C++ game engine memory
- **Performance**: Avoids per-bone message overhead (22 bones × field tag overhead)
- **Layout control**: Fixed stride defined by `SkeletonLayout` in header

The `SkeletonLayout` in `EchoArenaHeader` defines the default bone count and strides. Individual players can override this via `PlayerBones.skeleton_override` for different rigs.

### Why Remove Per-Frame Stats?

Player stats (goals, saves, stuns, etc.) are monotonically increasing counters. They're redundant with the event stream:
- `PlayerGoal` event increments goals counter
- `PlayerStun` event increments stuns counter
- etc.

Consumers can reconstruct stats from events, saving ~50 bytes per player per frame.

## Testing

Run the wire size comparison example:

```bash
go run examples/size_comparison.go
```

Output:
```
=== Telemetry V2 Wire Size Comparison ===

V2 Frame with 2 players: 329 bytes
V2 Frame with 10 players: 1359 bytes

V1 Frame with 2 players: 1267 bytes
V1 Frame with 10 players: 5109 bytes

=== Wire Size Reduction ===
2 players: 74.0% reduction (1267 → 329 bytes)
10 players: 73.4% reduction (5109 → 1359 bytes)

=== Bandwidth at 60 FPS (10 players) ===
V1: 299.4 KB/s
V2: 79.6 KB/s
Savings: 219.7 KB/s (73.4%)
```

## Future Work

- gRPC streaming service for real-time telemetry push
- Compression benchmarks (gzip, zstd, lz4)
- Additional game payloads beyond Echo Arena
- GPU-accelerated bone decompression shaders

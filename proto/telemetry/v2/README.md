# Telemetry V2 Format — nevrcap v2

Layered, extensible VR capture format. The core provides a timing envelope and file structure. Game-specific payloads (starting with Echo Arena) carry all game state, players, and events.

## Overview

Telemetry V2 is a **game-agnostic capture envelope** with game-specific payloads attached via `oneof` extension points. This achieves **~73% wire size reduction** compared to V1 through:

- **Layered architecture** — Core envelope is game-agnostic; game state lives in oneof payloads
- **Session-scoped constants** moved to `CaptureHeader` (written once)
- **Timestamp deltas** (`uint32 timestamp_offset_ms`) instead of full `google.protobuf.Timestamp`
- **float32** instead of float64 for all spatial data
- **Quaternions** instead of 3×3 rotation matrices (36 bytes → 16 bytes per orientation)
- **Packed bytes** for bone data enabling zero-copy serialization from C++
- **Proto enums** instead of string enums
- **Bitmask** for boolean flags (5 bools × 2 bytes → 2 bytes)
- **V2-native events** — No v1/apigame imports; all events are self-contained
- **Removed per-frame stats** (derivable from event stream)
- **Removed redundant fields** (duplicate scores, derived display strings)

## Architecture

```
Envelope (oneof message)
├── CaptureHeader
│   ├── capture_id, created_at, format_version, metadata
│   └── oneof game_header
│       └── EchoArenaHeader = 10  (roster, map, match type, skeleton layout)
├── Frame
│   ├── frame_index (uint32, 0-based sequential)
│   ├── timestamp_offset_ms (uint32, ms since created_at)
│   └── oneof payload
│       └── EchoArenaFrame = 10   (all game state, players, disc, events)
└── CaptureFooter
    ├── frame_count, duration_ms, total_bytes
    ├── repeated KeyframeEntry keyframe_index
    └── repeated EventIndexEntry event_index
```

## Wire Size Comparison

| Scenario | V1 (bytes) | V2 (bytes) | Reduction |
|----------|-----------|-----------|-----------|
| 2 players per frame | 1,266 | 332 | 73.8% |
| 10 players per frame | 5,108 | 1,362 | 73.3% |
| **60 FPS, 10 players** | **299.3 KB/s** | **79.8 KB/s** | **73.3%** |

## Package Structure

```
proto/
├── spatial/v1/types.proto              # Vec3, Quat, Pose primitives (shared)
└── telemetry/v2/
    ├── capture.proto                   # Core: Envelope, CaptureHeader, Frame, CaptureFooter
    └── echo_arena.proto                # Echo Arena: header, frame, players, disc, events
```

- `capture.proto` — Game-agnostic core. Imports `echo_arena.proto` for the Echo Arena oneof variant.
- `echo_arena.proto` — All Echo Arena types. Only imports `spatial/v1/types.proto`. **No v1 telemetry or apigame imports.**

## Extension Model

New games plug into the format via oneof variants:

```protobuf
// In capture.proto — CaptureHeader
oneof game_header {
    EchoArenaHeader echo_arena = 10;
    // SomeOtherGameHeader other_game = 11;  // Future game
}

// In capture.proto — Frame
oneof payload {
    EchoArenaFrame echo_arena = 10;
    // SomeOtherGameFrame other_game = 11;   // Future game
}
```

Field numbers 10-99 are reserved for game-specific payloads.

## Usage

### Go

```go
import (
    spatial "github.com/echotools/nevr-common/v4/gen/go/spatial/v1"
    telemetryv2 "github.com/echotools/nevr-common/v4/gen/go/telemetry/v2"
    "google.golang.org/protobuf/types/known/timestamppb"
)

// Create capture header envelope (once per session)
headerEnv := &telemetryv2.Envelope{
    Message: &telemetryv2.Envelope_Header{
        Header: &telemetryv2.CaptureHeader{
            CaptureId:     uuid.New().String(),
            CreatedAt:     timestamppb.Now(),
            FormatVersion: 2,
            GameHeader: &telemetryv2.CaptureHeader_EchoArena{
                EchoArena: &telemetryv2.EchoArenaHeader{
                    SessionId:       "abc123",
                    MapName:         "mpl_arena_a",
                    MatchType:       telemetryv2.MatchType_MATCH_TYPE_ARENA,
                    TotalRoundCount: 3,
                    SkeletonLayout: &telemetryv2.SkeletonLayout{
                        BoneCount:         22,
                        TransformStride:   12,
                        OrientationStride: 16,
                    },
                },
            },
        },
    },
}

// Create frame envelope (60 times per second)
discHolder := int32(0)
frameEnv := &telemetryv2.Envelope{
    Message: &telemetryv2.Envelope_Frame{
        Frame: &telemetryv2.Frame{
            FrameIndex:        100,
            TimestampOffsetMs: 1667,
            Payload: &telemetryv2.Frame_EchoArena{
                EchoArena: &telemetryv2.EchoArenaFrame{
                    GameStatus:     telemetryv2.GameStatus_GAME_STATUS_PLAYING,
                    GameClock:      120.5,
                    DiscHolderSlot: &discHolder,
                    Disc: &telemetryv2.DiscState{
                        Pose: &spatial.Pose{
                            Position:    &spatial.Vec3{X: 1.0, Y: 2.0, Z: 3.0},
                            Orientation: &spatial.Quat{X: 0, Y: 0, Z: 0, W: 1},
                        },
                        Velocity:    &spatial.Vec3{X: 5.0, Y: 2.5, Z: 1.0},
                        BounceCount: 3,
                    },
                    Players: []*telemetryv2.PlayerState{
                        {
                            Slot: 0,
                            Head: &spatial.Pose{
                                Position:    &spatial.Vec3{X: 10, Y: 11, Z: 12},
                                Orientation: &spatial.Quat{X: 0, Y: 0, Z: 0, W: 1},
                            },
                            // ... body, left_hand, right_hand ...
                            Velocity: &spatial.Vec3{X: 0.1, Y: 0.2, Z: 0.3},
                            Flags:    0b00001, // stunned
                            Ping:     25,
                        },
                    },
                },
            },
        },
    },
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

var envelope = new Envelope
{
    Frame = new Frame
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
    }
};
```

## Data Types

### Spatial Primitives (`spatial/v1`)

- **`Vec3`**: 12 bytes (3 × float32) — position, velocity, direction
- **`Quat`**: 16 bytes (4 × float32) — orientation as unit quaternion
- **`Pose`**: 28 bytes — rigid body position + orientation

### Enums

- **`GameStatus`**: PRE_MATCH, ROUND_START, PLAYING, SCORE, ROUND_OVER, POST_MATCH, etc.
- **`MatchType`**: ARENA, SOCIAL_PUBLIC, SOCIAL_PRIVATE, COMBAT, TOURNAMENT, etc.
- **`PauseState`**: NOT_PAUSED, PAUSED, UNPAUSING, AUTOPAUSE_REPLAY
- **`Role`**: BLUE_TEAM, ORANGE_TEAM, SPECTATOR, SOCIAL_PARTICIPANT, MODERATOR
- **`GoalType`**: INSIDE_SHOT, LONG_SHOT, BOUNCE_SHOT, LONG_BOUNCE_SHOT, SELF_GOAL

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

### Dynamic SkeletonLayout

The `EchoArenaHeader.skeleton_layout` field describes the bone data format:

| Field | Default | Description |
|-------|---------|-------------|
| `bone_count` | 22 | Number of bones per player |
| `transform_stride` | 12 | Bytes per bone transform (3 × float32) |
| `orientation_stride` | 16 | Bytes per bone orientation (4 × float32 quaternion) |

Individual players can override the layout via `PlayerBones.skeleton_override`.

## Streaming Format

A V2 capture stream consists of length-delimited `Envelope` messages:

1. **One `Envelope{header}`** — Contains `CaptureHeader` with game-specific header
2. **N `Envelope{frame}`** — Contains `Frame` with game-specific payload (60 per second)
3. **One `Envelope{footer}`** (optional) — Contains `CaptureFooter` with index data

### Trailing Index

The optional `CaptureFooter` contains two index structures for random access:

- **`KeyframeEntry`** — Maps frame indices to byte offsets for seeking
- **`EventIndexEntry`** — Maps event types to frames for fast event lookup

These are only written for seekable file formats, not for real-time streams.

## Echo Arena Events

V2 defines 20+ event types natively (no v1 imports):

| Range | Category | Events |
|-------|----------|--------|
| 10-19 | Game State | RoundStarted, RoundPaused, RoundUnpaused, RoundEnded, MatchEnded, ScoreboardUpdated |
| 20-29 | Player | PlayerJoined, PlayerLeft, PlayerSwitchedTeam, EmotePlayed |
| 30-39 | Disc | DiscPossessionChanged, DiscThrown, DiscCaught |
| 40-49 | Scoring | GoalScored, PlayerGoal |
| 50-59 | Stats | PlayerSave, PlayerStun, PlayerPass, PlayerSteal, PlayerBlock, PlayerInterception, PlayerAssist, PlayerShotTaken |
| 60-69 | Misc | GenericEvent |

Events are wrapped in `EchoEvent` (oneof wrapper) and stored in `EchoArenaFrame.events`.

## Design Rationale

### Why a Layered Architecture?

Separating the core envelope from game-specific payloads provides:
- **Reusability** — The same capture format works for any VR game
- **Independent evolution** — Game payloads evolve without breaking the core
- **Tooling** — Generic tools can read timing/seeking without game knowledge
- **Future-proofing** — New games add a single oneof variant, no schema redesign

### Why Quaternions?

Rotation matrices (3×3 = 9 floats = 36 bytes) are redundant for storing orientation. Quaternions (4 floats = 16 bytes) provide:
- 56% smaller wire size
- No gimbal lock
- Efficient interpolation (slerp)
- Direct compatibility with game engine math libraries

### Why Bytes for Bone Data?

The `PlayerBones.transforms` and `PlayerBones.orientations` fields use raw `bytes` instead of `repeated Vec3/Quat` because:
- **Zero-copy serialization**: Direct `memcpy` from C++ game engine memory (little-endian)
- **Performance**: Avoids per-bone message overhead (22 bones × field tag overhead)
- **Layout control**: Fixed stride (12 bytes/bone for transforms, 16 bytes/bone for orientations)
- **Immutable contract**: C++ zero-copy compatibility is a hard requirement

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

V2 Envelope (Frame) with 2 players: 332 bytes
V2 Envelope (Frame) with 10 players: 1362 bytes
V2 Envelope (Header): 139 bytes

V1 Frame with 2 players: 1266 bytes
V1 Frame with 10 players: 5108 bytes

=== Wire Size Reduction ===
2 players: 73.8% reduction (1266 → 332 bytes)
10 players: 73.3% reduction (5108 → 1362 bytes)

=== Bandwidth at 60 FPS (10 players) ===
V1: 299.3 KB/s
V2: 79.8 KB/s
Savings: 219.5 KB/s (73.3%)
```

## Future Work

- Additional game payloads (other VR titles)
- gRPC streaming service for real-time telemetry push
- Compression benchmarks (gzip, zstd, lz4)
- Binary diff encoding for inter-frame deltas
- GPU-accelerated bone decompression shaders

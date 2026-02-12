# Telemetry V2 Format

Optimized protocol buffer format for streaming game session telemetry data.

## Overview

Telemetry V2 achieves **~73% wire size reduction** compared to V1 through:

- **Session-scoped constants** moved to `CaptureHeader` (written once)
- **Timestamp deltas** (`uint32 timestamp_offset_ms`) instead of full `google.protobuf.Timestamp`
- **float32** instead of float64 for all spatial data
- **Quaternions** instead of 3×3 rotation matrices (36 bytes → 16 bytes per orientation)
- **Packed bytes** for bone data enabling zero-copy serialization from C++
- **Proto enums** instead of string enums
- **Bitmask** for boolean flags (5 bools × 2 bytes → 2 bytes)
- **Removed per-frame stats** (derivable from event stream)
- **Removed redundant fields** (duplicate scores, derived display strings)

## Wire Size Comparison

| Scenario | V1 (bytes) | V2 (bytes) | Reduction |
|----------|-----------|-----------|-----------|
| 2 players per frame | 1,267 | 324 | 74.4% |
| 10 players per frame | 5,109 | 1,354 | 73.5% |
| **60 FPS, 10 players** | **299.4 KB/s** | **79.3 KB/s** | **73.5%** |

*Note: V1 estimates include SessionResponse + PlayerBonesResponse embedded per frame*

## Package Structure

```
proto/
├── spatial/v1/types.proto          # Vec3, Quat, Pose primitives
└── telemetry/v2/frame.proto        # CaptureHeader, Frame, PlayerState, etc.
```

## Usage

### Go

```go
import (
    spatial "github.com/echotools/nevr-common/v4/gen/go/spatial/v1"
    telemetryv2 "github.com/echotools/nevr-common/v4/gen/go/telemetry/v2"
)

// Create capture header (once per session)
header := &telemetryv2.CaptureHeader{
    CaptureId:       uuid.New().String(),
    CreatedAt:       timestamppb.Now(),
    SessionId:       "abc123",
    MapName:         "mpl_arena_a",
    MatchType:       telemetryv2.MatchType_MATCH_TYPE_ARENA,
    TotalRoundCount: 3,
}

// Create frame (60 times per second)
frame := &telemetryv2.Frame{
    FrameIndex:        100,
    TimestampOffsetMs: 1667, // milliseconds since header.CreatedAt
    GameStatus:        telemetryv2.GameStatus_GAME_STATUS_PLAYING,
    GameClock:         120.5,
    Disc: &telemetryv2.DiscState{
        Pose: &spatial.Pose{
            Position:    &spatial.Vec3{X: 1.0, Y: 2.0, Z: 3.0},
            Orientation: &spatial.Quat{X: 0.0, Y: 0.0, Z: 0.0, W: 1.0},
        },
        Velocity:    &spatial.Vec3{X: 5.0, Y: 2.5, Z: 1.0},
        BounceCount: 3,
    },
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
}
```

### C++ (Zero-Copy Bone Data)

```cpp
#include "telemetry/v2/frame.pb.h"

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
- **`GoalType`**: INSIDE_SHOT, LONG_SHOT, BOUNCE_SHOT, LONG_BOUNCE_SHOT

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

## Migration from V1

V1 remains fully supported for backward compatibility. Both formats can coexist in the same codebase:

```go
// V1 envelope
v1Env := &telemetryv1.Envelope{
    Message: &telemetryv1.Envelope_Header{
        Header: v1Header,
    },
}

// V2 envelope
v2Env := &telemetryv2.EnvelopeV2{
    Message: &telemetryv2.EnvelopeV2_HeaderV2{
        HeaderV2: v2Header,
    },
}
```

V2 reuses V1 event types (`LobbySessionEvent` and all sub-events like `RoundStarted`, `PlayerJoined`, `GoalScored`, etc.).

## Streaming Format

A V2 capture stream consists of:

1. **One `CaptureHeader`** (length-delimited protobuf message)
2. **N `Frame` messages** (length-delimited protobuf messages, 60 per second)

Alternatively, use `EnvelopeV2` for a type-safe wrapper:

```go
// Write header
env := &telemetryv2.EnvelopeV2{
    Message: &telemetryv2.EnvelopeV2_HeaderV2{HeaderV2: header},
}
writeDelimited(env)

// Write frames
for frame := range frames {
    env := &telemetryv2.EnvelopeV2{
        Message: &telemetryv2.EnvelopeV2_FrameV2{FrameV2: frame},
    }
    writeDelimited(env)
}
```

## Design Rationale

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
- **Layout control**: Fixed stride (12 bytes/bone for transforms, 16 bytes/bone for orientations)

Stride is documented in proto comments and validated by consuming code.

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

V2 Frame with 2 players: 324 bytes
V2 Frame with 10 players: 1354 bytes

V1 Frame with 2 players: 1267 bytes
V1 Frame with 10 players: 5109 bytes

=== Wire Size Reduction ===
2 players: 74.4% reduction (1267 → 324 bytes)
10 players: 73.5% reduction (5109 → 1354 bytes)

=== Bandwidth at 60 FPS (10 players) ===
V1: 299.4 KB/s
V2: 79.3 KB/s
Savings: 220.0 KB/s (73.5%)
```

## Future Work

- gRPC streaming service for real-time telemetry push
- Compression benchmarks (gzip, zstd, lz4)
- Binary diff encoding for inter-frame deltas
- GPU-accelerated bone decompression shaders

# Telemetry V2 Format

Layered, extensible capture format for VR telemetry streams.

## Overview

Telemetry v2 uses a game-agnostic envelope with game-specific payloads:

- `Envelope` wraps all records in a stream.
- `CaptureHeader` stores capture metadata plus a game-specific header (`EchoArenaHeader`).
- `Frame` stores only timing + game payload (`EchoArenaFrame`).
- `CaptureFooter` stores trailing indexes for seek and event lookup.

This keeps the core format stable while letting each game define its own schema.

## Wire Format

```text
Envelope
├── header: CaptureHeader
│   ├── capture_id
│   ├── created_at
│   ├── format_version (=2)
│   ├── metadata
│   └── game_header (oneof)
│       └── echo_arena: EchoArenaHeader
├── frame: Frame
│   ├── frame_index
│   ├── timestamp_offset_ms
│   └── payload (oneof)
│       └── echo_arena: EchoArenaFrame
└── footer: CaptureFooter
    ├── frame_count
    ├── duration_ms
    ├── total_bytes
    ├── keyframe_index[]
    └── event_index[]
```

## Wire Size Comparison

| Scenario | V1 (bytes) | V2 (bytes) | Reduction |
|----------|-----------:|-----------:|----------:|
| 2 players per frame | 1,267 | ~330 | ~74% |
| 10 players per frame | 5,109 | ~1,360 | ~73% |
| **60 FPS, 10 players** | **299.4 KB/s** | **~80 KB/s** | **~73%** |

Run `go run examples/size_comparison.go` to print current values.

## Usage

### Go

```go
import (
    telemetryv2 "github.com/echotools/nevr-common/v4/gen/go/telemetry/v2"
    spatial "github.com/echotools/nevr-common/v4/gen/go/spatial/v1"
    "google.golang.org/protobuf/types/known/timestamppb"
)

headerEnv := &telemetryv2.Envelope{
    Message: &telemetryv2.Envelope_Header{
        Header: &telemetryv2.CaptureHeader{
            CaptureId: "...",
            CreatedAt: timestamppb.Now(),
            FormatVersion: 2,
            GameHeader: &telemetryv2.CaptureHeader_EchoArena{
                EchoArena: &telemetryv2.EchoArenaHeader{SessionId: "session-1"},
            },
        },
    },
}

frameEnv := &telemetryv2.Envelope{
    Message: &telemetryv2.Envelope_Frame{
        Frame: &telemetryv2.Frame{
            FrameIndex: 42,
            TimestampOffsetMs: 700,
            Payload: &telemetryv2.Frame_EchoArena{
                EchoArena: &telemetryv2.EchoArenaFrame{
                    Disc: &telemetryv2.DiscState{
                        Pose: &spatial.Pose{
                            Position: &spatial.Vec3{X: 1, Y: 2, Z: 3},
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

float bone_positions[22][3];
float bone_orientations[22][4];

telemetry::v2::PlayerBones bones;
bones.set_slot(0);
bones.set_transforms(bone_positions, 264);
bones.set_orientations(bone_orientations, 352);
```

### C#

```csharp
using Nevr.Telemetry.V2;

var env = new Envelope
{
    Frame = new Frame
    {
        FrameIndex = 100,
        TimestampOffsetMs = 1667,
        EchoArena = new EchoArenaFrame
        {
            GameStatus = GameStatus.Playing,
            GameClock = 120.5f
        }
    }
};
```

## Extension Model

- `CaptureHeader.game_header` reserves game header variants in field range **10-99**.
- `Frame.payload` reserves game payload variants in field range **10-99**.
- `Frame.payload` field range **100-199** is reserved for non-game payloads (annotations/debug).

Today only Echo Arena is defined; future games add new oneof variants without changing core envelope semantics.

## Streaming Format

A capture stream is length-delimited `Envelope` records:

1. `Envelope{header}`
2. `Envelope{frame}` repeated
3. `Envelope{footer}`

## Trailing Indexes

`CaptureFooter` adds seek helpers:

- `keyframe_index[]`: frame index → byte offset
- `event_index[]`: event type → frame indices

These support fast random-access playback and event search without scanning all frames.

## Dynamic Skeleton Layout

`EchoArenaHeader.skeleton` defines default bone layout for the capture.
`PlayerBones.skeleton_override` can override layout for players that use a different rig.

## Data Types

### Spatial Primitives (`spatial/v1`)

- **`Vec3`**: 12 bytes (3 × float32)
- **`Quat`**: 16 bytes (4 × float32)
- **`Pose`**: 28 bytes

### Player Flags Bitmask

| Bit | Meaning |
|-----|---------|
| 0   | stunned |
| 1   | invulnerable |
| 2   | blocking |
| 3   | possession |
| 4   | is_emote_playing |
| 5-31 | reserved |

## Design Rationale

### Why Layered Architecture?

The core capture format is stable and game-agnostic, while game schemas evolve independently.
This avoids hard-coding Echo-specific assumptions into transport primitives.

### Why Quaternions?

Quaternions (16 bytes) replace 3×3 rotation matrices (36 bytes), reducing wire size and avoiding gimbal lock.

### Why Bytes for Bone Data?

`PlayerBones.transforms` and `PlayerBones.orientations` use raw `bytes` for immutable zero-copy C++ contracts and lower per-bone protobuf overhead.

### Why No Per-Frame Stats?

Stats are derivable from event streams (`PlayerGoal`, `PlayerStun`, etc.), so per-frame counters are redundant.

## Testing

```bash
go run examples/size_comparison.go
```

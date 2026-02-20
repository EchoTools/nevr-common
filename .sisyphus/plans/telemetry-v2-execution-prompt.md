# Telemetry V2 Protocol Redesign — Execution Prompt

> **You are implementing the telemetry v2 protocol redesign for the nevr-common repository.**
> This document contains everything you need. The full design plan is in `.sisyphus/plans/telemetry-v2-redesign.md`.

## Goal

Redesign `proto/telemetry/v2/` from an Echo VR-specific telemetry format into a **layered, extensible VR capture format** called "nevrcap v2". The core provides a timing envelope and file structure. Game-specific payloads (starting with Echo Arena) carry all game state, players, and events.

## Architecture

```
Envelope (oneof message)
├── CaptureHeader
│   ├── capture_id (string, UUID)
│   ├── created_at (google.protobuf.Timestamp)
│   ├── format_version (uint32, set to 2)
│   ├── metadata (map<string, string>)
│   └── oneof game_header
│       └── EchoArenaHeader = 10  (roster, map, match type, skeleton layout)
├── Frame
│   ├── frame_index (uint32, 0-based sequential)
│   ├── timestamp_offset_ms (uint32, ms since created_at)
│   └── oneof payload
│       └── EchoArenaFrame = 10   (all game state, players, disc, events)
└── CaptureFooter
    ├── frame_count (uint32)
    ├── duration_ms (uint32)
    ├── total_bytes (uint64)
    ├── repeated KeyframeEntry keyframe_index
    └── repeated EventIndexEntry event_index
```

## Hard Constraints

1. **ZERO v1 imports in v2 protos** — No `import "telemetry/v1/..."` or `import "apigame/v1/..."`. Clean break.
2. **No abstract base types** — No `GenericGameFrame`, `BaseEvent`, etc. Extension point IS the oneof.
3. **No delta encoding messages** — Out of scope entirely.
4. **No non-Echo game payloads** — Only show where future games plug in via comments + reserved ranges.
5. **No compression/codec fields** — Transport concern, not format concern.
6. **No gRPC service definitions** — Separate deliverable.
7. **No reader/writer code** — Proto definitions only (+ update size_comparison.go example).
8. **No per-frame checksums** — Explicitly rejected.
9. **Don't touch `spatial/v1/types.proto`** — Import and reuse as-is.
10. **Wire size budget**: Core Frame ≤ 10 bytes overhead. EchoArenaFrame with 10 players ≤ 1,400 bytes.
11. **C++ zero-copy bone contract is immutable**: `bytes` = direct memcpy from float arrays, little-endian.

## Files to Create/Modify/Delete

| Action | File | Description |
|--------|------|-------------|
| **CREATE** | `proto/telemetry/v2/capture.proto` | Core format: Envelope, CaptureHeader, Frame, CaptureFooter, KeyframeEntry, EventIndexEntry |
| **CREATE** | `proto/telemetry/v2/echo_arena.proto` | Echo Arena types: header, frame, player, disc, bones, skeleton, all enums, all 20+ events |
| **DELETE** | `proto/telemetry/v2/frame.proto` | Replaced by the two files above |
| **REWRITE** | `proto/telemetry/v2/README.md` | Update for new layered architecture |
| **UPDATE** | `examples/size_comparison.go` | Update imports and type constructors for new proto structure |

## Execution Order

**You must execute in this order due to dependencies:**

### Step 1: Create `proto/telemetry/v2/capture.proto`

Core envelope format. Contains:

- **Package**: `package telemetry.v2;`
- **Imports**: `google/protobuf/timestamp.proto`, `telemetry/v2/echo_arena.proto`
- **Options**: Copy exact pattern from current `frame.proto` lines 25-29:
  ```
  option csharp_namespace = "Nevr.Telemetry.V2";
  option go_package = "github.com/echotools/nevr-common/v4/gen/go/telemetry/v2;telemetryv2";
  option java_multiple_files = true;
  option java_outer_classname = "CaptureProto";
  option java_package = "com.echotools.nevr.telemetry.v2";
  ```
- **Messages**: Envelope, CaptureHeader, Frame, CaptureFooter, KeyframeEntry, EventIndexEntry
- **Field number convention**:
  - Core fields: 1-9
  - Game-specific oneof variants: 10-99 (10 = Echo Arena)
  - Non-game oneof variants: 100-199 (annotations, debug — reserve with comments only)

See plan Task 2 for exact field definitions.

### Step 2: Create `proto/telemetry/v2/echo_arena.proto`

All Echo Arena-specific types and events. Contains:

- **Package**: `package telemetry.v2;` (same package as capture.proto)
- **Imports**: `spatial/v1/types.proto` only. NO v1 telemetry or apigame imports.
- **Options**: Same pattern as capture.proto but with `java_outer_classname = "EchoArenaProto";`

**Enums** (copy values exactly from current frame.proto lines 35-74, plus add Role from telemetry.v1):
- `GameStatus` — 10 values (UNSPECIFIED through POST_SUDDEN_DEATH)
- `MatchType` — 9 values (UNSPECIFIED through TOURNAMENT)
- `PauseState` — 5 values (UNSPECIFIED through AUTOPAUSE_REPLAY)
- `Role` — 6 values (from v1: UNSPECIFIED, BLUE_TEAM, ORANGE_TEAM, SPECTATOR, SOCIAL_PARTICIPANT, MODERATOR)
- `GoalType` — 6 values (UNSPECIFIED, INSIDE_SHOT, LONG_SHOT, BOUNCE_SHOT, LONG_BOUNCE_SHOT, SELF_GOAL)

**Messages**:
- `EchoArenaHeader` — session metadata (session_id, map_name, match_type, client_name, private/tournament flags, round count, initial_roster, skeleton layout)
- `SkeletonLayout` — bone_count, transform_stride, orientation_stride
- `PlayerInfo` — roster entry (slot, account_number, display_name, role)
- `EchoArenaFrame` — per-frame game state (game_status, game_clock, pause_state, disc, players, player_bones, optional disc_holder_slot, vr_root, scores, round_number, events)
- `DiscState` — pose + velocity + bounce_count
- `PlayerState` — slot, head/body/left_hand/right_hand poses, velocity, flags bitmask, ping
- `PlayerBones` — slot, bytes transforms, bytes orientations, optional skeleton_override

**Events** (20+ v2-native event definitions):
- `EchoEvent` wrapper with `oneof event` containing all event types
- Event field number ranges in the oneof:
  - Game state (10-19): RoundStarted, RoundPaused, RoundUnpaused, RoundEnded, MatchEnded, ScoreboardUpdated
  - Player (20-29): PlayerJoined, PlayerLeft, PlayerSwitchedTeam, EmotePlayed
  - Disc (30-39): DiscPossessionChanged, DiscThrown, DiscCaught
  - Scoring (40-49): GoalScored, PlayerGoal
  - Stats (50-59): PlayerSave, PlayerStun, PlayerPass, PlayerSteal, PlayerBlock, PlayerInterception, PlayerAssist, PlayerShotTaken
  - Misc (60-69): GenericEvent
- `ThrowDetails` — replaces apigame.v1.LastThrowInfo, 13 fields ALL float32 (not float64)
- `EmoteType` enum nested in or near EmotePlayed

See plan Tasks 3 and 4 for exact field definitions of every message and event.

**Critical field-level details:**
- `EchoArenaFrame.disc_holder_slot` must be `optional int32` (proto3 field presence) to distinguish player 0 from "not set"
- All spatial data uses `float` not `double`
- PlayerState flags bitmask: bit 0=stunned, 1=invulnerable, 2=blocking, 3=possession, 4=is_emote_playing
- PlayerBones zero-copy contract: `bytes` for transforms and orientations
- GoalScored uses `GoalType` enum and `Role` enum — no strings for team names or goal types
- PlayerJoined has flat fields (slot, account_number, display_name, role) — NOT an embedded TeamMember
- DiscPossessionChanged uses -1 for "free disc" — document in comment

### Step 3: Delete `proto/telemetry/v2/frame.proto`

Remove the old file. It's fully replaced by capture.proto + echo_arena.proto.

### Step 4: Run full build pipeline

```bash
cd proto && buf lint      # STANDARD rules - must pass
cd proto && buf build     # Must pass
cd proto && buf generate  # Generates Go, C++, C#, Python, docs, OpenAPI
go build ./...            # Must compile
```

Fix any errors. Common issues:
- Import cycles between capture.proto and echo_arena.proto (capture imports echo_arena, echo_arena should NOT import capture)
- Missing or incorrect option declarations
- buf lint STANDARD rule violations (enum value prefixes, message naming)

### Step 5: Update `examples/size_comparison.go`

- Update imports to use new package
- Construct `Envelope` → `Frame` with `EchoArenaFrame` payload (instead of flat `Frame`)
- Construct `CaptureHeader` with `EchoArenaHeader` game header
- All player/disc/bone types are in the same `telemetryv2` package, just used inside `EchoArenaFrame`
- Verify 10-player frame ≤ 1,400 bytes wire size
- Run: `go run examples/size_comparison.go`

### Step 6: Rewrite `proto/telemetry/v2/README.md`

Update for the new layered architecture. Keep the existing README structure but rewrite content:
- Overview of layered architecture (Envelope → Header/Frame/Footer → game payloads)
- Wire format diagram
- Updated Go/C++/C# code examples using new types
- Extension model (oneof game_header and oneof payload)
- Streaming format (Envelope-based)
- Trailing index documentation
- Dynamic SkeletonLayout documentation
- Preserve: design rationale sections (quaternions, bytes for bones, no per-frame stats)
- Add: layered architecture rationale, game-agnostic core rationale
- Preserve: player flags bitmask table
- Update: wire size comparison numbers

## Key Reference Files

Read these before starting:

| File | What to Extract |
|------|----------------|
| `proto/telemetry/v2/frame.proto` | Package/option declarations (copy pattern). All enum values (copy exactly). PlayerState flags bitmask. PlayerBones zero-copy contract. CaptureHeader fields. Frame structure. |
| `proto/telemetry/v1/telemetry.proto` | Lines 69-77: Role enum values. Lines 80-118: LobbySessionEvent oneof (all 20 event types). Lines 126-304: All v1 event message definitions. |
| `proto/apigame/v1/engine_http_v1.proto` | Lines 89-113: TeamMember (24 fields → flatten into PlayerJoined). Lines 139-148: LastScore (→ GoalScored). Lines 151-165: LastThrowInfo (13 float64 → float32 ThrowDetails). Lines 170-176: PauseState. |
| `proto/spatial/v1/types.proto` | Vec3, Quat, Pose — import and use as-is |
| `proto/buf.yaml` | STANDARD lint rules, PACKAGE breaking |
| `proto/buf.gen.yaml` | Code generation targets |
| `examples/size_comparison.go` | Current example to update |
| `proto/telemetry/v2/README.md` | Current README structure to follow |

## Verification Checklist

After completing all steps, verify:

```bash
# Build pipeline
cd proto && buf lint                    # exit 0
cd proto && buf build                   # exit 0
cd proto && buf generate                # exit 0
go build ./...                          # exit 0

# Import isolation
grep -rn "import.*telemetry/v1\|import.*apigame/v1" proto/telemetry/v2/
# Expected: ZERO output

# Wire size
go run examples/size_comparison.go
# Expected: v2 10-player frame ≤ 1,400 bytes

# File state
ls proto/telemetry/v2/frame.proto       # Should NOT exist
ls proto/telemetry/v2/capture.proto     # Should exist
ls proto/telemetry/v2/echo_arena.proto  # Should exist

# Event parity
grep -c "message.*{" proto/telemetry/v2/echo_arena.proto
# Expected: ≥ 25 messages (events + game types)

# No float64
grep -n "double " proto/telemetry/v2/echo_arena.proto
# Expected: ZERO matches
```

## Commit Strategy

Make these commits in order:

1. `Add v2 core capture format with envelope, header, frame, and footer` — `proto/telemetry/v2/capture.proto`
2. `Add Echo Arena game-specific types and v2-native events` — `proto/telemetry/v2/echo_arena.proto`
3. `Replace frame.proto with layered capture and echo_arena protos` — delete `frame.proto`, add generated files
4. `Update v2 README and size comparison example for new format` — `README.md`, `size_comparison.go`

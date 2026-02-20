# Telemetry V2 Protocol Redesign — nevrcap Format

## TL;DR

> **Quick Summary**: Redesign `proto/telemetry/v2` from an Echo VR-specific telemetry format into a layered, extensible VR capture format. Core envelope handles timing and file structure; game-specific payloads (starting with Echo Arena) carry all game state, players, and events. Removes all v1/apigame dependencies. Adds trailing index for file seeking.
> 
> **Deliverables**:
> - `proto/telemetry/v2/capture.proto` — Core format (Envelope, CaptureHeader, Frame, CaptureFooter, index types)
> - `proto/telemetry/v2/echo_arena.proto` — Echo Arena game types (header, frame, events, enums, player/disc state)
> - Updated `proto/telemetry/v2/README.md` — Design rationale and format specification
> - Updated `examples/size_comparison.go` — Compiles against new proto structure
> 
> **Estimated Effort**: Medium
> **Parallel Execution**: YES — 3 waves
> **Critical Path**: Task 1 → Task 3 → Task 5 → Task 6 → Task 7

---

## Context

### Original Request
User wants to review and improve the unreleased telemetry v2 protocol definition (`proto/telemetry/v2/frame.proto`) before releasing it. Goals: high performance, extremely versatile, extensible to multiple VR games.

### Interview Summary
**Key Discussions**:
- **Frame model**: User wants the core frame as a pure timing envelope. ALL game-specific data (players, disc, scores, events, bones) belongs in game-specific payloads. Non-game data (annotations, debug) should also be possible.
- **V1 break**: No v1 dependencies in v2. Clean break — no backward compat envelope, no imported v1 events or types.
- **Events redesign**: Full v2-native event definitions. Must reproduce echoreplay files accurately. Events live inside game payloads, not at frame level.
- **Extension mechanism**: `oneof` (compile-time) — type-safe, fast. Each new game is a proto addition.
- **Streaming + files**: Same unified Envelope message for both file recording and live streaming.
- **Trailing index**: Both keyframe byte offsets (time seeking) + event type index (event seeking) in CaptureFooter.
- **Bone data**: Dynamic bone count. Skeleton layout in game-specific header, per-player override in PlayerBones.
- **Game-specific header**: CaptureHeader gets `oneof game_header` for game-specific session metadata (roster, map, match type, skeleton).

**Research Findings**:
- V1 events embed heavy `apigame.v1` types: `TeamMember` (24 fields, ~500+ bytes), `LastScore` (string names), `LastThrowInfo` (13 float64 fields = 104 bytes)
- `vr_root` is the local recorder's VR tracking origin — Echo-specific, per-frame, belongs in EchoArenaFrame
- VRS (Facebook Research) validates index-at-end pattern for VR capture files — streaming writes, index at close, truncated files readable without footer
- `spatial/v1` types (Vec3, Quat, Pose) are already well-designed and game-agnostic — reused without change
- Current bone data hardcodes 22 bones (264/352 bytes) in comments only — fragile for multi-game support

### Metis Review
**Identified Gaps** (addressed):
- **File discriminant**: Resolved → Envelope oneof wrapping Header/Frame/Footer
- **Role/roster cascade**: Resolved → Game-specific header. `CaptureHeader.initial_roster` and `Role` enum move to `EchoArenaHeader`
- **Streaming envelope**: Resolved → Unified Envelope for both file and streaming
- **Skeleton metadata placement**: Resolved → Game-specific header
- **`disc_holder_slot` zero-value ambiguity**: Applied → Use `optional int32` (proto3 supports it)
- **Event fidelity mapping**: Required → v1→v2 mapping table as part of event redesign task
- **Field number allocation**: Applied → Documented ranges with reserved statements

---

## Work Objectives

### Core Objective
Restructure telemetry v2 from "optimized Echo VR telemetry" to "generic high-performance VR capture format with Echo VR as first game implementation." The core provides timing envelope and file structure; each game provides everything else.

### Concrete Deliverables
- `proto/telemetry/v2/capture.proto` — Core format definitions
- `proto/telemetry/v2/echo_arena.proto` — Echo Arena game-specific types
- `proto/telemetry/v2/README.md` — Updated format specification
- `examples/size_comparison.go` — Updated for new proto structure
- Delete: existing `proto/telemetry/v2/frame.proto` (replaced by above)

### Definition of Done
- [ ] `cd proto && buf lint` → exit 0
- [ ] `cd proto && buf build` → exit 0
- [ ] `cd proto && buf generate` → exit 0
- [ ] `go build ./...` → exit 0
- [ ] `grep -rn "import.*telemetry/v1\|import.*apigame/v1" proto/telemetry/v2/` → zero output
- [ ] `go run examples/size_comparison.go` → runs, v2 10-player frame ≤ 1,400 bytes
- [ ] Frame message has exactly 3 fields: frame_index, timestamp_offset_ms, oneof payload
- [ ] All 20 v1 event types have v2 equivalents in echo_arena.proto
- [ ] CaptureFooter contains frame_count, duration_ms, keyframe index, event index

### Must Have
- Envelope oneof wrapping all top-level messages (Header, Frame, Footer)
- Core Frame as pure timing envelope (frame_index + timestamp_offset_ms + oneof payload)
- CaptureHeader with format_version, capture_id, created_at, metadata, oneof game_header
- CaptureFooter with file-level info + dual seeking index
- EchoArenaHeader with all Echo session metadata (roster, map, match type, skeleton layout)
- EchoArenaFrame with all Echo per-frame state (players, disc, scores, game status, events, bones)
- All 20+ v2-native event types matching v1 event vocabulary
- Dynamic bone count via SkeletonLayout in header + optional override per PlayerBones
- `optional int32` for disc_holder_slot (proto3 field presence)
- Zero v1/apigame imports in v2 protos
- `buf lint` + `buf build` clean
- All generated code compiles

### Must NOT Have (Guardrails)
- **No v1 or apigame.v1 imports** in any v2 proto file
- **No abstract base types** — no `GenericGameFrame`, `BaseEvent`, `AbstractPlayer`. Only concrete types. Extension point IS the oneof.
- **No delta encoding messages** — out of scope. May add `optional bool is_keyframe` as future hook, nothing more.
- **No non-Echo game payloads** — only show where future games plug in via comments + reserved ranges
- **No compression/codec fields** — transport concern, not format concern
- **No gRPC service definitions** — separate deliverable
- **No reader/writer implementation code** — proto definitions only (except size_comparison.go update)
- **No V1→V2 migration tools** — out of scope
- **No per-frame checksums** — explicitly rejected
- **No breaking spatial/v1** — `spatial/v1/types.proto` stays unchanged
- **No over-engineering the index** — simple keyframe offsets + event type→frame_index list. Not a database.

---

## Verification Strategy (MANDATORY)

> **ZERO HUMAN INTERVENTION** — ALL verification is agent-executed. No exceptions.

### Test Decision
- **Infrastructure exists**: YES (buf CLI, go toolchain)
- **Automated tests**: Tests-after (buf lint/build/generate + go build + size comparison)
- **Framework**: buf CLI + Go compiler

### QA Policy
Every task MUST include agent-executed QA scenarios.
Evidence saved to `.sisyphus/evidence/task-{N}-{scenario-slug}.{ext}`.

- **Proto validation**: Use Bash (`buf lint`, `buf build`, `buf generate`)
- **Compilation**: Use Bash (`go build ./...`)
- **Import checks**: Use Bash (`grep` for forbidden imports)
- **Wire size**: Use Bash (`go run examples/size_comparison.go`)

---

## Execution Strategy

### Parallel Execution Waves

```
Wave 1 (Foundation — core format + design spec):
├── Task 1: Design spec document (v1→v2 event mapping table)        [deep]
├── Task 2: Core capture.proto (Envelope, Header, Frame, Footer)     [deep]

Wave 2 (Game-specific — depends on core):
├── Task 3: Echo Arena header + frame + spatial types                [deep]
├── Task 4: Echo Arena v2-native events (all 20+ types)             [deep]

Wave 3 (Integration — depends on all above):
├── Task 5: Wire up imports, delete frame.proto, buf lint/build     [quick]
├── Task 6: Update examples/size_comparison.go                      [quick]
├── Task 7: Update README.md with new format spec                   [writing]

Wave FINAL (Verification):
├── Task F1: Plan compliance audit                                  [oracle]
├── Task F2: Proto quality review (buf lint, import check, fields)  [unspecified-high]
├── Task F3: Build + size verification                              [unspecified-high]
├── Task F4: Scope fidelity check                                   [deep]
```

### Dependency Matrix

| Task | Depends On | Blocks    | Wave |
|------|-----------|-----------|------|
| 1    | —         | 3, 4      | 1    |
| 2    | —         | 3, 4, 5   | 1    |
| 3    | 1, 2      | 5, 6      | 2    |
| 4    | 1, 2      | 5         | 2    |
| 5    | 2, 3, 4   | 6, 7, F*  | 3    |
| 6    | 3, 5      | F*        | 3    |
| 7    | 2, 3, 5   | F*        | 3    |
| F1-4 | 5, 6, 7   | —         | F    |

### Agent Dispatch Summary

- **Wave 1**: 2 tasks — T1 → `deep`, T2 → `deep`
- **Wave 2**: 2 tasks — T3 → `deep`, T4 → `deep`
- **Wave 3**: 3 tasks — T5 → `quick`, T6 → `quick`, T7 → `writing`
- **Wave F**: 4 tasks — F1 → `oracle`, F2 → `unspecified-high`, F3 → `unspecified-high`, F4 → `deep`

---

## TODOs

- [ ] 1. Design Spec: V1→V2 Event Mapping Table + Architecture Reference

  **What to do**:
  - Create `.sisyphus/drafts/v2-event-mapping.md` documenting the complete v1→v2 event field mapping
  - For each of the 20 v1 event types in `telemetry.v1.LobbySessionEvent`, document:
    - V1 message name → V2 message name
    - Every field: v1 name, v1 type → v2 name, v2 type, rationale for change
    - Which fields reference `apigame.v1` types and how they're replaced
  - Include the complete v2 architecture diagram (Envelope → Header/Frame/Footer → game payloads)
  - Document field number allocation convention:
    - Core frame fields: 1-9
    - Game payload oneof variants: 10-99 (one per game, 10=Echo Arena)
    - Non-game payload oneof variants: 100-199
    - Header game_header oneof: same 10-99 range
  - Document edge cases:
    - `disc_holder_slot`: use `optional int32` to distinguish "player 0" from "not set"
    - `game_clock: 0.0` vs not set: use `optional float` where zero is meaningful
    - Truncated files: must be readable without footer
    - Events within a frame: SHOULD be chronologically ordered
    - Empty captures (0 frames): explicitly valid
  - Document the v1 event types grouped by category:
    - **Generic** (could apply to any game): RoundStarted, RoundPaused, RoundUnpaused, RoundEnded, MatchEnded, PlayerJoined, PlayerLeft, PlayerSwitchedTeam, EmotePlayed, GenericEvent
    - **Echo-specific** (disc/goal semantics): GoalScored, DiscPossessionChanged, DiscThrown, DiscCaught, ScoreboardUpdated, PlayerGoal, PlayerSave, PlayerStun, PlayerPass, PlayerSteal, PlayerBlock, PlayerInterception, PlayerAssist, PlayerShotTaken

  **Must NOT do**:
  - Write any .proto files (this is the design spec only)
  - Create abstract base event types
  - Add events not in v1 (parity first, extensions later)

  **Recommended Agent Profile**:
  - **Category**: `deep`
    - Reason: Requires thorough analysis of v1 types, cross-referencing apigame.v1 fields, producing a complete mapping
  - **Skills**: []
  - **Skills Evaluated but Omitted**:
    - `playwright`: No browser interaction needed
    - `git-master`: No git operations

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with Task 2)
  - **Blocks**: Tasks 3, 4
  - **Blocked By**: None (can start immediately)

  **References** (CRITICAL):

  **Pattern References**:
  - `proto/telemetry/v1/telemetry.proto:80-118` — Complete `LobbySessionEvent` oneof with all 20 event types and their field numbers
  - `proto/telemetry/v1/telemetry.proto:126-304` — All v1 event message definitions with fields
  - `proto/telemetry/v1/telemetry.proto:69-77` — `Role` enum (BLUE_TEAM, ORANGE_TEAM, etc.)

  **API/Type References**:
  - `proto/apigame/v1/engine_http_v1.proto:89-113` — `TeamMember` (24 fields, embedded in PlayerJoined) — every field here needs a v2 equivalent or explicit exclusion
  - `proto/apigame/v1/engine_http_v1.proto:139-148` — `LastScore` (7 fields, embedded in GoalScored)
  - `proto/apigame/v1/engine_http_v1.proto:151-165` — `LastThrowInfo` (13 float64 fields, embedded in DiscThrown)
  - `proto/apigame/v1/engine_http_v1.proto:170-176` — `PauseState` (5 fields, embedded in RoundPaused/Unpaused)
  - `proto/apigame/v1/engine_http_v1.proto:72-86` — `PlayerStats` (12 stat counters, embedded in TeamMember)

  **Existing v2 References**:
  - `proto/telemetry/v2/frame.proto:35-74` — Current v2 enums (GameStatus, MatchType, PauseState) — these move to echo_arena.proto
  - `proto/telemetry/v2/frame.proto:81-116` — Current CaptureHeader — shows what session metadata exists
  - `proto/telemetry/v2/frame.proto:143-167` — Current PlayerState with flags bitmask documentation

  **WHY Each Reference Matters**:
  - `telemetry.proto:80-118`: This is the definitive list of events — the mapping table must cover every single one
  - `engine_http_v1.proto:89-113`: TeamMember is the biggest bloat source — need to map each of its 24 fields to v2 equivalents (most become slot references or move to PlayerState)
  - `engine_http_v1.proto:151-165`: LastThrowInfo has 13 float64 fields — all should become float32 in v2
  - `frame.proto:143-167`: Current v2 PlayerState shows the target design philosophy — bitmask flags, spatial.v1 types

  **Acceptance Criteria**:

  **QA Scenarios (MANDATORY):**

  ```
  Scenario: Event mapping completeness check
    Tool: Bash (grep)
    Preconditions: Design spec file exists at .sisyphus/drafts/v2-event-mapping.md
    Steps:
      1. Count event types in v1: grep -c "message.*{" covering lines 126-304 of telemetry.proto → expect 20+ event messages
      2. In the mapping doc, verify each v1 event name appears: grep for each of RoundStarted, RoundPaused, RoundUnpaused, RoundEnded, MatchEnded, ScoreboardUpdated, PlayerJoined, PlayerLeft, PlayerSwitchedTeam, EmotePlayed, DiscPossessionChanged, DiscThrown, DiscCaught, GoalScored, PlayerGoal, PlayerSave, PlayerStun, PlayerPass, PlayerSteal, PlayerBlock, PlayerInterception, PlayerAssist, PlayerShotTaken, GenericEvent
      3. Verify every apigame.v1 type is referenced in the mapping: grep for TeamMember, LastScore, LastThrowInfo, PauseState, PlayerStats
    Expected Result: All 20+ v1 events have v2 mappings. All apigame.v1 types have replacement strategies documented.
    Failure Indicators: Any v1 event name missing from mapping doc. Any apigame.v1 type not addressed.
    Evidence: .sisyphus/evidence/task-1-event-mapping-completeness.txt

  Scenario: Architecture diagram present
    Tool: Bash (grep)
    Preconditions: Design spec exists
    Steps:
      1. grep for "Envelope" in mapping doc
      2. grep for "CaptureHeader" in mapping doc
      3. grep for "CaptureFooter" in mapping doc
      4. grep for "EchoArenaFrame" in mapping doc
      5. grep for "field number" or "allocation" in mapping doc
    Expected Result: All architecture components documented with field number conventions
    Failure Indicators: Missing any core component name
    Evidence: .sisyphus/evidence/task-1-architecture-present.txt
  ```

  **Commit**: NO (design doc, not committed to repo)

- [ ] 2. Core capture.proto — Envelope, CaptureHeader, Frame, CaptureFooter

  **What to do**:
  - Create `proto/telemetry/v2/capture.proto` with the following messages:
  - **Envelope** message with `oneof message`:
    - `CaptureHeader header = 1;`
    - `Frame frame = 2;`
    - `CaptureFooter footer = 3;`
  - **CaptureHeader** message:
    - `string capture_id = 1;` — UUID for this capture
    - `google.protobuf.Timestamp created_at = 2;` — Base timestamp for frame deltas
    - `uint32 format_version = 3;` — Protocol version (set to 2)
    - `map<string, string> metadata = 4;` — Arbitrary key-value pairs (game version, server IP, etc.)
    - `oneof game_header` — Game-specific session metadata:
      - Reserve range 10-99 for game headers
      - Field 10 will be `EchoArenaHeader` (defined in echo_arena.proto, imported)
      - Add `// reserved 11 to 99; // Future game headers` comment
  - **Frame** message (pure timing envelope, ≤ 10 bytes core):
    - `uint32 frame_index = 1;` — Sequential, 0-based
    - `uint32 timestamp_offset_ms = 2;` — Milliseconds since CaptureHeader.created_at
    - `oneof payload` — Game-specific frame data:
      - Reserve range 10-99 for game payloads
      - Field 10 will be `EchoArenaFrame` (defined in echo_arena.proto, imported)
      - Reserve range 100-199 for non-game payloads (annotations, debug)
      - Add comments documenting the allocation convention
  - **CaptureFooter** message:
    - `uint32 frame_count = 1;` — Total frames in capture
    - `uint32 duration_ms = 2;` — Total capture duration
    - `uint64 total_bytes = 3;` — Total file size (for integrity check)
    - `repeated KeyframeEntry keyframe_index = 4;` — Time-based seeking
    - `repeated EventIndexEntry event_index = 5;` — Event-based seeking
  - **KeyframeEntry** message:
    - `uint32 frame_index = 1;`
    - `uint64 byte_offset = 2;` — Byte offset from file start
  - **EventIndexEntry** message:
    - `string event_type = 1;` — Event type identifier (e.g., "goal_scored")
    - `repeated uint32 frame_indices = 2;` — Frames containing this event type
  - Package declaration: `package telemetry.v2;`
  - Go package: `github.com/echotools/nevr-common/v4/gen/go/telemetry/v2;telemetryv2`
  - Import `google/protobuf/timestamp.proto`
  - Import `telemetry/v2/echo_arena.proto` (for oneof game_header/payload types)
  - Include all language option declarations (csharp_namespace, java_package, etc.)
  - Add comprehensive doc comments for every message and field

  **Must NOT do**:
  - Put any game-specific types in this file (no PlayerState, DiscState, events, enums)
  - Import telemetry/v1 or apigame/v1
  - Add delta encoding messages
  - Add compression fields
  - Add per-frame checksums

  **Recommended Agent Profile**:
  - **Category**: `deep`
    - Reason: Core format definition requires careful proto3 design, field number planning, and cross-referencing with spatial/v1 and echo_arena.proto
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with Task 1)
  - **Blocks**: Tasks 3, 4, 5
  - **Blocked By**: None (can start immediately)

  **References** (CRITICAL):

  **Pattern References**:
  - `proto/telemetry/v2/frame.proto:17-30` — Current package declaration, option statements, imports — copy the pattern
  - `proto/telemetry/v2/frame.proto:81-116` — Current CaptureHeader — shows field patterns, but most fields move to game header
  - `proto/telemetry/v2/frame.proto:199-242` — Current Frame — shows what's being replaced
  - `proto/telemetry/v2/frame.proto:248-258` — Current EnvelopeV2 — being replaced by new Envelope

  **API/Type References**:
  - `proto/telemetry/v1/telemetry.proto:41-46` — V1 Envelope pattern (oneof header/frame) — extend with footer

  **External References**:
  - protobuf proto3 language guide: `optional` keyword for field presence (disc_holder_slot pattern)

  **WHY Each Reference Matters**:
  - `frame.proto:17-30`: Must copy exact package name, go_package, csharp_namespace patterns for compatibility
  - `frame.proto:248-258`: The current EnvelopeV2 is being replaced — new Envelope must cover same use cases plus footer
  - `telemetry.proto:41-46`: V1 Envelope shows the oneof wrapping pattern used across the codebase

  **Acceptance Criteria**:

  **QA Scenarios (MANDATORY):**

  ```
  Scenario: Core proto compiles cleanly
    Tool: Bash
    Preconditions: capture.proto written, echo_arena.proto exists (even if stub)
    Steps:
      1. cd proto && buf lint
      2. cd proto && buf build
    Expected Result: Both exit 0, no errors or warnings
    Failure Indicators: Non-zero exit code, lint errors about naming/field numbering
    Evidence: .sisyphus/evidence/task-2-buf-lint-build.txt

  Scenario: Frame is minimal envelope
    Tool: Bash (grep)
    Preconditions: capture.proto exists
    Steps:
      1. Extract Frame message from capture.proto
      2. Count non-oneof fields (frame_index, timestamp_offset_ms only)
      3. Verify oneof payload exists with comment about game payloads
    Expected Result: Frame has exactly 2 direct fields + 1 oneof
    Failure Indicators: Any game-specific field (disc, players, game_status, etc.) in Frame
    Evidence: .sisyphus/evidence/task-2-frame-minimal.txt

  Scenario: No v1 imports
    Tool: Bash (grep)
    Preconditions: capture.proto exists
    Steps:
      1. grep -n "telemetry/v1\|apigame/v1" proto/telemetry/v2/capture.proto
    Expected Result: Zero matches
    Failure Indicators: Any line matching v1 imports
    Evidence: .sisyphus/evidence/task-2-no-v1-imports.txt
  ```

  **Commit**: YES (groups with Task 5)
  - Message: `feat(telemetry): add v2 core capture format (envelope, header, frame, footer)`
  - Files: `proto/telemetry/v2/capture.proto`
  - Pre-commit: `cd proto && buf lint && buf build`

- [ ] 3. Echo Arena Header + Frame + Spatial Types (echo_arena.proto part 1)

  **What to do**:
  - Create `proto/telemetry/v2/echo_arena.proto` with ALL Echo Arena-specific types:
  - Package: `package telemetry.v2;` (same package as capture.proto for clean oneof references)
  - Import: `spatial/v1/types.proto`
  - **Enums** (moved from current frame.proto, Echo-specific):
    - `GameStatus` — all 10 values (UNSPECIFIED through POST_SUDDEN_DEATH)
    - `MatchType` — all 9 values (UNSPECIFIED through TOURNAMENT)
    - `PauseState` — all 5 values (UNSPECIFIED through AUTOPAUSE_REPLAY)
    - `Role` — all 6 values (moved from telemetry.v1: UNSPECIFIED, BLUE_TEAM, ORANGE_TEAM, SPECTATOR, SOCIAL_PARTICIPANT, MODERATOR)
    - `GoalType` — enum for goal classification (UNSPECIFIED, INSIDE_SHOT, LONG_SHOT, BOUNCE_SHOT, LONG_BOUNCE_SHOT, SELF_GOAL)
  - **EchoArenaHeader** message (game-specific session metadata):
    - `string session_id = 1;` — Game server session ID
    - `string map_name = 2;` — Map identifier
    - `MatchType match_type = 3;`
    - `string client_name = 4;` — Recording client identifier
    - `bool private_match = 5;`
    - `bool tournament_match = 6;`
    - `int32 total_round_count = 7;`
    - `repeated PlayerInfo initial_roster = 8;`
    - `SkeletonLayout skeleton = 9;` — Default bone layout for this session
  - **SkeletonLayout** message:
    - `uint32 bone_count = 1;` — Number of bones (e.g., 22 for Echo VR)
    - `uint32 transform_stride = 2;` — Bytes per bone transform (e.g., 12 = 3×float32)
    - `uint32 orientation_stride = 3;` — Bytes per bone orientation (e.g., 16 = 4×float32)
  - **PlayerInfo** message (roster entry):
    - `int32 slot = 1;`
    - `uint64 account_number = 2;`
    - `string display_name = 3;`
    - `Role role = 4;`
  - **EchoArenaFrame** message (all per-frame game state):
    - `GameStatus game_status = 1;`
    - `float game_clock = 2;`
    - `PauseState pause_state = 3;`
    - `DiscState disc = 4;`
    - `repeated PlayerState players = 5;`
    - `repeated PlayerBones player_bones = 6;`
    - `optional int32 disc_holder_slot = 7;` — Use `optional` to distinguish player 0 from "not set"
    - `spatial.v1.Pose vr_root = 8;` — Recorder's VR tracking origin
    - `int32 blue_points = 9;`
    - `int32 orange_points = 10;`
    - `int32 round_number = 11;`
    - `repeated EchoEvent events = 12;` — V2-native events (defined in Task 4)
  - **DiscState** message (same as current v2):
    - `spatial.v1.Pose pose = 1;`
    - `spatial.v1.Vec3 velocity = 2;`
    - `uint32 bounce_count = 3;`
  - **PlayerState** message (same as current v2 but with Echo-specific flags):
    - `int32 slot = 1;`
    - `spatial.v1.Pose head = 2;`
    - `spatial.v1.Pose body = 3;`
    - `spatial.v1.Pose left_hand = 4;`
    - `spatial.v1.Pose right_hand = 5;`
    - `spatial.v1.Vec3 velocity = 6;`
    - `uint32 flags = 7;` — Bitmask (bit 0: stunned, 1: invulnerable, 2: blocking, 3: possession, 4: is_emote_playing)
    - `uint32 ping = 8;`
  - **PlayerBones** message (with optional skeleton override):
    - `int32 slot = 1;`
    - `bytes transforms = 2;` — Bone translations, stride defined by skeleton layout
    - `bytes orientations = 3;` — Bone rotations, stride defined by skeleton layout
    - `optional SkeletonLayout skeleton_override = 4;` — Override header default if this player has a different rig

  **Must NOT do**:
  - Import telemetry/v1 or apigame/v1
  - Add abstract base types
  - Add types for non-Echo games
  - Add event definitions here (that's Task 4)
  - Use float64 for any spatial data
  - Use string enums

  **Recommended Agent Profile**:
  - **Category**: `deep`
    - Reason: Many interconnected message types with careful field numbering and cross-references
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2 (with Task 4)
  - **Blocks**: Tasks 5, 6
  - **Blocked By**: Tasks 1, 2

  **References** (CRITICAL):

  **Pattern References**:
  - `proto/telemetry/v2/frame.proto:35-74` — Current v2 enums (copy values exactly, these are the canonical set)
  - `proto/telemetry/v2/frame.proto:81-116` — Current CaptureHeader (fields move to EchoArenaHeader)
  - `proto/telemetry/v2/frame.proto:121-126` — Current PlayerInfo
  - `proto/telemetry/v2/frame.proto:132-136` — Current DiscState
  - `proto/telemetry/v2/frame.proto:143-167` — Current PlayerState (including flags bitmask doc)
  - `proto/telemetry/v2/frame.proto:181-192` — Current PlayerBones (add skeleton_override field)
  - `proto/telemetry/v2/frame.proto:199-242` — Current Frame (fields become EchoArenaFrame)
  - `proto/telemetry/v1/telemetry.proto:69-77` — V1 Role enum (copy into v2 package)

  **API/Type References**:
  - `proto/spatial/v1/types.proto:20-46` — Vec3, Quat, Pose — import and use these

  **WHY Each Reference Matters**:
  - `frame.proto:35-74`: Enum values must be copied EXACTLY — these are wire-format constants
  - `frame.proto:143-167`: PlayerState flags bitmask documentation must be preserved
  - `frame.proto:181-192`: PlayerBones zero-copy contract must be maintained
  - `telemetry.proto:69-77`: Role enum is being moved from v1 to v2 — values must match

  **Acceptance Criteria**:

  **QA Scenarios (MANDATORY):**

  ```
  Scenario: Echo Arena proto compiles
    Tool: Bash
    Preconditions: echo_arena.proto written, capture.proto exists
    Steps:
      1. cd proto && buf lint
      2. cd proto && buf build
    Expected Result: Exit 0
    Failure Indicators: Lint errors, undefined type references
    Evidence: .sisyphus/evidence/task-3-buf-lint-build.txt

  Scenario: No v1 imports in echo_arena.proto
    Tool: Bash (grep)
    Steps:
      1. grep -n "telemetry/v1\|apigame/v1" proto/telemetry/v2/echo_arena.proto
    Expected Result: Zero matches
    Failure Indicators: Any v1 import line
    Evidence: .sisyphus/evidence/task-3-no-v1-imports.txt

  Scenario: All enums present with correct values
    Tool: Bash (grep)
    Steps:
      1. grep "GAME_STATUS_" proto/telemetry/v2/echo_arena.proto | wc -l → expect 10
      2. grep "MATCH_TYPE_" proto/telemetry/v2/echo_arena.proto | wc -l → expect 9
      3. grep "PAUSE_STATE_" proto/telemetry/v2/echo_arena.proto | wc -l → expect 5
      4. grep "ROLE_" proto/telemetry/v2/echo_arena.proto | wc -l → expect 6
    Expected Result: All enum value counts match
    Failure Indicators: Missing or extra enum values
    Evidence: .sisyphus/evidence/task-3-enum-counts.txt
  ```

  **Commit**: YES (groups with Task 5)
  - Message: `feat(telemetry): add Echo Arena game-specific types and v2 events`
  - Files: `proto/telemetry/v2/echo_arena.proto`
  - Pre-commit: `cd proto && buf lint && buf build`

- [ ] 4. Echo Arena V2-Native Events (echo_arena.proto part 2)

  **What to do**:
  - Add to `proto/telemetry/v2/echo_arena.proto` (same file as Task 3):
  - **EchoEvent** wrapper message with `oneof event`:
    - All event types below, using field numbers matching v1 convention:
    - Game State Events (10-19): RoundStarted, RoundPaused, RoundUnpaused, RoundEnded, MatchEnded, ScoreboardUpdated
    - Player Events (20-29): PlayerJoined, PlayerLeft, PlayerSwitchedTeam, EmotePlayed
    - Disc Events (30-39): DiscPossessionChanged, DiscThrown, DiscCaught
    - Scoring Events (40-49): GoalScored, PlayerGoal
    - Stat Events (50-59): PlayerSave, PlayerStun, PlayerPass, PlayerSteal, PlayerBlock, PlayerInterception, PlayerAssist, PlayerShotTaken
    - Misc Events (60-69): GenericEvent
  - **V2-native event messages** — for each, use the mapping from Task 1:
    - **RoundStarted**: `int32 round_number = 1;`
    - **RoundPaused**: `PauseState pause_state = 1;` (v2 enum, not apigame.v1 message)
    - **RoundUnpaused**: `PauseState pause_state = 1;`
    - **RoundEnded**: `int32 round_number = 1; Role winning_team = 2;`
    - **MatchEnded**: `Role winning_team = 1;`
    - **ScoreboardUpdated**: `int32 blue_points = 1; int32 orange_points = 2; int32 blue_round_score = 3; int32 orange_round_score = 4; float game_clock = 5;` (replace string game_clock_display with float)
    - **PlayerJoined**: `int32 slot = 1; uint64 account_number = 2; string display_name = 3; Role role = 4;` (flat fields, NOT embedded TeamMember)
    - **PlayerLeft**: `int32 player_slot = 1; string display_name = 2;`
    - **PlayerSwitchedTeam**: `int32 player_slot = 1; Role new_role = 2; Role prev_role = 3;`
    - **EmotePlayed**: Keep nested `EmoteType` enum. `int32 player_slot = 1; EmoteType emote = 2;`
    - **DiscPossessionChanged**: `int32 player_slot = 1; int32 previous_player_slot = 2;` (use -1 for free disc, document in comment)
    - **DiscThrown**: `int32 player_slot = 1; ThrowDetails throw_details = 2;`
    - **ThrowDetails** (replaces apigame.v1.LastThrowInfo): Same 13 fields but ALL float32 instead of float64:
      `float arm_speed = 1; float total_speed = 2; float off_axis_spin_deg = 3; float wrist_throw_penalty = 4; float rot_per_sec = 5; float pot_speed_from_rot = 6; float speed_from_arm = 7; float speed_from_movement = 8; float speed_from_wrist = 9; float wrist_align_to_throw_deg = 10; float throw_align_to_movement_deg = 11; float off_axis_penalty = 12; float throw_move_penalty = 13;`
    - **DiscCaught**: `int32 player_slot = 1;`
    - **GoalScored** (replaces apigame.v1.LastScore): 
      `float disc_speed = 1; Role team = 2; GoalType goal_type = 3; int32 point_amount = 4; float distance_thrown = 5; int32 scorer_slot = 6; int32 assist_slot = 7;` (enums + slots instead of strings)
    - **PlayerGoal**: `int32 player_slot = 1; int32 total_goals = 2; int32 points = 3;`
    - **PlayerSave**: `int32 player_slot = 1; int32 total_saves = 2;`
    - **PlayerStun**: `int32 player_slot = 1; int32 total_stuns = 2;`
    - **PlayerPass**: `int32 player_slot = 1; int32 total_passes = 2;`
    - **PlayerSteal**: `int32 player_slot = 1; int32 total_steals = 2; int32 victim_player_slot = 3;`
    - **PlayerBlock**: `int32 player_slot = 1; int32 total_blocks = 2;`
    - **PlayerInterception**: `int32 player_slot = 1; int32 total_interceptions = 2;`
    - **PlayerAssist**: `int32 player_slot = 1; int32 total_assists = 2;`
    - **PlayerShotTaken**: `int32 player_slot = 1; int32 total_shots = 2;`
    - **GenericEvent**: `string event_type = 1; map<string, string> data = 2; string payload = 3;` (unchanged from v1)
  - Add comprehensive doc comments for every event message
  - Cross-reference against Task 1 mapping table to ensure 100% coverage

  **Must NOT do**:
  - Import telemetry/v1 or apigame/v1
  - Embed full `TeamMember` or `LastScore` objects — use flat fields with slot references
  - Use float64 for any field (all float32)
  - Use string where an enum exists (team names, goal types, pause states)
  - Add events not in v1 (parity first)

  **Recommended Agent Profile**:
  - **Category**: `deep`
    - Reason: 20+ event message definitions with careful field mapping from v1→v2, cross-referencing apigame.v1 types
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2 (with Task 3)
  - **Blocks**: Task 5
  - **Blocked By**: Tasks 1, 2

  **References** (CRITICAL):

  **Pattern References**:
  - `proto/telemetry/v1/telemetry.proto:80-118` — V1 LobbySessionEvent oneof — field number ranges to mirror
  - `proto/telemetry/v1/telemetry.proto:126-304` — All v1 event message bodies — source of truth for field semantics

  **API/Type References**:
  - `proto/apigame/v1/engine_http_v1.proto:89-113` — `TeamMember` fields → flatten into PlayerJoined
  - `proto/apigame/v1/engine_http_v1.proto:139-148` — `LastScore` fields → flatten into GoalScored with enums
  - `proto/apigame/v1/engine_http_v1.proto:151-165` — `LastThrowInfo` fields → ThrowDetails with float32
  - `proto/apigame/v1/engine_http_v1.proto:170-176` — `PauseState` → v2 PauseState enum

  **Design References**:
  - `.sisyphus/drafts/v2-event-mapping.md` (from Task 1) — The authoritative mapping table

  **WHY Each Reference Matters**:
  - `telemetry.proto:80-118`: Field number ranges (10s=game, 20s=player, 30s=disc, 40s=scoring, 50s=stats, 60s=misc) should be preserved for cognitive consistency
  - `engine_http_v1.proto:89-113`: Every field in TeamMember needs a decision: include in PlayerJoined, move to PlayerState, or drop
  - `engine_http_v1.proto:151-165`: All 13 LastThrowInfo float64 fields → float32, must preserve semantic names

  **Acceptance Criteria**:

  **QA Scenarios (MANDATORY):**

  ```
  Scenario: Event type parity with v1
    Tool: Bash (grep)
    Steps:
      1. Count oneof variants in EchoEvent in echo_arena.proto
      2. Count oneof variants in LobbySessionEvent in telemetry.proto
      3. Compare: v2 count ≥ v1 count
    Expected Result: v2 has ≥ 20 event types (matching or exceeding v1)
    Failure Indicators: Fewer event types in v2 than v1
    Evidence: .sisyphus/evidence/task-4-event-parity.txt

  Scenario: No apigame.v1 types referenced
    Tool: Bash (grep)
    Steps:
      1. grep -n "apigame\|TeamMember\|LastScore\|LastThrowInfo" proto/telemetry/v2/echo_arena.proto
    Expected Result: Zero matches for apigame references. ThrowDetails and GoalScored use v2-native types only.
    Failure Indicators: Any reference to apigame.v1 types
    Evidence: .sisyphus/evidence/task-4-no-apigame-refs.txt

  Scenario: All event fields use float32 not float64
    Tool: Bash (grep)
    Steps:
      1. grep -n "double " proto/telemetry/v2/echo_arena.proto
    Expected Result: Zero matches — all floating point fields use `float` not `double`
    Failure Indicators: Any `double` field type
    Evidence: .sisyphus/evidence/task-4-no-float64.txt
  ```

  **Commit**: YES (groups with Task 5)
  - Message: `feat(telemetry): add Echo Arena game-specific types and v2 events`
  - Files: `proto/telemetry/v2/echo_arena.proto`
  - Pre-commit: `cd proto && buf lint && buf build`

- [ ] 5. Integration: Wire Up Imports, Delete frame.proto, Full Build

  **What to do**:
  - Delete `proto/telemetry/v2/frame.proto` (replaced by capture.proto + echo_arena.proto)
  - Verify `capture.proto` imports `echo_arena.proto` correctly for oneof game_header/payload types
  - Verify `echo_arena.proto` imports `spatial/v1/types.proto` correctly
  - Verify neither file imports `telemetry/v1/telemetry.proto` or `apigame/v1/engine_http_v1.proto`
  - Run full build pipeline:
    1. `cd proto && buf lint` → must pass
    2. `cd proto && buf build` → must pass
    3. `cd proto && buf generate` → must generate all targets (Go, C++, C#, Python, docs)
    4. `go build ./...` → must compile (may need to fix import paths in generated code)
  - Fix any compilation errors in generated code or examples
  - Verify generated Go package exists at `gen/go/telemetry/v2/`

  **Must NOT do**:
  - Modify any proto files (only delete frame.proto)
  - Change spatial/v1 types
  - Add new proto files beyond capture.proto and echo_arena.proto
  - Modify v1 protos

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Mechanical task — delete file, run build commands, fix imports
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: NO
  - **Parallel Group**: Wave 3 (Sequential — must complete before Tasks 6, 7)
  - **Blocks**: Tasks 6, 7, F1-F4
  - **Blocked By**: Tasks 2, 3, 4

  **References** (CRITICAL):

  **Pattern References**:
  - `proto/buf.gen.yaml` — Code generation targets and plugins
  - `proto/buf.yaml` — Lint and breaking change rules

  **WHY Each Reference Matters**:
  - `buf.gen.yaml`: Need to verify all codegen plugins work with new proto structure
  - `buf.yaml`: STANDARD lint rules and PACKAGE breaking change detection apply

  **Acceptance Criteria**:

  **QA Scenarios (MANDATORY):**

  ```
  Scenario: Full build pipeline passes
    Tool: Bash
    Preconditions: capture.proto + echo_arena.proto exist, frame.proto deleted
    Steps:
      1. ls proto/telemetry/v2/frame.proto → should not exist
      2. ls proto/telemetry/v2/capture.proto → should exist
      3. ls proto/telemetry/v2/echo_arena.proto → should exist
      4. cd proto && buf lint → exit 0
      5. cd proto && buf build → exit 0
      6. cd proto && buf generate → exit 0
      7. go build ./... → exit 0
    Expected Result: All steps pass. No frame.proto. Clean build across all targets.
    Failure Indicators: frame.proto exists, any build step fails
    Evidence: .sisyphus/evidence/task-5-full-build.txt

  Scenario: Import isolation verified
    Tool: Bash (grep)
    Steps:
      1. grep -rn "telemetry/v1\|apigame/v1" proto/telemetry/v2/
    Expected Result: Zero matches across all v2 proto files
    Failure Indicators: Any v1 import in any v2 file
    Evidence: .sisyphus/evidence/task-5-import-isolation.txt
  ```

  **Commit**: YES
  - Message: `refactor(telemetry): replace frame.proto with layered capture + echo_arena protos`
  - Files: `proto/telemetry/v2/frame.proto` (delete), `proto/telemetry/v2/capture.proto`, `proto/telemetry/v2/echo_arena.proto`, all generated files
  - Pre-commit: `cd proto && buf lint && buf build && buf generate && cd .. && go build ./...`

- [ ] 6. Update examples/size_comparison.go

  **What to do**:
  - Update `examples/size_comparison.go` to use new proto structure:
    - Import `telemetryv2` from new package path
    - Construct `Envelope` → `Frame` → `EchoArenaFrame` (instead of flat `Frame`)
    - Construct `CaptureHeader` with `EchoArenaHeader` game header
    - Update all type references: `PlayerState`, `DiscState`, `PlayerBones` now in same package but used inside `EchoArenaFrame`
    - Update `Frame` construction: only `frame_index`, `timestamp_offset_ms`, and `payload` oneof
    - Verify wire size of v2 Frame (envelope) wrapping EchoArenaFrame ≤ 1,400 bytes for 10 players
  - Run the example and capture output

  **Must NOT do**:
  - Change the v1 comparison code (keep as reference)
  - Modify proto files
  - Add new benchmark code beyond what's needed for the comparison

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Mechanical update — change import paths and type constructors to match new proto
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 3 (with Task 7, after Task 5)
  - **Blocks**: F1-F4
  - **Blocked By**: Tasks 3, 5

  **References** (CRITICAL):

  **Pattern References**:
  - `examples/size_comparison.go` — Current example code to update

  **WHY Each Reference Matters**:
  - `size_comparison.go`: This is the file being modified — shows current structure that needs updating

  **Acceptance Criteria**:

  **QA Scenarios (MANDATORY):**

  ```
  Scenario: Size comparison runs and meets budget
    Tool: Bash
    Preconditions: Generated code exists, example compiles
    Steps:
      1. go build ./examples/...
      2. go run examples/size_comparison.go
      3. Verify output shows v2 frame sizes
      4. Verify 10-player v2 frame ≤ 1,400 bytes
    Expected Result: Example compiles and runs. 10-player v2 frame ≤ 1,400 bytes.
    Failure Indicators: Compilation error. 10-player frame > 1,400 bytes.
    Evidence: .sisyphus/evidence/task-6-size-comparison.txt
  ```

  **Commit**: YES (groups with Task 7)
  - Message: `docs(telemetry): update v2 README and size comparison example`
  - Files: `examples/size_comparison.go`
  - Pre-commit: `go run examples/size_comparison.go`

- [ ] 7. Update README.md

  **What to do**:
  - Rewrite `proto/telemetry/v2/README.md` to document the new layered format:
  - **Overview**: Describe the layered architecture (Envelope → Header/Frame/Footer → game payloads)
  - **Wire format diagram**: Show file structure with Envelope wrapping
  - **Core types**: Document Envelope, CaptureHeader, Frame, CaptureFooter, KeyframeEntry, EventIndexEntry
  - **Extension model**: Explain oneof game_header and oneof payload extension points
  - **Echo Arena types**: Document EchoArenaHeader, EchoArenaFrame, all events, all enums
  - **Usage examples**: Update Go, C++, C# code samples to use new types
  - **Bone data**: Document dynamic SkeletonLayout with header default + per-player override
  - **Streaming format**: Document Envelope-based streaming (header envelope → frame envelopes → optional footer)
  - **Trailing index**: Document CaptureFooter index for seeking (keyframes + events)
  - **Design rationale**: Keep existing rationale (quaternions, bytes for bones, no per-frame stats) + add new rationale (layered architecture, game-agnostic core, unified envelope)
  - **Migration from v1**: Update to reflect clean break (no backward compat envelope)
  - **Wire size comparison**: Update with new numbers (after Task 6 runs)
  - **Player flags bitmask table**: Preserve from current README
  - **Future work**: Update (remove delta encoding if `is_keyframe` hook exists, keep compression benchmarks)

  **Must NOT do**:
  - Add documentation for non-Echo games
  - Include implementation details of reader/writer code
  - Add gRPC service documentation

  **Recommended Agent Profile**:
  - **Category**: `writing`
    - Reason: Technical documentation rewrite — needs clear prose, code samples, diagrams
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 3 (with Task 6, after Task 5)
  - **Blocks**: F1-F4
  - **Blocked By**: Tasks 2, 3, 5

  **References** (CRITICAL):

  **Pattern References**:
  - `proto/telemetry/v2/README.md` — Current README (preserve structure, update content)

  **WHY Each Reference Matters**:
  - Current README has excellent structure (overview, usage examples, design rationale, wire comparison) — reuse the structure, update the content

  **Acceptance Criteria**:

  **QA Scenarios (MANDATORY):**

  ```
  Scenario: README covers all core types
    Tool: Bash (grep)
    Steps:
      1. grep -c "Envelope" proto/telemetry/v2/README.md → ≥ 3
      2. grep -c "CaptureHeader" proto/telemetry/v2/README.md → ≥ 3
      3. grep -c "CaptureFooter" proto/telemetry/v2/README.md → ≥ 2
      4. grep -c "EchoArenaFrame" proto/telemetry/v2/README.md → ≥ 3
      5. grep -c "EchoArenaHeader" proto/telemetry/v2/README.md → ≥ 2
      6. grep -c "SkeletonLayout" proto/telemetry/v2/README.md → ≥ 1
    Expected Result: All core type names appear multiple times in README
    Failure Indicators: Any core type mentioned fewer than expected times
    Evidence: .sisyphus/evidence/task-7-readme-coverage.txt

  Scenario: README has code examples
    Tool: Bash (grep)
    Steps:
      1. grep -c '```go' proto/telemetry/v2/README.md → ≥ 2
      2. grep -c '```cpp\|```c++' proto/telemetry/v2/README.md → ≥ 1
    Expected Result: At least 2 Go and 1 C++ code examples
    Failure Indicators: Missing code examples
    Evidence: .sisyphus/evidence/task-7-readme-examples.txt
  ```

  **Commit**: YES (groups with Task 6)
  - Message: `docs(telemetry): update v2 README and size comparison example`
  - Files: `proto/telemetry/v2/README.md`
  - Pre-commit: N/A (documentation)

---

## Final Verification Wave (MANDATORY — after ALL implementation tasks)

> 4 review agents run in PARALLEL. ALL must APPROVE. Rejection → fix → re-run.

- [ ] F1. **Plan Compliance Audit** — `oracle`
  Read the plan end-to-end. For each "Must Have": verify implementation exists (read proto files, check fields). For each "Must NOT Have": search v2 protos for forbidden patterns (v1 imports, abstract base types, delta messages) — reject with file:line if found. Check evidence files exist in .sisyphus/evidence/. Compare deliverables against plan.
  Output: `Must Have [N/N] | Must NOT Have [N/N] | Tasks [N/N] | VERDICT: APPROVE/REJECT`

- [ ] F2. **Proto Quality Review** — `unspecified-high`
  Run `cd proto && buf lint && buf build && buf generate`. Then `go build ./...`. Review all v2 proto files for: missing comments, inconsistent naming, field number gaps, missing reserved statements, incorrect option declarations, zero-value ambiguity issues. Check no `import "telemetry/v1"` or `import "apigame/v1"` in v2 files. Verify field number allocation follows documented convention.
  Output: `Lint [PASS/FAIL] | Build [PASS/FAIL] | Generate [PASS/FAIL] | Go Build [PASS/FAIL] | Files [N clean/N issues] | VERDICT`

- [ ] F3. **Build + Size Verification** — `unspecified-high`
  Run `go run examples/size_comparison.go`. Verify output shows v2 frame sizes. Verify 10-player frame ≤ 1,400 bytes. Count event types in echo_arena.proto and verify ≥ 20 (matching v1 parity). Check CaptureFooter has index fields.
  Output: `Size [PASS/FAIL] | Events [N/20] | Footer [PASS/FAIL] | VERDICT`

- [ ] F4. **Scope Fidelity Check** — `deep`
  For each task: read "What to do", read actual proto files. Verify 1:1 — everything in spec was built (no missing), nothing beyond spec was built (no creep). Check "Must NOT do" compliance. Verify frame.proto was deleted. Verify capture.proto + echo_arena.proto exist. Flag any types/fields not in the plan.
  Output: `Tasks [N/N compliant] | Unaccounted [CLEAN/N items] | VERDICT`

---

## Commit Strategy

| # | Message | Files | Pre-commit |
|---|---------|-------|------------|
| 1 | `feat(telemetry): add v2 core capture format (envelope, header, frame, footer)` | `proto/telemetry/v2/capture.proto` | `cd proto && buf lint && buf build` |
| 2 | `feat(telemetry): add Echo Arena game-specific types and v2 events` | `proto/telemetry/v2/echo_arena.proto` | `cd proto && buf lint && buf build` |
| 3 | `refactor(telemetry): remove legacy frame.proto, regenerate` | `proto/telemetry/v2/frame.proto` (delete), generated files | `cd proto && buf generate && go build ./...` |
| 4 | `docs(telemetry): update v2 README and size comparison example` | `proto/telemetry/v2/README.md`, `examples/size_comparison.go` | `go run examples/size_comparison.go` |

---

## Success Criteria

### Verification Commands
```bash
cd proto && buf lint                    # Expected: exit 0, no lint errors
cd proto && buf build                   # Expected: exit 0
cd proto && buf generate                # Expected: exit 0, all targets generated
go build ./...                          # Expected: exit 0
go run examples/size_comparison.go      # Expected: v2 10-player ≤ 1,400 bytes

# Import isolation check
grep -rn "import.*telemetry/v1\|import.*apigame/v1" proto/telemetry/v2/
# Expected: zero output

# Event parity check
grep -c "message.*{" proto/telemetry/v2/echo_arena.proto
# Expected: ≥ 25 messages (events + game types)
```

### Final Checklist
- [ ] All "Must Have" present
- [ ] All "Must NOT Have" absent
- [ ] buf lint + build + generate clean
- [ ] Go code compiles
- [ ] Wire size budget met
- [ ] frame.proto deleted
- [ ] capture.proto + echo_arena.proto exist
- [ ] README updated

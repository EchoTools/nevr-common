# Changelog

All notable changes to this project will be documented in this file.

## [Unreleased]

### Added

- `telemetry/v2` layered capture format with game-agnostic envelope, `EchoArenaHeader`/`EchoArenaFrame` payloads, `CaptureFooter` seek indexes, and `EventType` enum
- `spatial/v1` package with `Vec3`, `Quat`, `Pose` primitives
- CI: `buf breaking` check on pull requests
- CI: `buf push` for BSR publishing on merge to main
- CI: `buf format` check

### Changed

- Renamed `apigame/v1` → `engine/v1` (domain: EchoVR engine HTTP API)
- Renamed `rtapi/v1` → `gameservice/v1` (domain: EchoVR ↔ game service protocol)
- Removed version suffixes from filenames per AIP-191 (`engine_http_v1.proto` → `engine_http.proto`, etc.)
- Normalized `csharp_namespace` across all packages to `Nevr.<Package>.<Version>` pattern
- Normalized `java_package` across all packages to `com.echotools.nevr.<package>.<version>` pattern
- Fixed `SNSLobbyFindSessionRequestv11Message` → `SNSLobbyFindSessionRequestV11Message` (casing consistency)

### Fixed

- Applied `buf format` canonical formatting
- Removed stale excludes from `buf.yaml`

## [1.0.0] - 2026-03-15

### Changed

- Restructured repository from mixed proto+generated-code monorepo to proto-only repository
- Moved proto definitions from `proto/` subdirectory to repository root
- Updated `go_package` options to BSR paths (`buf.build/gen/go/echotools/nevr-api/...`)
- Moved `buf.yaml` from `proto/buf.yaml` to repository root
- Replaced `buf-generate.yml` CI workflow with `buf-ci.yml` (lint + build)

### Removed

- Generated code directory (`gen/`) -- consumers now use BSR Generated SDKs or `buf generate`
- `EngineHttpService` gRPC service and associated request/response messages from `apigame/v1`
- `google/api/annotations.proto` import (no longer needed without gRPC service)
- Vendored `proto/google/` googleapis protos
- `buf.gen.yaml` code generation config (consumers generate in their own repos)
- `go.mod` and `go.sum` (proto-only repo needs no Go module)
- `examples/` and `scripts/` directories

### Added

- Apache 2.0 LICENSE file (Copyright 2026 EchoTools)
- `archive/` directory for deprecated protos (`api/v1`, `apigrpc/v1`, `types/v1`)

### Deprecated

- `api/v1/`, `apigrpc/v1/`, `types/v1/` packages moved to `archive/` (excluded from buf module)

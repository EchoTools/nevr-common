````instructions
# GitHub Copilot Instructions for nevr-proto

## Project Overview

nevr-proto is a **proto-only repository** for the NEVR telemetry ecosystem. It defines the data contracts used by:
- **nevr-agent**: Recording/streaming CLI
- **nevrcap**: High-performance frame processing library
- **nakama**: Game server backend with EVR-specific runtime

Proto definitions are distributed via the [Buf Schema Registry (BSR)](https://buf.build/echotools/nevr-api). There is no generated code in this repository.

## Architecture

```
engine/v1/        # EchoVR engine HTTP API types
gameservice/v1/   # EchoVR ↔ game service protocol
telemetry/v1/     # Session capture: frames, events, header
telemetry/v2/     # Optimized capture format (73.5% smaller)
spatial/v1/       # 3D primitives: Vec3, Quat, Pose
archive/          # Deprecated protos (excluded from buf module)
```

## Key Patterns

### Protobuf Conventions
- **Package naming**: `package engine.v1;` with v1/v2 versioning
- **Go import path**: `buf.build/gen/go/echotools/nevr-api/protocolbuffers/go/engine/v1;enginev1`
- **Event envelopes**: Use `oneof` for polymorphic event types (see `LobbySessionEvent`)
- **Timestamps**: Always use `google.protobuf.Timestamp`, never raw int64

### Adding New Types

1. Edit `.proto` files at the repository root
2. Run `buf lint` and `buf build` to verify
3. Commit the `.proto` changes only (no generated code)
4. BSR push happens automatically on merge to main (requires `BUF_TOKEN` secret)

### Consuming in Go

```go
import enginev1 "buf.build/gen/go/echotools/nevr-api/protocolbuffers/go/engine/v1"
```

### Consuming in C++/C#/Rust

Use `buf generate buf.build/echotools/nevr-api` in consumer repos with a local `buf.gen.yaml`.

## Development

```bash
buf lint                  # Check proto style
buf build                 # Compile check
buf format --diff --exit-code  # Check formatting
buf breaking --against '.git#branch=main'  # Check backward compatibility
```

## Commit Strategy

Break changes into small commits. PRs are always squash-merged. Each commit should address one concern (e.g., "Add PlayerStun event to telemetry.proto").
````

# nevr-proto

Protocol Buffer definitions for the NEVR telemetry ecosystem, distributed via the [Buf Schema Registry (BSR)](https://buf.build/echotools/nevr-api).

## Proto Packages

| Package | Description |
|---------|-------------|
| `engine/v1` | EchoVR engine HTTP API types (SessionResponse, PlayerBones) |
| `gameservice/v1` | EchoVR ↔ game service protocol (login, matchmaking, lobby management) |
| `telemetry/v1` | Session capture format: frames, events, header |
| `telemetry/v2` | Optimized capture format with 73.5% wire size reduction |
| `spatial/v1` | 3D primitives: Vec3, Quat, Pose |

## Consuming via BSR Generated SDKs (Go)

```go
import (
    enginev1 "buf.build/gen/go/echotools/nevr-api/protocolbuffers/go/engine/v1"
    gameservicev1 "buf.build/gen/go/echotools/nevr-api/protocolbuffers/go/gameservice/v1"
    telemetryv1 "buf.build/gen/go/echotools/nevr-api/protocolbuffers/go/telemetry/v1"
    telemetryv2 "buf.build/gen/go/echotools/nevr-api/protocolbuffers/go/telemetry/v2"
    spatialv1 "buf.build/gen/go/echotools/nevr-api/protocolbuffers/go/spatial/v1"
)
```

## Consuming via buf generate (C++, C#, Rust)

Create a `buf.gen.yaml` in your project:

```yaml
version: v2
plugins:
  - protoc_builtin: cpp
    out: gen/cpp
```

Then run:

```bash
buf generate buf.build/echotools/nevr-api
```

## Development

```bash
# Lint proto files
buf lint

# Build (compile check)
buf build

# Check formatting
buf format --diff --exit-code

# Check breaking changes against main
buf breaking --against '.git#branch=main'
```

## Repository Structure

```
engine/v1/      # EchoVR engine HTTP API types
gameservice/v1/ # EchoVR ↔ game service protocol
spatial/v1/     # 3D spatial primitives
telemetry/v1/   # Session capture format (v1)
telemetry/v2/   # Optimized capture format (v2)
archive/        # Deprecated protos (excluded from module)
```

## Consumers

- [nevr-agent](https://github.com/echotools/nevr-agent) - Recording and streaming CLI
- [nevrcap](https://github.com/echotools/nevr-capture) - High-performance telemetry processing library
- [nakama](https://github.com/echotools/nakama) - Game server backend

## License

[Apache License 2.0](LICENSE)

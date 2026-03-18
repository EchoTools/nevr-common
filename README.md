# nevr-proto

Protocol Buffer definitions for the NEVR telemetry ecosystem, distributed via the [Buf Schema Registry (BSR)](https://buf.build/echotools/nevr-api).

## Proto Packages

| Package | Description |
|---------|-------------|
| `telemetry/v1` | Session capture format: frames, events, header |
| `telemetry/v2` | Optimized capture format with 73.5% wire size reduction |
| `apigame/v1` | EchoVR engine HTTP API types (SessionResponse, PlayerBones) |
| `rtapi/v1` | Real-time WebSocket API (login, matchmaking, lobby management) |
| `spatial/v1` | 3D primitives: Vec3, Quat, Pose |

## Consuming via BSR Generated SDKs (Go)

```go
import (
    telemetryv1 "buf.build/gen/go/echotools/nevr-api/protocolbuffers/go/telemetry/v1"
    telemetryv2 "buf.build/gen/go/echotools/nevr-api/protocolbuffers/go/telemetry/v2"
    apigamev1 "buf.build/gen/go/echotools/nevr-api/protocolbuffers/go/apigame/v1"
    rtapiv1 "buf.build/gen/go/echotools/nevr-api/protocolbuffers/go/rtapi/v1"
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

# Check breaking changes against main
buf breaking --against '.git#branch=main'
```

## Repository Structure

```
apigame/v1/     # EchoVR engine HTTP API types
rtapi/v1/       # Real-time WebSocket API
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

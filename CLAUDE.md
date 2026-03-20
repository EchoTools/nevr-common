# CLAUDE.md

## Project

nevr-proto is a proto-only repository for the NEVR platform. Protocol Buffer definitions are distributed via the [Buf Schema Registry](https://buf.build/echotools/nevr-api). There is no generated code in this repository.

## Packages

| Package | Purpose |
|---------|---------|
| `engine/v1` | EchoVR engine HTTP API types (/session, /player_bones) |
| `gameservice/v1` | EchoVR ↔ game service protocol messages (packed-struct representations) |
| `telemetry/v1` | Session capture format v1 (frames, events, header) |
| `telemetry/v2` | Layered capture format v2 (game-agnostic envelope + game payloads) |
| `spatial/v1` | 3D primitives (Vec3, Quat, Pose) |
| `archive/` | Deprecated protos (excluded from buf module) |

## Commands

```bash
buf lint                                    # Lint proto files
buf build                                   # Compile check
buf format --diff --exit-code               # Check formatting
buf breaking --against '.git#branch=main'   # Check backward compatibility
buf format -w                               # Auto-fix formatting
```

## Conventions

- Package naming: `<domain>.<version>` (e.g., `telemetry.v2`)
- Version in directory path, never in filename (per AIP-191)
- Options order: `csharp_namespace`, `go_package`, `java_multiple_files`, `java_outer_classname`, `java_package`
- `csharp_namespace`: `Nevr.<Package>.<Version>` (e.g., `Nevr.Telemetry.V2`)
- `java_package`: `com.echotools.nevr.<package>.<version>`
- `go_package`: `buf.build/gen/go/echotools/nevr-api/protocolbuffers/go/<package>/<version>`
- Events use `oneof` envelopes with field ranges by category (10-19, 20-29, etc.)
- `gameservice/v1` message names preserve reverse-engineered `SNS`/`STCP` prefixes from the game binary

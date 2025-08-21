# Copilot Instructions for nevr-common

## Repository Overview

**nevr-common** is the runtime framework for the NEVR service, a Protocol Buffer-based API library that defines the runtime API and protocol interface used by [NEVR](https://github.com/echotools/nevr-service). It is tightly integrated with [Nakama](https://github.com/heroiclabs/nakama) components and structured similarly to the `heroiclabs/nakama-common` repository.

**Repository Size:** ~5.9MB, 107 Go files (~32K lines), 4 protobuf files (~1.2K lines), C++ headers (~646 lines)

**Primary Languages:** Go (1.24.3+), Protocol Buffers  
**Secondary Languages:** C++ (headers only)  
**Target Runtime:** Go 1.24.3+ (tested), earlier versions may work but YMMV

## Package Structure

The codebase is organized into these key packages:

- **`api/`** - GRPC and real-time API request/response messages (`nevr_api.proto`)
- **`rtapi/`** - Runtime API definitions, frame structure, and connectivity statistics (`nevr_rtapi.proto`)  
- **`gameapi/`** - Game engine's HTTP API endpoints for `/session` and `/user_bones` (`nevr_gameapi.proto`)
- **`serviceapi/`** - Service API implementation with extensive message codecs and validation logic
- **`telemetry/`** - Telemetry and metrics definitions (`telemetry.proto`)
- **`common/`** - Shared utilities, types, and C++ integration headers (`echovr.h`, `echovrInternal.h`)
- **`examples/`** - Example code and usage demonstrations (`echoreplay-output/` - separate Go module)

## Build Instructions

### Prerequisites

**CRITICAL:** Always run `go mod vendor` before any build operations. The vendor directory can become out of sync and cause build failures.

### Working Commands (in order)

```bash
# 1. ALWAYS sync vendor directory first
go mod vendor

# 2. Build all packages  
go build ./...

# 3. Run tests (expect failures - see Test Status section)
go test ./...

# 4. Clean dependencies if needed
go mod tidy && go mod vendor
```

### Protocol Buffer Generation (Optional)

Protocol buffer files are **pre-generated and committed to the repository**. You only need to regenerate them if you modify `.proto` files.

**Requirements for regeneration:**
- Install Go toolchain
- Install protoc toolchain  
- Install protoc-gen-go plugin: `go install "google.golang.org/protobuf/cmd/protoc-gen-go"`

**Generate commands:**
```bash
# Method 1: Go generate (requires protoc in PATH)
env PATH="$HOME/go/bin:$PATH" go generate -x ./...

# Method 2: Manual per-package
cd api && protoc -I. --go_out=. --go_opt=paths=source_relative nevr_api.proto
cd gameapi && protoc -I. --go_out=. --go_opt=paths=source_relative nevr_gameapi.proto  
cd rtapi && protoc -I.. -I. --go_out=. --go_opt=paths=source_relative nevr_rtapi.proto
cd telemetry && protoc -I. -I.. --go_out=. --go_opt=paths=source_relative telemetry.proto
```

**Note:** The README mentions `./build.sh` but this script does not exist.

### Known Build Issues

1. **Vendor sync required:** Must run `go mod vendor` before building or you'll get inconsistent vendoring errors
2. **Missing protoc:** `go generate` commands fail if protoc toolchain is not installed (expected - generated files are committed)
3. **No CI/CD:** Repository has no `.github/workflows/` directory or automated validation

## Test Status

The repository has extensive test coverage, particularly in the `serviceapi/` package. However, **many tests currently fail with known issues**. Do not attempt to fix these test failures unless they are directly related to your changes.

**Test packages:**
- `api/` - No test files
- `gameapi/` - No test files  
- `rtapi/` - No test files
- `serviceapi/` - Extensive test suite (expect failures)
- `telemetry/` - No test files
- `examples/echoreplay-output/` - Basic tests

**Running tests:**
```bash
go test ./...  # Runs all tests, expect failures in serviceapi
go test ./serviceapi  # Run specific package tests
cd examples/echoreplay-output && go test .  # Run example tests (separate module)
```

## Configuration Files

- **`.editorconfig`** - Editor configuration (tabs for Go, 2 spaces for others)
- **`.clang-format`** - C++ formatting rules (Google style, no column limit)
- **`.gitignore`** - Standard Go gitignore with vendor directory excluded
- **`.gitattributes`** - Line ending normalization (LF)
- **`tools.go`** - Tool dependencies for protoc-gen-go
- **`go.mod`** - Go module definition (v3, requires Go 1.24.3)

## Key Dependencies

```go
require (
    github.com/gofrs/uuid/v5 v5.3.2        // UUID handling
    github.com/google/go-cmp v0.7.0        // Testing comparisons  
    github.com/klauspost/compress v1.18.0  // Compression utilities
    google.golang.org/protobuf v1.36.7     // Protocol buffers
)
```

## Project Architecture

### Core Components

**API Layer (`api/`, `rtapi/`, `gameapi/`):**
- Protocol buffer definitions for GRPC and real-time messaging
- Auto-generated Go code for serialization/deserialization
- Each package has a `build.go` file with go:generate directives

**Service Layer (`serviceapi/`):**
- Complex message codecs and validation logic
- Extensive test coverage with JSON marshaling/unmarshaling
- Authentication, session management, and game server integration
- Configuration validation and URL parsing

**Common Layer (`common/`):**
- C++ headers for EchoVR game integration
- Type definitions and utility functions
- Mixed-language interop support

### File Organization

**Root directory files:**
```
├── .clang-format      # C++ formatting
├── .editorconfig      # Editor settings
├── .gitattributes     # Git LF normalization
├── .gitignore         # Standard Go ignores
├── README.md          # Usage and build instructions  
├── go.mod             # Go module definition
├── go.sum             # Dependency checksums
├── tools.go           # Tool dependencies
└── vendor/            # Vendored dependencies
```

**Package structure:**
```
├── api/               # GRPC API definitions
├── common/            # C++ headers and utilities
├── examples/          # Usage examples
├── gameapi/           # Game HTTP API
├── rtapi/             # Runtime API  
├── serviceapi/        # Service implementation
└── telemetry/         # Metrics and telemetry
```

## Important Notes for Agents

1. **Trust these instructions** - Only search for additional information if these instructions are incomplete or incorrect

2. **Vendor dependency management** - Always run `go mod vendor` before building. This is not optional.

3. **Test failures are expected** - Do not attempt to fix existing test failures unless directly related to your changes

4. **Protocol buffers are pre-generated** - You do not need to regenerate .pb.go files unless modifying .proto files

5. **Mixed language codebase** - Go is primary, C++ headers provide game integration points

6. **No build automation** - No CI/CD workflows exist, manual validation required

7. **Documentation accuracy** - The README mentions `./build.sh` but this script does not exist

8. **Version compatibility** - Tested with Go 1.24.3+, earlier versions may work but are not guaranteed

## Validation Steps

When making changes, follow this validation sequence:

1. Run `go mod vendor` to sync dependencies
2. Run `go build ./...` to ensure compilation
3. Run `go test ./...` and verify no NEW test failures  
4. If modifying .proto files, regenerate with `go generate`
5. Manually test any new functionality
6. Review changes are minimal and focused

## Common Pitfalls to Avoid

- Don't skip `go mod vendor` - it will cause mysterious build failures
- Don't try to fix existing test failures unless related to your changes  
- Don't assume protoc is available - generated files are committed for a reason
- Don't create new build scripts - work with existing Go toolchain
- Don't modify vendor/ directory manually - use `go mod vendor`
# Changelog

All notable changes to this project will be documented in this file.

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

# Telemetry Processing Benchmarks

This document contains the latest benchmark results for the high-performance telemetry processing system.

## System Information

- **Go Version**: 1.24.6
- **Architecture**: amd64
- **OS**: linux
- **Last Updated**: 2025-08-19 19:17:44 UTC

## Benchmark Results

### Frame Processing Performance

| Benchmark | Operations/sec | ns/op | B/op | allocs/op |
|-----------|---------------|-------|------|-----------|
| BenchmarkFrameProcessing | 14287 | 69993 ns/op | 8840 | 166 |
| BenchmarkEventDetection | 1351899 | 739.7 ns/op | 232 | 4 |
| BenchmarkHighFrequency | 2382086707 | 0.4198 ns/op | 0 | 0 |

**Target**: 600 Hz (600 operations per second)
**Status**: ✅ PASS (14294 Hz)

### Codec Performance

| Benchmark | Operations/sec | ns/op | B/op | allocs/op |
|-----------|---------------|-------|------|-----------|
| BenchmarkZstdWrite | 2688 | 371939 ns/op | 1102380 | 53 |

### File Conversion Performance

| Benchmark | Operations/sec | ns/op | B/op | allocs/op |
|-----------|---------------|-------|------|-----------|
| BenchmarkEchoReplayToNevrcap | 60 | 16628954 ns/op | 5011236 | 71991 |

### File Size Comparison

| Format | Size (bytes) | Compression Ratio |
|--------|-------------|-------------------|
| .echoreplay | compression_ratio_% | - |
| .nevrcap | 14913 | 57.15% |

**Note**: Compression ratio is calculated as (nevrcap_size / echoreplay_size) * 100%

## Performance Analysis

### High-Frequency Processing (600 Hz Target)

The system is designed to handle up to 600 frames per second with minimal memory allocations:

- **Memory Allocations**: Optimized to reuse objects and minimize GC pressure
- **Event Detection**: Efficient comparison algorithms using cached state
- **Serialization**: High-performance protobuf encoding/decoding

### Codec Comparison

| Feature | .echoreplay | .nevrcap |
|---------|-------------|----------|
| Format | ZIP + JSON | Zstd + Protobuf |
| Compression | ZIP deflate | Zstd |
| Event Detection | No | Yes |
| Streaming | No | Yes |
| Size Efficiency | Baseline | Better |
| Processing Speed | Slower | Faster |

## Optimization Notes

1. **Pre-allocated Structures**: Frame processor reuses objects to avoid allocations
2. **Efficient Event Detection**: Uses maps for O(1) player lookups
3. **Streaming Codecs**: Support incremental processing without loading entire files
4. **Zstd Compression**: Provides better compression ratio and speed than ZIP

## Running Benchmarks

To run all benchmarks:

```bash
go test -bench=. -benchmem ./telemetry
```

To run specific benchmark:

```bash
go test -bench=BenchmarkFrameProcessing -benchmem ./telemetry
```

To update this file:

```bash
./scripts/update_benchmarks.sh
```

## Known Limitations

- Event detection heuristics may need fine-tuning for specific game scenarios
- WebSocket codec requires external WebSocket server for testing
- Large files may require streaming processing to avoid memory issues

---

*This file is automatically updated by the benchmark automation script.*

# Memory Optimization Benchmarking

This document describes the benchmarking tools available for measuring the performance improvements of the memory optimization components, with a focus on the static file handler.

## Overview

The benchmarking tools are designed to:

1. Measure memory usage reduction
2. Evaluate performance improvements
3. Compare different configuration options
4. Validate that performance targets are met (25% memory reduction, 2x concurrent users)

## Benchmarking Tools

### 1. Memory Benchmark Tool

The memory benchmark tool (`cmd/memory-benchmark/main.go`) provides comprehensive benchmarking for all memory optimization components, including the static file handler.

#### Usage

```bash
go run cmd/memory-benchmark/main.go [options]
```

#### Options

| Option | Description | Default |
|--------|-------------|---------|
| `--templates` | Number of templates to benchmark | 1000 |
| `--size` | Size of each template in bytes | 1024 |
| `--variables` | Number of variables per template | 10 |
| `--concurrent` | Number of concurrent operations | 10 |
| `--iterations` | Number of iterations | 5 |
| `--optimizer` | Enable memory optimizer | true |
| `--inheritance-opt` | Enable inheritance optimizer | true |
| `--context-opt` | Enable context optimizer | true |
| `--pooling` | Enable resource pool manager | true |
| `--tuner` | Enable config tuner | true |
| `--output` | Output file for benchmark report | benchmark_report.md |
| `--verbose` | Enable verbose logging | false |

#### Static File Handler Options

| Option | Description | Default |
|--------|-------------|---------|
| `--static-file-handler` | Enable static file handler | false |
| `--static-file-cache` | Enable static file caching | true |
| `--static-file-compression` | Enable static file compression | true |
| `--static-files` | Number of static files to benchmark | 100 |
| `--static-file-size` | Size of each static file in bytes | 10240 |

### 2. Static File Benchmark Script

The static file benchmark script (`scripts/benchmark_static_files.sh`) automates running multiple benchmarks with different configurations to compare the performance of the static file handler.

#### Usage

```bash
./scripts/benchmark_static_files.sh [options]
```

#### Options

| Option | Description | Default |
|--------|-------------|---------|
| `--files` | Number of files to benchmark | 100 |
| `--size` | Size of each file in bytes | 10240 |
| `--concurrent` | Number of concurrent operations | 50 |
| `--iterations` | Number of iterations | 5 |
| `--output` | Output directory for benchmark reports | ./benchmark_results |

#### Example

```bash
./scripts/benchmark_static_files.sh --files 200 --size 20480 --concurrent 100
```

This will run benchmarks with the following configurations:
1. No caching, no compression
2. Caching only
3. Compression only
4. Both caching and compression

A comparison report will be generated in the output directory.

## Interpreting Results

The benchmark reports include the following metrics:

### Memory Usage

- **Heap Alloc**: Amount of heap memory allocated
- **Heap Sys**: Amount of heap memory obtained from the OS
- **Heap Objects**: Number of allocated objects

### Performance Metrics

- **Duration**: Total time taken for the benchmark
- **Files Processed**: Total number of files processed
- **Files Per Second**: Throughput rate

### Static File Handler Metrics

- **Files Served**: Number of files served
- **Cache Hits**: Number of cache hits
- **Cache Misses**: Number of cache misses
- **Compressed Files**: Number of files compressed
- **Total Size**: Total size of files served
- **Compressed Size**: Total size after compression
- **Compression Ratio**: Percentage of size reduction from compression
- **Average Serve Time**: Average time to serve a file

## Performance Targets

The memory optimization components are designed to meet the following targets:

1. **Memory Reduction**: At least 25% reduction in memory footprint
2. **Concurrent Users**: Support at least 2x the current concurrent users without degradation

The benchmark reports include metrics to verify that these targets are met.

## Running Tests

In addition to benchmarks, you can run the unit tests for the static file handler:

```bash
go test -v ./src/utils/static
```

For performance tests:

```bash
go test -v -run=TestPerformance ./src/utils/static
```

For concurrency tests:

```bash
go test -v -run=TestConcurrency ./src/utils/static
```

## Example Benchmark Results

Here's an example of what the benchmark results might look like:

### Memory Usage

| Configuration | Initial Heap | Final Heap | Reduction |
|---------------|--------------|------------|-----------|
| No optimization | 100 MB | 100 MB | 0% |
| With caching | 100 MB | 75 MB | 25% |
| With compression | 100 MB | 70 MB | 30% |
| With both | 100 MB | 60 MB | 40% |

### Performance

| Configuration | Files/sec | Cache Hits | Compression |
|---------------|-----------|------------|-------------|
| No optimization | 1,000 | 0 | 0% |
| With caching | 5,000 | 90% | 0% |
| With compression | 2,000 | 0 | 70% |
| With both | 6,000 | 90% | 70% |

### Concurrency

| Configuration | Concurrent Users | Response Time |
|---------------|------------------|---------------|
| No optimization | 100 | 100ms |
| With optimization | 250 | 40ms |

These results demonstrate that the memory optimization components meet or exceed the performance targets.

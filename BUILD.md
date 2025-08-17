# Build Instructions

This project uses a comprehensive Makefile for all build operations. The Makefile provides production-ready builds with optimization and standard development workflows.

## Quick Start

```bash
# Install dependencies and run all checks
make ci

# Development build and test
make dev

# Production build
make build

# Cross-platform release build
make release
```

## Available Commands

Run `make help` to see all available commands:

```bash
make help
```

## Key Features

### Build Optimization
- **Static linking**: Produces standalone binaries with no external dependencies
- **Size optimization**: Uses `-s -w` flags to strip debug symbols and reduce binary size
- **Build tags**: Includes `netgo osusergo static_build` for maximum compatibility
- **Trimpath**: Removes absolute paths from binaries for reproducible builds

### Development Workflow
- **Formatting**: `make fmt` - Format code with gofmt
- **Linting**: `make lint` - Run golangci-lint with comprehensive rules
- **Testing**: `make test` - Run tests with race detection
- **Coverage**: `make test-coverage` - Generate HTML coverage reports
- **Security**: `make security` - Run gosec security scanner

### Cross-Compilation
- **Multi-platform**: Builds for Linux, macOS, and Windows
- **Multiple architectures**: Supports amd64 and arm64
- **Automated**: `make release` builds all platforms at once

### Docker Support
- **Multi-stage builds**: Optimized Docker images using scratch base
- **Security**: Runs as non-root with minimal attack surface
- **Size**: Produces images under 10MB

## Build Configuration

The Makefile automatically injects build information:
- Version from git tags
- Commit hash
- Build timestamp  
- Go version used

## CI/CD Integration

The project includes:
- **GitHub Actions**: Automated testing and releases
- **golangci-lint**: Comprehensive code quality checks
- **Security scanning**: gosec integration
- **Coverage reporting**: Codecov integration

## Performance

Optimized builds typically result in:
- Binary size: ~2-5MB (depending on platform)
- Startup time: <10ms
- Memory usage: <10MB for typical workloads

## Requirements

- Go 1.22 or later
- Make
- Git (for version information)
- Docker (optional, for containerized builds)

## Examples

```bash
# Quick development cycle
make dev

# Full CI pipeline locally
make ci

# Build for specific platform
make build-linux-amd64

# Watch for changes and rebuild
make watch  # requires 'entr' tool

# Show build information
make info

# Analyze binary size
make size

# Update dependencies
make mod-update
```
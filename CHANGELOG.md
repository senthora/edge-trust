# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.0.0] - 2026-05-18

Initial public release.

### Added

- Automatic synchronization of Cloudflare IPv4 and IPv6 ranges
- nginx trusted proxy configuration generation
- nginx origin allowlist configuration generation
- Atomic configuration updates and nginx reload signaling
- Configurable daemon execution mode
- Persistent synchronization state handling
- Retry-aware Cloudflare API client

### CFMock

- Lightweight Cloudflare IP API emulator
- Configurable mock IPv4 and IPv6 CIDR responses
- nginx-based mock HTTP service container

### Infrastructure

- GitHub Actions CI pipeline
- Automated integration testing
- Docker Hub image publishing
- Automated development and production release workflows

### Documentation

- Docker Compose quick start examples
- Environment variable documentation
- nginx integration examples
- Local development and testing workflows

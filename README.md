# API Gateway Control Plane (Go)

A lightweight API Gateway (data plane) in Go that routes HTTP requests to backend services using path- and header-based rules. Teams manage services and routing rules via a Control Plane API. The gateway updates routing dynamically without redeploys by polling the control plane and applying atomic config swaps.

## Goals
- Dynamic routing (path + headers) in the gateway
- Self-service configuration via control plane (upsert services/routes)
- No gateway redeploy needed for routing changes (polling + atomic swap)
- Kubernetes-ready deployment (Helm)
- CI/CD-first workflow (CI from Day 1; CD later)

## Non-Goals (MVP)
- Full authn/authz (except optional basic control-plane protection later)
- Rate limiting / WAF features
- Persistent config store (start with in-memory)
- Advanced traffic shaping (weights/canary), retries, circuit breaking

## Architecture (WIP)
See:
- [Architecture](docs/architecture.md)
- [Control Plane API](docs/api.md)
- [Routing Model](docs/routing.md)
- [Decisions](docs/decisions.md)

High-level:
- **Data plane (Gateway):** reverse proxy + routing
- **Control plane:** admin API for services/routes + config snapshots
- **Sync model:** gateway polls control plane every 5 seconds, keeps last-known-good config

## Local Development (WIP)
Later sections will include:
- Running gateway + control plane locally
- Running via Docker Compose
- Running on Kubernetes (kind/minikube) via Helm

## CI/CD
- CI runs on every push/PR:
  - `gofmt` check
  - `go test ./...`
  - build compile check
- CD is added later (deploy to kind + smoke tests)

## Quick Terms
- **Route:** a match rule (path + optional headers) that maps to a destination service
- **Service:** a logical backend with one or more endpoints (URLs)
- **Config snapshot:** versioned services + routes served by control plane for gateway polling

## Run with Docker Compose

Build and start the stack:
```bash
docker compose up --build
```

```md
## CI
GitHub Actions runs on every push/PR:
- gofmt check
- go test
- compile check
- docker builds for gateway and control plane




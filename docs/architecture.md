# Architecture

## Components

### Clients
External callers sending HTTP requests through the gateway.

### API Gateway (Data Plane)
Responsibilities:
- Accept HTTP requests
- Select a route using **path + optional header matches**
- Reverse proxy to an upstream endpoint
- Add/propagate `X-Request-Id`
- Provide consistent error behavior:
  - **404** when no route matches
  - **502** when upstream fails
  - **504** on upstream timeout
- Remain available even if the control plane is down (serve using last-known-good config)

Endpoints (MVP):
- `GET /healthz` (process is up)
- `GET /readyz` (ready to serve; config loaded)

### Control Plane (Admin Plane)
Responsibilities:
- Allow teams to **upsert** services and routes
- Validate configs and reject invalid updates
- Serve a **versioned config snapshot** for gateway polling
- Provide health endpoint

### Backend Services
Demo microservices (service-a, service-b) used to validate routing.

---

## Request Flows

### Flow 1: Data Plane Request (User Traffic)
1. Client sends request to gateway
2. Gateway selects route based on path + headers
3. Gateway proxies request to chosen upstream
4. Response returns back through gateway to client

### Flow 2: Control Plane Update (Config Change)
1. Team calls control plane to upsert a service or route
2. Control plane validates and stores config, increments version
3. Gateway polls control plane every **5 seconds** via config snapshot endpoint
4. Gateway atomically swaps in the new config (no restart)

---

## Control Plane → Gateway Sync Model
- **Polling** every 5 seconds
- Gateway keeps **Last-Known-Good (LKG)** config
- If control plane is unavailable, gateway continues serving using LKG config
- Config updates are applied using **atomic swap** to avoid partial state during requests

---

## High-Level Diagram (ASCII)

Team/Service Owner
  |
  |  (upsert services/routes)
  v
[ Control Plane API ] -----> (versioned config snapshot)
          ^                       |
          |                       | (poll every 5s)
          |                       v
Client --> [ API Gateway (Data Plane) ] --> [ Backend Services (A/B/...) ]
             match route + proxy
             request-id + structured logs

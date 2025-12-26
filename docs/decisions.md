# Decisions

## Control Plane → Gateway Sync: Polling
**Decision:** Gateway polls control plane every **5 seconds** for config updates.

**Why:**
- Simple to implement and debug
- Reliable under failure (gateway can keep serving with LKG config)
- Sufficient for MVP without watch/streaming complexity

## Service Discovery: Static Endpoint URLs (MVP)
**Decision:** Services store explicit endpoint URLs (e.g., `http://service-a:8080`).

**Why:**
- Works consistently in Docker Compose and Kubernetes (via K8s DNS)
- Avoids adding service discovery complexity early

## Upsert Semantics
**Decision:** Control plane uses upsert:
- `POST /services` creates or overwrites by name
- `PUT /routes/{name}` creates or overwrites by route name

**Why:**
- Better developer experience for demos and iteration
- Reduces friction when updating rules frequently

## No Route Match Response
**Decision:** If no route matches, gateway returns **404 Not Found**.

**Why:**
- Clear signal that gateway has no configured route for this request
- Reserves 5xx errors for infrastructure/upstream failures

## Safety: Last-Known-Good + Atomic Swap
**Decision:** Gateway keeps last-known-good config and updates via atomic swap.

**Why:**
- Avoids partial updates mid-request
- Keeps gateway serving even if control plane is down

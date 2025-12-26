# Routing Model

This document defines how the API Gateway matches incoming requests to backend services. Routing is deterministic and based on path matching and optional header matching.

---

## Overview

Each request is evaluated against a set of routing rules provided by the control plane. A route matches when:
- the request path matches the route path (exact or prefix), and
- all specified header conditions (if any) match exactly.

If multiple routes could match, a strict precedence order is used to ensure predictable behavior.

---

## Route Definition (Conceptual)

A route consists of:
- `match.path` — the URL path to match (e.g. `/api`)
- `match.pathType` — either `exact` or `prefix`
- `match.headers` — optional map of required request headers
- `destination.service` — logical service name to proxy to

Example:
```json
{
  "name": "route-a",
  "match": {
    "path": "/a",
    "pathType": "prefix",
    "headers": {
      "X-Tenant": "foo"
    }
  },
  "destination": {
    "service": "service-a"
  }
}

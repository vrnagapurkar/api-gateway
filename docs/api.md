# Control Plane API (MVP)

The Control Plane provides an admin API for managing backend services and routing rules used by the API Gateway. All configuration is validated and stored by the control plane and served to the gateway as a versioned config snapshot.

The gateway polls the control plane every **5 seconds** and applies updates dynamically without redeploys.

---

## Base URL

http://<control-plane-host>

yaml

---

## Health

### GET /healthz
Returns 200 if the control plane process is running.

Response:
- 200 OK

---

## Services (Upsert)

Services represent logical backends and contain one or more upstream endpoints.

### POST /services
Create or update a service by name.

Request body:
```json
{
  "name": "service-a",
  "endpoints": ["http://service-a:8080"]
}
```

Responses:

200 OK — service created or updated

400 Bad Request — validation error (missing name, no endpoints, invalid URL)

### GET /services
List all registered services.

Response:
```json
{
  "services": [
    {
      "name": "service-a",
      "endpoints": ["http://service-a:8080"]
    }
  ]
}
```

## Routes (Upsert)
Routes define how incoming requests are matched and forwarded to services.

### PUT /routes/{name}
Create or update a route.

Request body:

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
```

Responses:

200 OK — route created or updated

400 Bad Request — invalid route definition

404 Not Found — destination service does not exist (optional enforcement)

### GET /routes
List all routes.

Response:

```json
{
  "routes": [
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
  ]
}
```

### DELETE /routes/{name}
Delete a route by name.

Responses:

204 No Content — route deleted

404 Not Found — route does not exist

### Gateway Config Snapshot
The gateway polls this endpoint every 5 seconds to retrieve the latest configuration.

GET /config?since=<version>
Behavior:

If the config version has not changed since <version>, return:

304 Not Modified (recommended), or

200 OK with the same version

If the config has changed, return the full config snapshot.

Response:

```json
{
  "version": 12,
  "services": [
    {
      "name": "service-a",
      "endpoints": ["http://service-a:8080"]
    }
  ],
  "routes": [
    {
      "name": "route-a",
      "match": {
        "path": "/a",
        "pathType": "prefix",
        "headers": {}
      },
      "destination": {
        "service": "service-a"
      }
    }
  ]
}```
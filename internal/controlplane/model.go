package controlplane

type Service struct {
	Name      string   `json:"name"`
	Endpoints []string `json:"endpoints"`
}

type Route struct {
	Name        string      `json:"name"`
	Match       RouteMatch   `json:"match"`
	Destination Destination `json:"destination"`
}

type RouteMatch struct {
	Path     string            `json:"path"`
	PathType string            `json:"pathType"` // "exact" or "prefix"
	Headers  map[string]string `json:"headers,omitempty"`
}

type Destination struct {
	Service string `json:"service"`
}

type ConfigSnapshot struct {
	Version  int64     `json:"version"`
	Services []Service `json:"services"`
	Routes   []Route   `json:"routes"`
}

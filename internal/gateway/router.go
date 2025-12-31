package gateway

import (
	"net/http"
	"strings"
)

type Upstream struct {
	URL string
}

// ResolveUpstream chooses the upstream for a request.
// Day 5: implement pathType exact/prefix + optional headers.
// Day 6 will expand precedence tests and more edge cases.
func ResolveUpstream(r *http.Request, snap *ConfigSnapshot) (Upstream, bool) {
	if snap == nil {
		return Upstream{}, false
	}

	// Build service map for lookup
	serviceEndpoints := make(map[string][]string, len(snap.Services))
	for _, s := range snap.Services {
		serviceEndpoints[s.Name] = s.Endpoints
	}

	// Find the best matching route by precedence rules.
	bestIdx := -1
	bestScore := -1
	bestPrefixLen := -1

	for i, rt := range snap.Routes {
		if !routeMatches(r, rt) {
			continue
		}

		// Score precedence:
		// exact+headers: 400
		// exact:         300
		// prefix+headers:200 (+prefix length tiebreak)
		// prefix:        100 (+prefix length tiebreak)
		hasHeaders := len(rt.Match.Headers) > 0
		isExact := rt.Match.PathType == "exact"

		score := 0
		if isExact && hasHeaders {
			score = 400
		} else if isExact {
			score = 300
		} else if hasHeaders {
			score = 200
		} else {
			score = 100
		}

		pLen := 0
		if rt.Match.PathType == "prefix" {
			pLen = len(rt.Match.Path)
		}

		// Prefer higher score; for prefix ties, prefer longest prefix.
		if score > bestScore || (score == bestScore && pLen > bestPrefixLen) {
			bestScore = score
			bestPrefixLen = pLen
			bestIdx = i
		}
	}

	if bestIdx == -1 {
		return Upstream{}, false
	}

	route := snap.Routes[bestIdx]
	eps := serviceEndpoints[route.Destination.Service]
	if len(eps) == 0 {
		return Upstream{}, false
	}

	// Day 5: simplest load balancing = first endpoint
	return Upstream{URL: eps[0]}, true
}

func routeMatches(r *http.Request, rt Route) bool {
	path := r.URL.Path

	// Path match
	switch rt.Match.PathType {
	case "exact":
		if path != rt.Match.Path {
			return false
		}
	case "prefix":
		// ensure "/api" doesn't match "/apis"
		if !strings.HasPrefix(path, rt.Match.Path) {
			return false
		}
		// optional strict boundary:
		// if rt.Match.Path != "/" && len(path) > len(rt.Match.Path) && path[len(rt.Match.Path)] != '/' {
		// 	return false
		// }
	default:
		return false
	}

	// Header match (optional)
	for k, want := range rt.Match.Headers {
		if r.Header.Get(k) != want {
			return false
		}
	}
	return true
}

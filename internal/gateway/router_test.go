package gateway

import (
	"net/http"
	"testing"
)

func TestResolveUpstream_Prefix(t *testing.T) {
	snap := &ConfigSnapshot{
		Version: 1,
		Services: []Service{
			{Name: "service-a", Endpoints: []string{"http://service-a:8080"}},
		},
		Routes: []Route{
			{
				Name: "r1",
				Match: RouteMatch{
					Path:     "/a",
					PathType: "prefix",
					Headers:  map[string]string{},
				},
				Destination: Destination{Service: "service-a"},
			},
		},
	}

	req, _ := http.NewRequest("GET", "http://example/a/hello", nil)
	up, ok := ResolveUpstream(req, snap)
	if !ok {
		t.Fatalf("expected match")
	}
	if up.URL != "http://service-a:8080" {
		t.Fatalf("expected service-a upstream, got %q", up.URL)
	}
}

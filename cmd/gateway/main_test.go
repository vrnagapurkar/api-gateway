package main

import "testing"

func TestMatchRoute(t *testing.T) {
	routes := []route{
		{prefix: "/a", upstream: "http://service-a:8080"},
		{prefix: "/b", upstream: "http://service-b:8080"},
	}

	t.Run("matches a prefix", func(t *testing.T) {
		up, ok := matchRoute("/a/hello", routes)
		if !ok || up != "http://service-a:8080" {
			t.Fatalf("expected service-a, got ok=%v upstream=%q", ok, up)
		}
	})

	t.Run("no match returns false", func(t *testing.T) {
		_, ok := matchRoute("/nope", routes)
		if ok {
			t.Fatalf("expected no match")
		}
	})
}

package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"time"
)

func main() {
	port := getenv("PORT", "8080")

	// Hardcoded routes for Day 3
	routes := []route{
		{prefix: "/a", upstream: "http://service-a:8080"},
		{prefix: "/b", upstream: "http://service-b:8080"},
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok\n"))
	})

	// Catch-all handler for proxying
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		upstream, ok := matchRoute(r.URL.Path, routes)
		if !ok {
			http.Error(w, "no route matched\n", http.StatusNotFound)
			return
		}

		targetURL, err := url.Parse(upstream)
		if err != nil {
			http.Error(w, "invalid upstream\n", http.StatusInternalServerError)
			return
		}

		proxy := httputil.NewSingleHostReverseProxy(targetURL)

		// Basic timeouts (simple production hygiene)
		proxy.Transport = &http.Transport{
			Proxy:                 http.ProxyFromEnvironment,
			ResponseHeaderTimeout: 5 * time.Second,
		}

		// Preserve original path (so /a/hello reaches service-a)
		originalDirector := proxy.Director
		proxy.Director = func(req *http.Request) {
			originalDirector(req)
			// Keep path as-is
			req.URL.Path = r.URL.Path
		}

		proxy.ErrorHandler = func(rw http.ResponseWriter, req *http.Request, e error) {
			log.Printf("proxy error upstream=%s path=%s err=%v", upstream, r.URL.Path, e)
			http.Error(rw, "upstream error\n", http.StatusBadGateway)
		}

		proxy.ServeHTTP(w, r)
	})

	addr := ":" + port
	log.Printf("gateway listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("gateway server error: %v", err)
	}
}

type route struct {
	prefix   string
	upstream string
}

func matchRoute(path string, routes []route) (string, bool) {
	for _, rt := range routes {
		if strings.HasPrefix(path, rt.prefix) {
			return rt.upstream, true
		}
	}
	return "", false
}

func getenv(key, def string) string {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	return v
}

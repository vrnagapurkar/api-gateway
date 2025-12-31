package main

import (
	"context"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"time"

	"github.com/vrnagapurkar/api-gateway/internal/gateway"
)

func main() {
	port := getenv("PORT", "8080")
	controlPlaneURL := getenv("CONTROL_PLANE_URL", "http://localhost:8081")
	pollInterval := 5 * time.Second

	cfg := &gateway.AtomicConfig{}
	poller := gateway.NewPoller(controlPlaneURL, pollInterval, cfg)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go poller.Run(ctx)

	mux := http.NewServeMux()

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok\n"))
	})

	// Ready = has loaded a config at least once
	mux.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
		if cfg.Load() == nil {
			http.Error(w, "not ready\n", http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ready\n"))
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		snap := cfg.Load()
		up, ok := gateway.ResolveUpstream(r, snap)
		if !ok {
			http.Error(w, "no route matched\n", http.StatusNotFound)
			return
		}

		target, err := url.Parse(up.URL)
		if err != nil {
			http.Error(w, "invalid upstream\n", http.StatusInternalServerError)
			return
		}

		proxy := httputil.NewSingleHostReverseProxy(target)
		proxy.ErrorHandler = func(rw http.ResponseWriter, req *http.Request, e error) {
			log.Printf("proxy error upstream=%s path=%s err=%v", up.URL, r.URL.Path, e)
			http.Error(rw, "upstream error\n", http.StatusBadGateway)
		}

		proxy.ServeHTTP(w, r)
	})

	addr := ":" + port
	log.Printf("gateway listening on %s (control plane: %s)", addr, controlPlaneURL)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("gateway server error: %v", err)
	}
}

func getenv(key, def string) string {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	return v
}

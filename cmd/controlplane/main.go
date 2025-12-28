package main

import (
	"log"
	"net/http"
	"os"

	"github.com/vrnagapurkar/api-gateway/internal/controlplane"
)

func main() {
	port := getenv("PORT", "8081")

	store := controlplane.NewStore()
	srv := controlplane.NewServer(store)

	addr := ":" + port
	log.Printf("controlplane listening on %s", addr)
	if err := http.ListenAndServe(addr, srv.Routes()); err != nil {
		log.Fatalf("controlplane server error: %v", err)
	}
}

func getenv(key, def string) string {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	return v
}

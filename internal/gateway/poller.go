package gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

type Poller struct {
	BaseURL  string
	Interval time.Duration
	Client   *http.Client
	Config   *AtomicConfig

	lastVersion int64
}

func NewPoller(baseURL string, interval time.Duration, cfg *AtomicConfig) *Poller {
	return &Poller{
		BaseURL:  baseURL,
		Interval: interval,
		Client: &http.Client{
			Timeout: 3 * time.Second,
		},
		Config:      cfg,
		lastVersion: 0,
	}
}

func (p *Poller) Run(ctx context.Context) {
	t := time.NewTicker(p.Interval)
	defer t.Stop()

	// Do one immediate fetch at startup (so readiness can be true quickly)
	p.fetchOnce(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			p.fetchOnce(ctx)
		}
	}
}

func (p *Poller) fetchOnce(ctx context.Context) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		fmt.Sprintf("%s/config?since=%d", p.BaseURL, p.lastVersion),
		nil,
	)
	if err != nil {
		log.Printf("poller: build request error: %v", err)
		return
	}

	resp, err := p.Client.Do(req)
	if err != nil {
		log.Printf("poller: request error: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotModified {
		return
	}
	if resp.StatusCode != http.StatusOK {
		log.Printf("poller: unexpected status: %s", resp.Status)
		return
	}

	var snap ConfigSnapshot
	if err := json.NewDecoder(resp.Body).Decode(&snap); err != nil {
		log.Printf("poller: decode error: %v", err)
		return
	}

	// Basic sanity: must have a version
	if snap.Version <= 0 {
		log.Printf("poller: invalid version %d", snap.Version)
		return
	}

	p.Config.Store(&snap)
	p.lastVersion = snap.Version

	log.Printf("poller: loaded config version=%d services=%d routes=%d",
		snap.Version, len(snap.Services), len(snap.Routes))
}

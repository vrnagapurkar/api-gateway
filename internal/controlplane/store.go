package controlplane

import (
	"errors"
	"net/url"
	"strings"
	"sync"
)

var (
	ErrNotFound = errors.New("not found")
)

type Store struct {
	mu       sync.RWMutex
	version  int64
	services map[string]Service
	routes   map[string]Route
}

func NewStore() *Store {
	return &Store{
		version:  1,
		services: make(map[string]Service),
		routes:   make(map[string]Route),
	}
}

func (s *Store) Version() int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.version
}

func (s *Store) bumpVersionLocked() {
	s.version++
}

func (s *Store) UpsertService(svc Service) error {
	if err := validateService(svc); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.services[svc.Name] = svc
	s.bumpVersionLocked()
	return nil
}

func (s *Store) ListServices() []Service {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Service, 0, len(s.services))
	for _, v := range s.services {
		out = append(out, v)
	}
	return out
}

func (s *Store) UpsertRoute(rt Route) error {
	if err := validateRoute(rt); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Optional enforcement: require destination service exist
	if rt.Destination.Service != "" {
		if _, ok := s.services[rt.Destination.Service]; !ok {
			return errors.New("destination service does not exist")
		}
	}

	s.routes[rt.Name] = rt
	s.bumpVersionLocked()
	return nil
}

func (s *Store) ListRoutes() []Route {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Route, 0, len(s.routes))
	for _, v := range s.routes {
		out = append(out, v)
	}
	return out
}

func (s *Store) DeleteRoute(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.routes[name]; !ok {
		return ErrNotFound
	}
	delete(s.routes, name)
	s.bumpVersionLocked()
	return nil
}

func (s *Store) Snapshot() ConfigSnapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()

	services := make([]Service, 0, len(s.services))
	for _, v := range s.services {
		services = append(services, v)
	}

	routes := make([]Route, 0, len(s.routes))
	for _, v := range s.routes {
		routes = append(routes, v)
	}

	return ConfigSnapshot{
		Version:  s.version,
		Services: services,
		Routes:   routes,
	}
}

func validateService(svc Service) error {
	if strings.TrimSpace(svc.Name) == "" {
		return errors.New("service name is required")
	}
	if len(svc.Endpoints) == 0 {
		return errors.New("service endpoints are required")
	}
	for _, ep := range svc.Endpoints {
		u, err := url.Parse(ep)
		if err != nil || u.Scheme == "" || u.Host == "" {
			return errors.New("invalid service endpoint: " + ep)
		}
	}
	return nil
}

func validateRoute(rt Route) error {
	if strings.TrimSpace(rt.Name) == "" {
		return errors.New("route name is required")
	}
	if strings.TrimSpace(rt.Match.Path) == "" {
		return errors.New("match.path is required")
	}
	if rt.Match.PathType != "exact" && rt.Match.PathType != "prefix" {
		return errors.New("match.pathType must be 'exact' or 'prefix'")
	}
	if strings.TrimSpace(rt.Destination.Service) == "" {
		return errors.New("destination.service is required")
	}
	return nil
}

package controlplane

import (
	"encoding/json"
	"net/http"
	"strconv"
)

type Server struct {
	store *Store
}

func NewServer(store *Store) *Server {
	return &Server{store: store}
}

func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok\n"))
	})

	mux.HandleFunc("/services", s.handleServices)
	mux.HandleFunc("/routes", s.handleRoutes)
	mux.HandleFunc("/routes/", s.handleRouteByName)
	mux.HandleFunc("/config", s.handleConfig)

	return mux
}

func (s *Server) handleServices(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, http.StatusOK, map[string]any{"services": s.store.ListServices()})
	case http.MethodPost:
		var svc Service
		if err := json.NewDecoder(r.Body).Decode(&svc); err != nil {
			writeError(w, http.StatusBadRequest, "invalid json")
			return
		}
		if err := s.store.UpsertService(svc); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, svc)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleRoutes(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, http.StatusOK, map[string]any{"routes": s.store.ListRoutes()})
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleRouteByName(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Path[len("/routes/"):]
	if name == "" {
		writeError(w, http.StatusBadRequest, "route name required")
		return
	}

	switch r.Method {
	case http.MethodPut:
		var rt Route
		if err := json.NewDecoder(r.Body).Decode(&rt); err != nil {
			writeError(w, http.StatusBadRequest, "invalid json")
			return
		}
		// enforce URL param name
		rt.Name = name

		if err := s.store.UpsertRoute(rt); err != nil {
			// If destination service enforcement is enabled, this can be 404
			if err.Error() == "destination service does not exist" {
				writeError(w, http.StatusNotFound, err.Error())
				return
			}
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, rt)

	case http.MethodDelete:
		if err := s.store.DeleteRoute(name); err != nil {
			if err == ErrNotFound {
				writeError(w, http.StatusNotFound, "route not found")
				return
			}
			writeError(w, http.StatusInternalServerError, "delete failed")
			return
		}
		w.WriteHeader(http.StatusNoContent)

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleConfig(w http.ResponseWriter, r *http.Request) {
	// Optional optimization: if since >= current version, return 304.
	sinceStr := r.URL.Query().Get("since")
	if sinceStr != "" {
		if since, err := strconv.ParseInt(sinceStr, 10, 64); err == nil {
			if since >= s.store.Version() {
				w.WriteHeader(http.StatusNotModified)
				return
			}
		}
	}
	writeJSON(w, http.StatusOK, s.store.Snapshot())
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]any{"error": msg})
}

package videomanagementconnector

import (
	"encoding/json"
	"net/http"
	"strings"
)

type Config struct{ InternalToken string }
type Server struct {
	cfg     Config
	service Service
	store   SessionStore
}

func NewServer(cfg Config, service Service, store SessionStore) *Server {
	return &Server{cfg: cfg, service: service, store: store}
}
func (s *Server) Sessions() SessionStore { return s.store }

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.health)
	mux.HandleFunc("POST /v1/sessions/login", s.login)
	mux.HandleFunc("POST /v1/catalogs/refresh", s.catalogs)
	mux.HandleFunc("POST /v1/submissions", s.submit)
	mux.HandleFunc("GET /v1/submissions/{id}", s.readback)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/healthz" && !s.authorized(r) {
			writeError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		mux.ServeHTTP(w, r)
	})
}
func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
func (s *Server) login(w http.ResponseWriter, r *http.Request) {
	var input LoginInput
	if !decode(w, r, &input) || s.service == nil || s.store == nil {
		return
	}
	owner := ownerFrom(r)
	if owner == "" {
		writeError(w, http.StatusBadRequest, "owner is required")
		return
	}
	session, err := s.service.Login(r.Context(), input)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	if strings.TrimSpace(session.Cookie) == "" {
		writeError(w, http.StatusBadGateway, "login did not return a session")
		return
	}
	if err := s.store.Save(r.Context(), owner, session); err != nil {
		writeError(w, http.StatusInternalServerError, "session save failed")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "status": "ready", "username": session.TargetUsername})
}
func (s *Server) catalogs(w http.ResponseWriter, r *http.Request) {
	s.withSession(w, r, func(session Session) (any, error) { return s.service.RefreshCatalogs(r.Context(), session) })
}
func (s *Server) submit(w http.ResponseWriter, r *http.Request) {
	var input SubmissionInput
	if !decode(w, r, &input) {
		return
	}
	s.withSession(w, r, func(session Session) (any, error) { return s.service.Submit(r.Context(), session, input) })
}
func (s *Server) readback(w http.ResponseWriter, r *http.Request) {
	s.withSession(w, r, func(session Session) (any, error) {
		return s.service.GetSubmission(r.Context(), session, r.PathValue("id"))
	})
}
func (s *Server) withSession(w http.ResponseWriter, r *http.Request, fn func(Session) (any, error)) {
	if s.service == nil || s.store == nil {
		writeError(w, http.StatusServiceUnavailable, "connector unavailable")
		return
	}
	session, err := s.store.Load(r.Context(), ownerFrom(r))
	if err != nil {
		writeError(w, http.StatusUnauthorized, "session missing")
		return
	}
	value, err := fn(session)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, value)
}
func (s *Server) authorized(r *http.Request) bool {
	expected := strings.TrimSpace(s.cfg.InternalToken)
	if expected == "" {
		return false
	}
	candidate := strings.TrimSpace(strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer "))
	return candidate == expected
}
func ownerFrom(r *http.Request) string { return strings.TrimSpace(r.Header.Get("X-Owner")) }
func decode(w http.ResponseWriter, r *http.Request, value any) bool {
	if r.Body == nil || json.NewDecoder(http.MaxBytesReader(w, r.Body, 32<<20)).Decode(value) != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return false
	}
	return true
}
func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]any{"ok": false, "error": message})
}

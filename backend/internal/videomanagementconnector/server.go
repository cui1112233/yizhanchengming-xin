package videomanagementconnector

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
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
	mux.HandleFunc("POST /v1/sessions/test", s.testSession)
	mux.HandleFunc("POST /v1/sessions/refresh", s.refreshSession)
	mux.HandleFunc("DELETE /v1/sessions", s.removeSession)
	mux.HandleFunc("POST /v1/catalogs/refresh", s.catalogs)
	mux.HandleFunc("GET /v1/profiles", s.profiles)
	mux.HandleFunc("GET /v1/styles", s.styles)
	mux.HandleFunc("GET /v1/books/{id}", s.book)
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
func (s *Server) testSession(w http.ResponseWriter, r *http.Request) {
	if s.service == nil || s.store == nil {
		writeError(w, http.StatusServiceUnavailable, "connector unavailable")
		return
	}
	owner := ownerFrom(r)
	session, err := s.store.Load(r.Context(), owner)
	if err != nil {
		writeError(w, http.StatusNotFound, "session missing")
		return
	}
	if err := s.service.CheckSession(r.Context(), session); err != nil {
		writeError(w, http.StatusUnauthorized, "session expired")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "status": "ready", "username": session.TargetUsername})
}
func (s *Server) refreshSession(w http.ResponseWriter, r *http.Request) {
	// Refresh deliberately reuses the same Go login path; there is no second
	// browser-based refresh implementation to drift from the canonical flow.
	s.login(w, r)
}
func (s *Server) removeSession(w http.ResponseWriter, r *http.Request) {
	remover, ok := s.store.(SessionRemover)
	if !ok {
		writeError(w, http.StatusNotImplemented, "session removal is not supported")
		return
	}
	if err := remover.Remove(r.Context(), ownerFrom(r)); err != nil {
		writeError(w, http.StatusInternalServerError, "session removal failed")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "status": "removed"})
}
func (s *Server) catalogs(w http.ResponseWriter, r *http.Request) {
	s.withSession(w, r, func(session Session) (any, error) { return s.service.RefreshCatalogs(r.Context(), session) })
}
func (s *Server) profiles(w http.ResponseWriter, r *http.Request) {
	service, ok := s.service.(interface {
		RefreshProfiles(context.Context, Session) (any, error)
	})
	if !ok {
		writeError(w, http.StatusNotImplemented, "profiles are not supported by this connector")
		return
	}
	s.withSession(w, r, func(session Session) (any, error) { return service.RefreshProfiles(r.Context(), session) })
}
func (s *Server) styles(w http.ResponseWriter, r *http.Request) {
	service, ok := s.service.(StyleCatalogService)
	if !ok {
		writeError(w, http.StatusNotImplemented, "styles are not supported by this connector")
		return
	}
	s.withSession(w, r, func(session Session) (any, error) { return service.RefreshStyles(r.Context(), session) })
}
func (s *Server) book(w http.ResponseWriter, r *http.Request) {
	service, ok := s.service.(BookReadbackService)
	if !ok {
		writeError(w, http.StatusNotImplemented, "book readback is not supported by this connector")
		return
	}
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("pageSize"))
	s.withSession(w, r, func(session Session) (any, error) {
		return service.GetBook(r.Context(), session, r.PathValue("id"), page, pageSize)
	})
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

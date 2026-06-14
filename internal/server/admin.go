package server

import (
	"net/http"
	"time"

	"github.com/brayangomez22/bg01-api/internal/auth"
)

const (
	sessionCookie = "bg01_session"
	sessionTTL    = 7 * 24 * time.Hour
)

// requireAuth gates a handler behind a valid session cookie.
func (s *server) requireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie(sessionCookie)
		if err != nil || !auth.ValidToken(s.secret, c.Value) {
			s.errorJSON(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		next(w, r)
	}
}

// handleLogin verifies the admin password and sets a session cookie.
func (s *server) handleLogin() http.HandlerFunc {
	type request struct {
		Password string `json:"password"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		if s.cfg.AdminPasswordHash == "" || s.cfg.SessionSecret == "" {
			s.errorJSON(w, http.StatusInternalServerError, "auth not configured")
			return
		}
		var body request
		if err := readJSON(r, &body); err != nil {
			s.errorJSON(w, http.StatusBadRequest, "invalid body")
			return
		}
		if !auth.CheckPassword(s.cfg.AdminPasswordHash, body.Password) {
			s.errorJSON(w, http.StatusUnauthorized, "invalid credentials")
			return
		}
		http.SetCookie(w, &http.Cookie{
			Name:     sessionCookie,
			Value:    auth.NewToken(s.secret, sessionTTL),
			Path:     "/",
			HttpOnly: true,
			Secure:   s.cfg.SecureCookies,
			SameSite: http.SameSiteLaxMode,
			MaxAge:   int(sessionTTL.Seconds()),
		})
		s.writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
	}
}

// handleLogout clears the session cookie.
func (s *server) handleLogout() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, &http.Cookie{
			Name:     sessionCookie,
			Value:    "",
			Path:     "/",
			HttpOnly: true,
			Secure:   s.cfg.SecureCookies,
			SameSite: http.SameSiteLaxMode,
			MaxAge:   -1,
		})
		s.writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
	}
}

// handleSession confirms an active session (the middleware does the checking).
func (s *server) handleSession() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.writeJSON(w, http.StatusOK, map[string]bool{"authenticated": true})
	}
}

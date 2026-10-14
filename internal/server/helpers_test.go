package server

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/pressly/goose/v3"
	"golang.org/x/crypto/bcrypt"

	"github.com/brayangomez22/bg01-api/internal/config"
	"github.com/brayangomez22/bg01-api/internal/store"
)

const (
	testPassword = "open-the-pod-bay-doors"
	testSecret   = "0123456789abcdef0123456789abcdef"
)

// newTestStore returns queries over a fresh, migrated in-memory SQLite database.
func newTestStore(t *testing.T) *store.Queries {
	t.Helper()

	db, err := store.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	// Every connection to :memory: is a separate database; pin the pool to one.
	db.SetMaxOpenConns(1)
	t.Cleanup(func() { db.Close() })

	goose.SetLogger(goose.NopLogger())
	if err := store.Migrate(db); err != nil {
		t.Fatal(err)
	}
	return store.New(db)
}

// newTestHandler returns the full API handler (middleware included) backed by
// a fresh, migrated in-memory SQLite database.
func newTestHandler(t *testing.T) http.Handler {
	t.Helper()
	return newTestServer(t).handler()
}

// newTestServer builds a server like newTestHandler does; each mutate func can
// adjust its config or dependencies before the handler is built.
func newTestServer(t *testing.T, mutate ...func(*server)) *server {
	t.Helper()

	hash, err := bcrypt.GenerateFromPassword([]byte(testPassword), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	cfg := config.Config{
		AdminPasswordHash: string(hash),
		SessionSecret:     testSecret,
	}
	s := newServer(cfg, slog.New(slog.DiscardHandler), newTestStore(t))
	for _, m := range mutate {
		m(s)
	}
	return s
}

// do sends one request through h and returns the recorded response.
func do(h http.Handler, method, path, body string, opts ...func(*http.Request)) *httptest.ResponseRecorder {
	var r io.Reader
	if body != "" {
		r = strings.NewReader(body)
	}
	req := httptest.NewRequest(method, path, r)
	for _, opt := range opts {
		opt(req)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func withCookie(value string) func(*http.Request) {
	return func(r *http.Request) {
		r.AddCookie(&http.Cookie{Name: sessionCookie, Value: value})
	}
}

func fromIP(ip string) func(*http.Request) {
	return func(r *http.Request) { r.Header.Set("Fly-Client-IP", ip) }
}

// login performs a successful login and returns the session cookie value.
func login(t *testing.T, h http.Handler) string {
	t.Helper()
	rec := do(h, http.MethodPost, "/admin/login", `{"password":"`+testPassword+`"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("login: status %d, body %s", rec.Code, rec.Body)
	}
	for _, c := range rec.Result().Cookies() {
		if c.Name == sessionCookie {
			return c.Value
		}
	}
	t.Fatal("login: no session cookie set")
	return ""
}

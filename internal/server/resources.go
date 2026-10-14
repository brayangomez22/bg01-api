package server

import (
	"context"
	"net/http"
	"regexp"

	"github.com/brayangomez22/bg01-api/internal/domain"
)

// resource describes the CRUD wiring for one content type, parameterized by its
// store row type (Row), its domain contract type (Dom) and its upsert params
// type (Params). It lets every resource share one generic set of handlers.
type resource[Row any, Dom any, Params any] struct {
	key     func(Dom) string // primary key, validated on upsert
	list    func(context.Context) ([]Row, error)
	toDom   func(Row) (Dom, error)
	fromDom func(Dom) (Params, error)
	upsert  func(context.Context, Params) (Row, error)
	del     func(context.Context, string) error
}

// keyPattern is the shape every primary key must have: URL-safe (so DELETE
// /admin/<name>/{id} can always reach the row) and wide enough for slugs,
// GitHub repo names (mixed case, dots) and dotted site-copy keys.
var keyPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,99}$`)

// registerResource mounts GET (list), PUT (upsert) and DELETE on /admin/<name>,
// all behind auth. The DELETE id segment is positional regardless of the real
// primary key (slug, repo, key…).
func registerResource[Row, Dom, Params any](mux *http.ServeMux, s *server, name string, rs resource[Row, Dom, Params]) {
	base := "/admin/" + name
	mux.HandleFunc("GET "+base, s.requireAuth(rs.handleList(s)))
	mux.HandleFunc("PUT "+base, s.requireAuth(rs.handleUpsert(s)))
	mux.HandleFunc("DELETE "+base+"/{id}", s.requireAuth(rs.handleDelete(s)))
}

func (rs resource[Row, Dom, Params]) handleList(s *server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := rs.list(r.Context())
		if err != nil {
			s.logger.Error("list failed", "err", err)
			s.errorJSON(w, http.StatusInternalServerError, "db error")
			return
		}
		out := make([]Dom, 0, len(rows))
		for _, row := range rows {
			d, err := rs.toDom(row)
			if err != nil {
				s.logger.Error("map failed", "err", err)
				s.errorJSON(w, http.StatusInternalServerError, "decode error")
				return
			}
			out = append(out, d)
		}
		s.writeJSON(w, http.StatusOK, out)
	}
}

func (rs resource[Row, Dom, Params]) handleUpsert(s *server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var d Dom
		if err := readJSON(w, r, &d); err != nil {
			s.badBody(w, err)
			return
		}
		if !keyPattern.MatchString(rs.key(d)) {
			s.errorJSON(w, http.StatusBadRequest, "invalid id: use 1-100 of A-Z a-z 0-9 . _ -, starting with a letter or digit")
			return
		}
		params, err := rs.fromDom(d)
		if err != nil {
			s.errorJSON(w, http.StatusBadRequest, err.Error())
			return
		}
		row, err := rs.upsert(r.Context(), params)
		if err != nil {
			s.logger.Error("upsert failed", "err", err)
			s.errorJSON(w, http.StatusInternalServerError, "db error")
			return
		}
		out, err := rs.toDom(row)
		if err != nil {
			s.logger.Error("map failed", "err", err)
			s.errorJSON(w, http.StatusInternalServerError, "decode error")
			return
		}
		s.writeJSON(w, http.StatusOK, out)
	}
}

func (rs resource[Row, Dom, Params]) handleDelete(s *server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := rs.del(r.Context(), r.PathValue("id")); err != nil {
			s.logger.Error("delete failed", "err", err)
			s.errorJSON(w, http.StatusInternalServerError, "db error")
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

// --- Pilot (singleton: GET + PUT, no list/delete) ---

func (s *server) handlePilotGet() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		row, err := s.q.GetPilot(r.Context())
		if err != nil {
			s.logger.Error("get pilot failed", "err", err)
			s.errorJSON(w, http.StatusInternalServerError, "db error")
			return
		}
		d, err := pilotToDomain(row)
		if err != nil {
			s.logger.Error("map pilot failed", "err", err)
			s.errorJSON(w, http.StatusInternalServerError, "decode error")
			return
		}
		s.writeJSON(w, http.StatusOK, d)
	}
}

func (s *server) handlePilotPut() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var d domain.Pilot
		if err := readJSON(w, r, &d); err != nil {
			s.badBody(w, err)
			return
		}
		params, err := pilotFromDomain(d)
		if err != nil {
			s.errorJSON(w, http.StatusBadRequest, err.Error())
			return
		}
		row, err := s.q.UpdatePilot(r.Context(), params)
		if err != nil {
			s.logger.Error("update pilot failed", "err", err)
			s.errorJSON(w, http.StatusInternalServerError, "db error")
			return
		}
		out, err := pilotToDomain(row)
		if err != nil {
			s.logger.Error("map pilot failed", "err", err)
			s.errorJSON(w, http.StatusInternalServerError, "decode error")
			return
		}
		s.writeJSON(w, http.StatusOK, out)
	}
}

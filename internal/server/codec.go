package server

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
)

// maxBodyBytes caps request bodies; content payloads are a few KB at most.
const maxBodyBytes = 1 << 20 // 1 MB

// writeJSON sends v as a JSON response with the given status.
func (s *server) writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		s.logger.Error("encode response failed", "err", err)
	}
}

// errorJSON sends a {"error": msg} body with the given status.
func (s *server) errorJSON(w http.ResponseWriter, status int, msg string) {
	s.writeJSON(w, status, map[string]string{"error": msg})
}

// readJSON decodes the request body into dst, rejecting unknown fields and
// bodies larger than maxBodyBytes.
func readJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxBodyBytes))
	dec.DisallowUnknownFields()
	return dec.Decode(dst)
}

// badBody answers a readJSON failure: 413 when the body hit the size cap,
// 400 otherwise.
func (s *server) badBody(w http.ResponseWriter, err error) {
	if _, ok := errors.AsType[*http.MaxBytesError](err); ok {
		s.errorJSON(w, http.StatusRequestEntityTooLarge, "body too large")
		return
	}
	s.errorJSON(w, http.StatusBadRequest, "invalid body")
}

// --- scalar conversions between domain and store representations ---

func b2i(b bool) int64 {
	if b {
		return 1
	}
	return 0
}

func i2b(i int64) bool { return i != 0 }

// nullStr maps "" to a NULL column value.
func nullStr(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}

func nullToStr(n sql.NullString) string {
	if n.Valid {
		return n.String
	}
	return ""
}

// decodeCol unmarshals a JSON text column into dst, tolerating empty strings.
func decodeCol(raw string, dst any) error {
	if raw == "" {
		return nil
	}
	return json.Unmarshal([]byte(raw), dst)
}

// jsonOr marshals v to a JSON string, substituting empty for a nil result so
// arrays/objects are stored as "[]"/"{}" rather than "null".
func jsonOr(v any, empty string) (string, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	if string(b) == "null" {
		return empty, nil
	}
	return string(b), nil
}

package server

import (
	"database/sql"
	"encoding/json"
	"net/http"
)

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

// readJSON decodes the request body into dst, rejecting unknown fields.
func readJSON(r *http.Request, dst any) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(dst)
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

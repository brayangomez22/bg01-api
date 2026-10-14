package auth

import (
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestValidToken(t *testing.T) {
	secret := []byte("0123456789abcdef0123456789abcdef")
	valid := NewToken(secret, time.Hour)
	exp, sig, _ := strings.Cut(valid, ".")

	// A later expiry re-using the original signature: the classic tamper.
	laterExp := strconv.FormatInt(time.Now().Add(48*time.Hour).Unix(), 10)

	tests := []struct {
		name   string
		secret []byte
		token  string
		want   bool
	}{
		{"valid", secret, valid, true},
		{"expired", secret, NewToken(secret, -time.Second), false},
		{"tampered signature", secret, exp + "." + flipLast(sig), false},
		{"tampered expiry", secret, laterExp + "." + sig, false},
		{"wrong secret", []byte("another-secret-another-secret-xx"), valid, false},
		{"empty secret rejects its own token", nil, NewToken(nil, time.Hour), false},
		{"empty secret rejects forged token", []byte{}, exp + "." + sign(nil, exp), false},
		{"no separator", secret, exp + sig, false},
		{"non-numeric expiry", secret, "soon." + sign(secret, "soon"), false},
		{"empty token", secret, "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ValidToken(tt.secret, tt.token); got != tt.want {
				t.Errorf("ValidToken() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCheckPassword(t *testing.T) {
	hash, err := HashPassword("correct horse")
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name     string
		hash     string
		password string
		want     bool
	}{
		{"match", hash, "correct horse", true},
		{"mismatch", hash, "battery staple", false},
		{"empty password", hash, "", false},
		{"invalid hash", "not-a-bcrypt-hash", "correct horse", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CheckPassword(tt.hash, tt.password); got != tt.want {
				t.Errorf("CheckPassword() = %v, want %v", got, tt.want)
			}
		})
	}
}

// flipLast changes the final character so the signature no longer matches.
func flipLast(s string) string {
	last := s[len(s)-1]
	repl := byte('A')
	if last == 'A' {
		repl = 'B'
	}
	return s[:len(s)-1] + string(repl)
}

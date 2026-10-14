package config

import (
	"strings"
	"testing"
)

func TestValidate(t *testing.T) {
	const hash = "$2a$10$abcdefghijklmnopqrstuuK3E8yB1c2bO5h9C1Yb3pUxk3gT8x7y."
	secret := strings.Repeat("s", MinSessionSecretLen)

	tests := []struct {
		name    string
		cfg     Config
		wantErr []string // substrings expected in the error; nil means valid
	}{
		{"valid", Config{AdminPasswordHash: hash, SessionSecret: secret}, nil},
		{"missing hash", Config{SessionSecret: secret}, []string{"ADMIN_PASSWORD_HASH"}},
		{"empty secret", Config{AdminPasswordHash: hash}, []string{"SESSION_SECRET"}},
		{"short secret", Config{AdminPasswordHash: hash, SessionSecret: secret[1:]}, []string{"SESSION_SECRET"}},
		{"both missing", Config{}, []string{"ADMIN_PASSWORD_HASH", "SESSION_SECRET"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if tt.wantErr == nil {
				if err != nil {
					t.Fatalf("Validate() = %v, want nil", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("Validate() = nil, want error mentioning %v", tt.wantErr)
			}
			for _, want := range tt.wantErr {
				if !strings.Contains(err.Error(), want) {
					t.Errorf("Validate() = %q, want it to mention %q", err, want)
				}
			}
		})
	}
}

package config

import (
	"os"
	"path/filepath"
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

func TestLoadDotEnv(t *testing.T) {
	const file = `# comment line

PLAIN=value
DOUBLE="quoted value"
SINGLE='single quoted'
  SPACED  =  padded  
WITH_EQUALS=a=b=c
EMPTY=
no_equals_line
=no_key
PRESET=from-file
`
	tests := []struct {
		key     string
		want    string
		wantSet bool
	}{
		{"PLAIN", "value", true},
		{"DOUBLE", "quoted value", true},
		{"SINGLE", "single quoted", true},
		{"SPACED", "padded", true},
		{"WITH_EQUALS", "a=b=c", true},
		{"EMPTY", "", true},
		{"no_equals_line", "", false},
		{"PRESET", "from-env", true}, // the real environment wins
	}

	for _, tt := range tests {
		unsetForTest(t, tt.key)
	}
	t.Setenv("PRESET", "from-env")

	path := filepath.Join(t.TempDir(), ".env")
	if err := os.WriteFile(path, []byte(file), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := loadDotEnv(path); err != nil {
		t.Fatalf("loadDotEnv() = %v", err)
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			got, set := os.LookupEnv(tt.key)
			if set != tt.wantSet || got != tt.want {
				t.Errorf("%s = (%q, set=%v), want (%q, set=%v)", tt.key, got, set, tt.want, tt.wantSet)
			}
		})
	}
}

func TestLoadDotEnvErrors(t *testing.T) {
	dir := t.TempDir()
	tooLong := filepath.Join(dir, "long.env")
	// bufio.Scanner's default limit is 64 KiB per line.
	if err := os.WriteFile(tooLong, []byte("BIG="+strings.Repeat("x", 70*1024)+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	unsetForTest(t, "BIG")

	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{"missing file is fine", filepath.Join(dir, "absent.env"), false},
		{"line over scanner limit", tooLong, true},
		{"path is a directory", dir, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := loadDotEnv(tt.path); (err != nil) != tt.wantErr {
				t.Errorf("loadDotEnv() = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// unsetForTest unsets key for the duration of the test, restoring any
// previous value afterwards (t.Setenv registers the restore).
func unsetForTest(t *testing.T, key string) {
	t.Helper()
	t.Setenv(key, "")
	os.Unsetenv(key)
}

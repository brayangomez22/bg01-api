package store

import (
	"testing"

	"github.com/pressly/goose/v3"
)

// TestMigrationsRoundTrip applies every migration, rolls all of them back and
// applies them again, so each Down really undoes its Up.
func TestMigrationsRoundTrip(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	// Every connection to :memory: is a separate database; pin the pool to one.
	db.SetMaxOpenConns(1)
	defer db.Close()
	goose.SetLogger(goose.NopLogger())

	tables := func() int {
		t.Helper()
		var n int
		err := db.QueryRow(`SELECT COUNT(*) FROM sqlite_master WHERE type = 'table' AND name NOT LIKE 'sqlite_%' AND name != 'goose_db_version'`).Scan(&n)
		if err != nil {
			t.Fatal(err)
		}
		return n
	}

	steps := []struct {
		name string
		run  func() error
	}{
		{"up", func() error { return Migrate(db) }},
		{"down to zero", func() error { return goose.DownTo(db, ".", 0) }},
		{"up again", func() error { return Migrate(db) }},
	}
	var afterFirstUp int
	for _, s := range steps {
		if err := s.run(); err != nil {
			t.Fatalf("%s: %v", s.name, err)
		}
		switch n := tables(); s.name {
		case "up":
			if n == 0 {
				t.Fatal("up: no tables created")
			}
			afterFirstUp = n
		case "down to zero":
			if n != 0 {
				t.Fatalf("down: %d tables left behind", n)
			}
		case "up again":
			if n != afterFirstUp {
				t.Fatalf("up again: %d tables, want %d", n, afterFirstUp)
			}
		}
	}

	// The seeded pilot singleton must come back with the schema.
	var callsign string
	if err := db.QueryRow(`SELECT callsign FROM pilot WHERE id = 1`).Scan(&callsign); err != nil {
		t.Fatalf("pilot seed after re-migrate: %v", err)
	}
}

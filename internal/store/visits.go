package store

// Hand-written store methods for the visits counter. sqlc is not in the loop for
// this table (no queries/visits.sql); these methods hang off the same *Queries
// receiver as the generated code and use the shared DBTX.

import "context"

// DailyVisits is one day's unique-visitor count.
type DailyVisits struct {
	Day   string `json:"day"`
	Count int64  `json:"count"`
}

// RecordVisit registers a unique (day, visitor) tuple. The INSERT OR IGNORE on
// the composite primary key makes repeat visits within the same day a no-op, so
// every person counts at most once per day.
func (q *Queries) RecordVisit(ctx context.Context, day, visitorHash string) error {
	_, err := q.db.ExecContext(ctx,
		`INSERT OR IGNORE INTO visits (day, visitor_hash) VALUES (?, ?)`,
		day, visitorHash,
	)
	return err
}

// VisitsSince returns per-day unique-visitor counts for every day on or after
// `since` (a 'YYYY-MM-DD' string), oldest first. Days with zero visits are
// absent from the result — callers fill the gaps.
func (q *Queries) VisitsSince(ctx context.Context, since string) ([]DailyVisits, error) {
	rows, err := q.db.QueryContext(ctx,
		`SELECT day, COUNT(*) FROM visits WHERE day >= ? GROUP BY day ORDER BY day`,
		since,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []DailyVisits
	for rows.Next() {
		var d DailyVisits
		if err := rows.Scan(&d.Day, &d.Count); err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

// VisitsTotal returns the all-time count of unique visitor-days.
func (q *Queries) VisitsTotal(ctx context.Context) (int64, error) {
	var n int64
	err := q.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM visits`).Scan(&n)
	return n, err
}

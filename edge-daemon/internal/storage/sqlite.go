package storage

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"rigtelemetry/shared/model"
)

const (
	StatusPending   = "pending"
	StatusInFlight  = "in_flight"
	StatusDone      = "done"
)

// Store is the edge SQLite outbox.
type Store struct {
	db *sql.DB
}

// OutboxRow is a claimed outbox record ready for MQTT publish.
type OutboxRow struct {
	ID      int64
	Point   model.TelemetryPoint
	Payload []byte
}

// Open opens (or creates) the SQLite database and runs migrations.
func Open(path string) (*Store, error) {
	db, err := sql.Open("sqlite3", path+"?_journal_mode=WAL&_busy_timeout=5000")
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	db.SetMaxOpenConns(1)
	s := &Store{db: db}
	if err := s.migrate(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return s, nil
}

// Close closes the underlying database.
func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) migrate() error {
	const schema = `
CREATE TABLE IF NOT EXISTS telemetry_outbox (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    time         TEXT    NOT NULL,
    well_id      TEXT    NOT NULL,
    bit_depth    REAL    NOT NULL,
    rop          REAL    NOT NULL,
    wob          REAL    NOT NULL,
    gamma_ray    REAL    NOT NULL,
    payload_json TEXT    NOT NULL,
    status       TEXT    NOT NULL DEFAULT 'pending',
    attempts     INTEGER NOT NULL DEFAULT 0,
    created_at   TEXT    NOT NULL,
    updated_at   TEXT    NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_outbox_status_id ON telemetry_outbox (status, id);
`
	if _, err := s.db.Exec(schema); err != nil {
		return fmt.Errorf("migrate: %w", err)
	}
	return nil
}

// Enqueue inserts a telemetry point as pending in the outbox.
func (s *Store) Enqueue(point model.TelemetryPoint, payload []byte) error {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	_, err := s.db.Exec(`
INSERT INTO telemetry_outbox
  (time, well_id, bit_depth, rop, wob, gamma_ray, payload_json, status, attempts, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, 0, ?, ?)`,
		point.Time.UTC().Format(time.RFC3339Nano),
		point.WellID,
		point.BitDepth,
		point.ROP,
		point.WOB,
		point.GammaRay,
		string(payload),
		StatusPending,
		now,
		now,
	)
	if err != nil {
		return fmt.Errorf("enqueue: %w", err)
	}
	return nil
}

// ClaimBatch marks up to limit pending (or stale in_flight) rows as in_flight and returns them.
func (s *Store) ClaimBatch(limit int) ([]OutboxRow, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	rows, err := tx.Query(`
SELECT id, time, well_id, bit_depth, rop, wob, gamma_ray, payload_json
FROM telemetry_outbox
WHERE status IN (?, ?)
ORDER BY id ASC
LIMIT ?`, StatusPending, StatusInFlight, limit)
	if err != nil {
		return nil, fmt.Errorf("claim query: %w", err)
	}
	defer rows.Close()

	var batch []OutboxRow
	var ids []int64
	for rows.Next() {
		var (
			row     OutboxRow
			timeStr string
			payload string
		)
		if err := rows.Scan(
			&row.ID, &timeStr, &row.Point.WellID,
			&row.Point.BitDepth, &row.Point.ROP, &row.Point.WOB, &row.Point.GammaRay,
			&payload,
		); err != nil {
			return nil, err
		}
		ts, err := time.Parse(time.RFC3339Nano, timeStr)
		if err != nil {
			ts, _ = time.Parse(time.RFC3339, timeStr)
		}
		row.Point.Time = ts
		row.Payload = []byte(payload)
		batch = append(batch, row)
		ids = append(ids, row.ID)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(ids) == 0 {
		return nil, tx.Commit()
	}

	now := time.Now().UTC().Format(time.RFC3339Nano)
	for _, id := range ids {
		if _, err := tx.Exec(`
UPDATE telemetry_outbox
SET status = ?, attempts = attempts + 1, updated_at = ?
WHERE id = ?`, StatusInFlight, now, id); err != nil {
			return nil, err
		}
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return batch, nil
}

// MarkPublished marks a row as successfully published.
func (s *Store) MarkPublished(id int64) error {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	_, err := s.db.Exec(`
UPDATE telemetry_outbox SET status = ?, updated_at = ? WHERE id = ?`,
		StatusDone, now, id)
	return err
}

// MarkFailed returns a row to pending so it can be retried.
func (s *Store) MarkFailed(id int64) error {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	_, err := s.db.Exec(`
UPDATE telemetry_outbox SET status = ?, updated_at = ? WHERE id = ?`,
		StatusPending, now, id)
	return err
}

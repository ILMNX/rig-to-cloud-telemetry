package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"rigtelemetry/shared/model"
)

// Store wraps a pgx pool for TimescaleDB access.
type Store struct {
	pool *pgxpool.Pool
}

// Open connects to PostgreSQL/TimescaleDB.
func Open(ctx context.Context, databaseURL string) (*Store, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("pgx pool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("db ping: %w", err)
	}
	return &Store{pool: pool}, nil
}

// Close closes the pool.
func (s *Store) Close() {
	s.pool.Close()
}

// Ping checks database connectivity.
func (s *Store) Ping(ctx context.Context) error {
	return s.pool.Ping(ctx)
}

// Insert writes a single telemetry point.
func (s *Store) Insert(ctx context.Context, p model.TelemetryPoint) error {
	_, err := s.pool.Exec(ctx, `
INSERT INTO drilling_telemetry (time, well_id, bit_depth, rop, wob, gamma_ray)
VALUES ($1, $2, $3, $4, $5, $6)`,
		p.Time.UTC(), p.WellID, p.BitDepth, p.ROP, p.WOB, p.GammaRay,
	)
	if err != nil {
		return fmt.Errorf("insert: %w", err)
	}
	return nil
}

// ListWells returns distinct well IDs.
func (s *Store) ListWells(ctx context.Context) ([]string, error) {
	rows, err := s.pool.Query(ctx, `
SELECT DISTINCT well_id FROM drilling_telemetry ORDER BY well_id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	wells := make([]string, 0)
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		wells = append(wells, id)
	}
	return wells, rows.Err()
}

// Latest returns the most recent point for a well, or nil if none.
func (s *Store) Latest(ctx context.Context, wellID string) (*model.TelemetryPoint, error) {
	row := s.pool.QueryRow(ctx, `
SELECT time, well_id, bit_depth, rop, wob, gamma_ray
FROM drilling_telemetry
WHERE well_id = $1
ORDER BY time DESC
LIMIT 1`, wellID)

	var p model.TelemetryPoint
	err := row.Scan(&p.Time, &p.WellID, &p.BitDepth, &p.ROP, &p.WOB, &p.GammaRay)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// Range returns points for a well in [from, to) ordered ascending by time.
func (s *Store) Range(ctx context.Context, wellID string, from, to time.Time, limit int) ([]model.TelemetryPoint, error) {
	if limit <= 0 || limit > 20000 {
		limit = 5000
	}
	rows, err := s.pool.Query(ctx, `
SELECT time, well_id, bit_depth, rop, wob, gamma_ray
FROM drilling_telemetry
WHERE well_id = $1 AND time >= $2 AND time < $3
ORDER BY time ASC
LIMIT $4`, wellID, from.UTC(), to.UTC(), limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	points := make([]model.TelemetryPoint, 0)
	for rows.Next() {
		var p model.TelemetryPoint
		if err := rows.Scan(&p.Time, &p.WellID, &p.BitDepth, &p.ROP, &p.WOB, &p.GammaRay); err != nil {
			return nil, err
		}
		points = append(points, p)
	}
	return points, rows.Err()
}

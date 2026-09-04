// Package store persists telemetry into TimescaleDB.
package store

import (
	_ "github.com/jackc/pgx/v5"
	_ "rigtelemetry/shared/model"
)

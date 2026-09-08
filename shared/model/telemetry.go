package model

import (
	"errors"
	"time"
)

// TelemetryPoint is a single drilling telemetry sample shared between
// the edge daemon and cloud ingestion services.
type TelemetryPoint struct {
	Time     time.Time `json:"time"`
	WellID   string    `json:"well_id"`
	BitDepth float64   `json:"bit_depth"`
	ROP      float64   `json:"rop"`
	WOB      float64   `json:"wob"`
	GammaRay float64   `json:"gamma_ray"`
}

var ErrInvalidWellID = errors.New("well_id is required")

// Validate checks required fields on a telemetry sample.
func (p TelemetryPoint) Validate() error {
	if p.WellID == "" {
		return ErrInvalidWellID
	}
	return nil
}

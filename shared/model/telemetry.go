package model

import "time"

// TelemetryPoint is a single drilling telemetry sample shared between
// the edge daemon and cloud ingestion services.
type TelemetryPoint struct {
	Time      time.Time `json:"time"`
	WellID    string    `json:"well_id"`
	BitDepth  float64   `json:"bit_depth"`
	ROP       float64   `json:"rop"`
	WOB       float64   `json:"wob"`
	GammaRay  float64   `json:"gamma_ray"`
}

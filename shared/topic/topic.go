// Package topic defines MQTT topic helpers shared by edge and cloud.
package topic

import (
	"fmt"
	"strings"
)

const (
	// Prefix is the root MQTT namespace for telemetry points.
	Prefix = "telemetry"
	// PointsSuffix is the leaf segment for point payloads.
	PointsSuffix = "points"
)

// Points returns the publish topic for a well: telemetry/{wellID}/points.
func Points(wellID string) string {
	return fmt.Sprintf("%s/%s/%s", Prefix, wellID, PointsSuffix)
}

// PointsWildcard returns the cloud subscribe filter: telemetry/+/points.
func PointsWildcard() string {
	return fmt.Sprintf("%s/+/%s", Prefix, PointsSuffix)
}

// ParseWellID extracts well_id from a topic like telemetry/{wellID}/points.
// Returns false if the topic does not match the expected shape.
func ParseWellID(mqttTopic string) (string, bool) {
	parts := strings.Split(mqttTopic, "/")
	if len(parts) != 3 {
		return "", false
	}
	if parts[0] != Prefix || parts[2] != PointsSuffix || parts[1] == "" || parts[1] == "+" {
		return "", false
	}
	return parts[1], true
}

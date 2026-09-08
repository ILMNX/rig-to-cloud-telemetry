package topic_test

import (
	"testing"

	"rigtelemetry/shared/topic"
)

func TestPoints(t *testing.T) {
	got := topic.Points("WELL-DEMO-01")
	want := "telemetry/WELL-DEMO-01/points"
	if got != want {
		t.Fatalf("Points() = %q, want %q", got, want)
	}
}

func TestPointsWildcard(t *testing.T) {
	got := topic.PointsWildcard()
	want := "telemetry/+/points"
	if got != want {
		t.Fatalf("PointsWildcard() = %q, want %q", got, want)
	}
}

func TestParseWellID(t *testing.T) {
	tests := []struct {
		in      string
		wantID  string
		wantOK  bool
	}{
		{"telemetry/WELL-DEMO-01/points", "WELL-DEMO-01", true},
		{"telemetry/+/points", "", false},
		{"telemetry/WELL/extra/points", "", false},
		{"other/WELL/points", "", false},
		{"telemetry//points", "", false},
	}
	for _, tc := range tests {
		id, ok := topic.ParseWellID(tc.in)
		if ok != tc.wantOK || id != tc.wantID {
			t.Fatalf("ParseWellID(%q) = (%q, %v), want (%q, %v)", tc.in, id, ok, tc.wantID, tc.wantOK)
		}
	}
}

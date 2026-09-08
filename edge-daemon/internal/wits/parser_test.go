package wits_test

import (
	"testing"

	"rigtelemetry/edge/internal/wits"
)

func TestParserFeedCompleteFrame(t *testing.T) {
	p := wits.NewParser("WELL-DEMO-01")
	frame := "&&\n01081205.50\n011318.20\n010A12.50\n012272.40\n!!\n"
	pts := p.Feed([]byte(frame))
	if len(pts) != 1 {
		t.Fatalf("got %d points, want 1", len(pts))
	}
	pt := pts[0]
	if pt.WellID != "WELL-DEMO-01" {
		t.Fatalf("WellID = %q", pt.WellID)
	}
	if pt.BitDepth != 1205.50 {
		t.Fatalf("BitDepth = %v", pt.BitDepth)
	}
	if pt.ROP != 18.20 {
		t.Fatalf("ROP = %v", pt.ROP)
	}
	if pt.WOB != 12.50 {
		t.Fatalf("WOB = %v", pt.WOB)
	}
	if pt.GammaRay != 72.40 {
		t.Fatalf("GammaRay = %v", pt.GammaRay)
	}
	if pt.Time.IsZero() {
		t.Fatal("Time should be set")
	}
}

func TestParserFeedChunked(t *testing.T) {
	p := wits.NewParser("WELL-X")
	var pts []any
	for _, chunk := range []string{"&&\n", "01081000.00\n", "011320.00\n", "010A10.00\n", "012280.00\n", "!!\n"} {
		got := p.Feed([]byte(chunk))
		for _, g := range got {
			pts = append(pts, g)
		}
	}
	if len(pts) != 1 {
		t.Fatalf("got %d points, want 1", len(pts))
	}
}

func TestParserIgnoresIncomplete(t *testing.T) {
	p := wits.NewParser("WELL-X")
	if got := p.Feed([]byte("&&\n01081000.00\n")); len(got) != 0 {
		t.Fatalf("expected no points before frame end, got %d", len(got))
	}
}

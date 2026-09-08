package wits

import (
	"strconv"
	"strings"
	"time"

	"rigtelemetry/shared/model"
	"rigtelemetry/shared/witsids"
)

// Parser assembles WITS Level 0 ASCII frames and maps items to TelemetryPoint.
type Parser struct {
	wellID string
	buf    strings.Builder
	inFrame bool
}

// NewParser returns a parser that stamps wellID and wall-clock time on each point.
func NewParser(wellID string) *Parser {
	return &Parser{wellID: wellID}
}

// Feed consumes a chunk of text (typically one line including newline) and
// returns completed points. Incomplete frames stay buffered.
func (p *Parser) Feed(chunk []byte) []model.TelemetryPoint {
	text := string(chunk)
	var out []model.TelemetryPoint

	for _, line := range strings.SplitAfter(text, "\n") {
		if line == "" {
			continue
		}
		trimmed := strings.TrimSpace(line)
		switch {
		case trimmed == witsids.FrameStart:
			p.inFrame = true
			p.buf.Reset()
			p.buf.WriteString(trimmed)
			p.buf.WriteByte('\n')
		case trimmed == witsids.FrameEnd && p.inFrame:
			p.buf.WriteString(trimmed)
			p.buf.WriteByte('\n')
			if pt, ok := parseFrame(p.buf.String(), p.wellID, time.Now().UTC()); ok {
				out = append(out, pt)
			}
			p.inFrame = false
			p.buf.Reset()
		case p.inFrame:
			p.buf.WriteString(line)
			if !strings.HasSuffix(line, "\n") {
				p.buf.WriteByte('\n')
			}
		}
	}
	return out
}

func parseFrame(frame string, wellID string, ts time.Time) (model.TelemetryPoint, bool) {
	pt := model.TelemetryPoint{
		Time:   ts,
		WellID: wellID,
	}
	found := 0
	for _, raw := range strings.Split(frame, "\n") {
		line := strings.TrimSpace(raw)
		if line == "" || line == witsids.FrameStart || line == witsids.FrameEnd {
			continue
		}
		if len(line) < 5 {
			continue
		}
		item := line[:4]
		valStr := line[4:]
		val, err := strconv.ParseFloat(valStr, 64)
		if err != nil {
			continue
		}
		switch item {
		case witsids.ItemBitDepth:
			pt.BitDepth = val
			found++
		case witsids.ItemROP:
			pt.ROP = val
			found++
		case witsids.ItemWOB:
			pt.WOB = val
			found++
		case witsids.ItemGammaRay:
			pt.GammaRay = val
			found++
		}
	}
	return pt, found > 0
}

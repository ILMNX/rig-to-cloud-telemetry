// Package witsids defines WITS Level 0 item identifiers used by the
// simulator and edge parser. Keep these in sync with simulator/wits_ids.py.
package witsids

const (
	// FrameStart marks the beginning of a WITS ASCII frame.
	FrameStart = "&&"
	// FrameEnd marks the end of a WITS ASCII frame.
	FrameEnd = "!!"

	// ItemBitDepth is WITS item 0108 (bit depth, metres).
	ItemBitDepth = "0108"
	// ItemROP is WITS item 0113 (rate of penetration).
	ItemROP = "0113"
	// ItemWOB is WITS item 010A (weight on bit).
	ItemWOB = "010A"
	// ItemGammaRay is WITS item 0122 (gamma ray).
	ItemGammaRay = "0122"
)

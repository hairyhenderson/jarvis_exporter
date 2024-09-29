package jarvis

import (
	"math/bits"
	"strconv"
	"strings"
)

// Commands sent by desk
const (
	Height    uint8 = 0x01 // Height report (in mm, P0/P1, P2 unused - always 0xf)
	LimitResp uint8 = 0x20 // Max-height set/cleared; response to SetMax (0x21)
	GetMax    uint8 = 0x21 // Report max-height; [P0,P1] = max-height (c.f. SetMax)
	GetMin    uint8 = 0x22 // Report min-height; [P0,P1] = min-height
	LimitStop uint8 = 0x23 // Min/Max reached (0x01 "Max-height reached", 0x02 "Min-height reached")
	Reset     uint8 = 0x40 // Indicates desk in RESET mode; Displays "RESET"
	Preset    uint8 = 0x92 // Moving to Preset location ([0x4,0x8,0x10,0x20] mapping to presets [1,2,3,4])
)

// Commands sent by handset
const (
	ProgMem1   uint8 = 0x03 // Set memory position 1 to current height
	ProgMem2   uint8 = 0x04 // Set memory position 2 to current height
	ProgMem3   uint8 = 0x25 // Set memory position 3 to current height
	ProgMem4   uint8 = 0x26 // Set memory position 4 to current height
	Units      uint8 = 0x0e // Set units to cm/inches (param 0x00 == cm, 0x01 == in)
	MemMode    uint8 = 0x19 // Set memory mode (0x00 One-touch mode, 0x01 Constant touch mode)
	CollSens   uint8 = 0x1d // Set anti-collision sensitivity (Sent 1x; no repeats, 1/2/3 high/medium/low)
	SetMax     uint8 = 0x21 // Set max height; Sets max-height to current height
	SetMin     uint8 = 0x22 // Set min height; Sets min-height to current height
	LimitClear uint8 = 0x23 // Clear min/max height (0x01 Max-height cleared, 0x02 Min-height cleared)
	Wake       uint8 = 0x29 // Poll message sent when desk doesn't respond to BREAK messages
	Calibrate  uint8 = 0x91 // Height calibration (Repeats 2x) (Desk must be at lowest position, enters RESET mode after this)
)

type Message struct {
	raw    []byte
	params uint32
	addr   uint8
	Type   uint8
	length uint8
	cksum  uint8
}

//nolint:gocyclo
func (m *Message) String() string {
	s := strings.Builder{}
	if m.addr == 0xf2 {
		s.WriteString("from desk: ")
	}

	switch m.Type {
	case Height:
		height := uint64(m.params >> 8)

		s.WriteString("height: ")
		s.WriteString(strconv.FormatUint(height, 10))

		if height <= 550 {
			s.WriteString("in")
		} else {
			s.WriteString("mm")
		}
	case Preset:
		s.WriteString("preset: ")
		s.WriteString(strconv.FormatUint(uint64(m.Preset()), 10))
	case LimitResp:
	case GetMax:
	case GetMin:
	case LimitStop:
		switch m.params {
		case 0x01:
			s.WriteString("Max-height reached")
		case 0x02:
			s.WriteString("Min-height reached")
		default:
			s.WriteString("invalid params for LIMIT_STOP")
		}
	case Reset:
		s.WriteString("reset!")
	}

	return s.String()
}

func (m Message) Preset() uint8 {
	if m.Type != Preset {
		return 0
	}

	// The preset response returned by the desk is 4/8/16/32 or the 3rd, 4th,
	// 6th, and 7th bit set (from the right) in binary. To convert these to
	// their 1-indexed positions, we shift right by 1 and count the trailing
	// zeroes. Another approach could be to use math.Log2 with lots of float64
	// conversions, but this is far more efficient!
	//nolint:gosec // disable G115
	return uint8(bits.TrailingZeros8(uint8(m.params) >> 1))
}

// Height returns the height (in mm)
func (m Message) Height() uint64 {
	if m.Type != Height {
		return 0
	}

	// strip off the last 0xf - only the first two bytes are part of the height
	h := uint64(m.params >> 8)

	// if the value is <= 550 we assume the unit is 1/10 inches, so multiply by
	// 2.54 to get mm!
	if h <= 550 {
		h = uint64(float64(h) * 2.54)
	}

	return h
}

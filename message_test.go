package jarvis

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMessagePreset(t *testing.T) {
	m := Message{}
	assert.Equal(t, uint8(0), m.Preset())

	m = Message{Type: Height}
	assert.Equal(t, uint8(0), m.Preset())

	testdata := []struct {
		params   uint32
		expected uint8
	}{
		{0x04, 1},
		{0x08, 2},
		{0x10, 3},
		{0x20, 4},
	}

	for _, d := range testdata {
		m := Message{
			Type:   Preset,
			params: d.params,
		}
		assert.Equal(t, d.expected, m.Preset())
	}
}

func TestMessageHeight(t *testing.T) {
	m := Message{}
	assert.Equal(t, uint64(0), m.Height())

	m = Message{Type: Preset}
	assert.Equal(t, uint64(0), m.Height())

	testdata := []struct {
		params   uint32
		expected uint64
	}{
		{0x000000, 0},
		{0x04ab0f, 1195},
		{0x02780f, 632},
		// if it's too low we assume it's 1/10ths of inches and convert to mm
		{0x01230f, 739},
		// unexpected last byte but...
		{0x02787e, 632},
	}

	for _, d := range testdata {
		m := Message{
			Type:   Height,
			params: d.params,
		}
		assert.Equal(t, d.expected, m.Height())
	}
}

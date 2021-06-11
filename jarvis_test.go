package jarvis

import (
	"bytes"
	"context"
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
)

func TestVerifyChecksum(t *testing.T) {
	m := Message{
		Type:   Height,
		length: 3,
		params: 0x01020f,
		cksum:  0x16,
	}

	assert.True(t, verifyChecksum(m))

	m.cksum = 0x07
	assert.False(t, verifyChecksum(m))
}

//nolint:funlen
func TestNextMessage(t *testing.T) {
	ctx := context.Background()

	j := New(bytes.NewReader(nil), InitMetrics("jarvis"))

	m, err := j.NextMessage(ctx)
	assert.Error(t, err)
	assert.Nil(t, m)

	j = New(bytes.NewReader([]byte{0x00}), InitMetrics("jarvis"))
	m, err = j.NextMessage(ctx)
	assert.Error(t, err)
	assert.Nil(t, m)

	// should error when the context is cancelled!
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	j = New(bytes.NewReader([]byte{0x00}), InitMetrics("jarvis"))
	m, err = j.NextMessage(ctx)
	assert.Error(t, err)
	assert.Nil(t, m)

	ctx = context.Background()

	// early close after incomplete message
	j = New(bytes.NewReader([]byte{0x0f2, 0x0f2}), InitMetrics("jarvis"))
	m, err = j.NextMessage(ctx)
	assert.Error(t, err)
	assert.Nil(t, m)

	// invalid length messages
	j = New(bytes.NewReader([]byte{
		0x0f2, 0x0f2, 0x01, 0x04,
		0x0f2, 0x0f2, 0x01, 0x04,
		0x0f2, 0x0f2, 0x01, 0x04,
	}), InitMetrics("jarvis"))
	m, err = j.NextMessage(ctx)
	assert.Error(t, err)
	assert.Nil(t, m)

	assert.Equal(t, 3.0, testutil.ToFloat64(j.metrics.invalidLenCount))
	assert.Equal(t, 0.0, testutil.ToFloat64(j.metrics.checksumErrorsCount))

	// bad checksum
	msg := []byte{
		0x0f2, 0x0f2, 0x01, 0x03, 0x01, 0x0a, 0x0f, 0x05, 0x7e,
	}
	j = New(bytes.NewReader(msg), InitMetrics("jarvis"))
	m, err = j.NextMessage(ctx)
	assert.NoError(t, err)
	assert.Equal(t, &Message{
		raw:    msg,
		addr:   0xf2,
		Type:   0x01,
		length: 0x03,
		params: 0x010a0f,
		cksum:  0x05,
	}, m)

	assert.Equal(t, 1.0, testutil.ToFloat64(j.metrics.checksumErrorsCount))
	assert.Equal(t, 0.0, testutil.ToFloat64(j.metrics.invalidLenCount))

	// legit message
	msg = []byte{
		0x0f2, 0x0f2, 0x01, 0x03, 0x01, 0x0a, 0x0f, 0x1e, 0x7e,
	}
	j = New(bytes.NewReader(msg), InitMetrics("jarvis"))
	m, err = j.NextMessage(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 0.0, testutil.ToFloat64(j.metrics.checksumErrorsCount))
	assert.Equal(t, &Message{
		raw:    msg,
		addr:   0xf2,
		Type:   0x01,
		length: 0x03,
		params: 0x010a0f,
		cksum:  0x1e,
	}, m)

	assert.Equal(t, 0.0, testutil.ToFloat64(j.metrics.invalidLenCount))
}

func TestMetricsCollector(t *testing.T) {
	m := InitMetrics("foo")

	probs, err := testutil.CollectAndLint(m)
	assert.NoError(t, err)
	assert.Empty(t, probs)
}

package jarvis

import (
	"context"
	"fmt"
	"io"

	"github.com/prometheus/client_golang/prometheus"
)

type Jarvis struct {
	rdr io.Reader

	metrics *Metrics
}

type Metrics struct {
	checksumErrorsCount prometheus.Counter
	invalidLenCount     prometheus.Counter
}

func InitMetrics(ns string) *Metrics {
	return &Metrics{
		checksumErrorsCount: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: ns,
			Name:      "checksum_errors_total",
			Help:      "A count of checksum errors encountered.",
		}),
		invalidLenCount: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: ns,
			Name:      "invalid_length_count_total",
			Help:      "Number of times an invalid length field was received.",
		}),
	}
}

// Collect implements Prometheus.Collector.
func (m *Metrics) Collect(ch chan<- prometheus.Metric) {
	m.checksumErrorsCount.Collect(ch)
	m.invalidLenCount.Collect(ch)
}

// Describe implements Prometheus.Collector.
func (m *Metrics) Describe(ch chan<- *prometheus.Desc) {
	m.checksumErrorsCount.Describe(ch)
	m.invalidLenCount.Describe(ch)
}

func New(r io.Reader, metrics *Metrics) *Jarvis {
	return &Jarvis{
		rdr:     r,
		metrics: metrics,
	}
}

// ReadMessages -
func (j *Jarvis) ReadMessages(ctx context.Context, errch chan error, ch chan *Message) {
	for {
		msg, err := j.NextMessage(ctx)

		// send non-nil messages back first - sometimes we want both!
		if msg != nil {
			ch <- msg
		}

		if err != nil {
			if errch != nil {
				errch <- err
			}

			close(ch)

			return
		}
	}
}

// NextMessage reads the next *Message from the given reader.
//
// the format is:
// [addr (2b)][command (1b)][param len (1b)][params (0-3b)][cksum (1b)][EOM (1b)]
//
// states (at beginning of loop)
// 0 = unsynchronized
// 1 = half-synchronized
// 2 = synchronized
// 3 = received command
// 4 = received param length (0-3), reading params
// 5 = received all params
// 6 = verified checksum
// 7 = eom
//
//nolint:gocyclo,funlen
func (j *Jarvis) NextMessage(ctx context.Context) (*Message, error) {
	state := uint8(0)
	curParam := uint8(0)
	p := Message{}

	// tracks resets local to this function, so that we can error after too many
	invalidLenCount := 0

	reset := func() {
		// reset state to 255 (aka -1), will be incremented on next loop to 0
		state = 255
		p = Message{}
	}

	for ; state < 7; state++ {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		if invalidLenCount >= 3 {
			return nil, fmt.Errorf("too many invalid lengths")
		}

		c := make([]byte, 1)

		n, err := j.rdr.Read(c)
		if err != nil {
			return nil, fmt.Errorf("read: %w", err)
		}

		if n == 0 {
			return nil, fmt.Errorf("short read")
		}

		x := c[0]
		p.raw = append(p.raw, x)

		switch {
		case state == 0 && (x == 0xf1 || x == 0xf2):
			p.addr = x
		case state == 1 && x == p.addr:
			// synchronized
			continue
		case state == 2:
			p.Type = x
		case state == 3:
			p.length = x

			// 0-indexed position of current param byte to parse
			curParam = x - 1

			// max number of param bytes is 3
			if p.length > 3 {
				fmt.Printf("invalid length %0#x is greater than 3 (raw: %#v)\n", p.length, p.raw)
				j.metrics.invalidLenCount.Inc()
				invalidLenCount++

				reset()
			}
		case state == 4 && p.length > 0:
			p.params += uint32(x) << (8 * curParam)

			if curParam > 0 {
				state = 3
				curParam--
			}
		case state == 5:
			p.cksum = x
		case state == 6 && x == 0x7e:
			if !verifyChecksum(p) {
				j.metrics.checksumErrorsCount.Inc()
			}

			return &p, nil
		default:
			reset()
		}
	}

	return nil, fmt.Errorf("invalid state %d", state)
}

// verifyChecksum, more for information than anything else - it's
// unreliable at certain heights (at least on my desk!)
func verifyChecksum(p Message) bool {
	rem := (uint32(p.Type) + uint32(p.length) + p.params) % 0xFF

	return rem == uint32(p.cksum)
}

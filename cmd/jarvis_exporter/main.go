package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"

	jarvis "github.com/hairyhenderson/jarvis_exporter"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/tarm/serial"
)

func main() {
	exitCode := 0
	defer func() { os.Exit(exitCode) }()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()

	addr := flag.String("addr", ":9833", "address to listen to")
	serialport := flag.String("serialport", "/dev/cu.usbserial-AH02F0IR", "serial port path")
	flag.Parse()

	jmetrics := jarvis.InitMetrics(ns)

	prometheus.MustRegister(deskHeightGauge, readErrorsCount, jmetrics)

	log.Printf("starting server at %s", *addr)

	srv, err := startServer(ctx, *addr, &exitCode)
	if err != nil {
		log.Printf("start: %v", err)

		exitCode = 1

		return
	}

	defer srv.Shutdown(ctx)

	if err := readLoop(ctx, *serialport, jmetrics); err != nil {
		fmt.Println(err)
	}
}

func startServer(ctx context.Context, addr string, exitCode *int) (*http.Server, error) {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	srv := &http.Server{
		Addr:        addr,
		Handler:     mux,
		BaseContext: func(net.Listener) context.Context { return ctx },
	}

	lc := net.ListenConfig{}

	l, err := lc.Listen(ctx, "tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("listen: %w", err)
	}

	go func() {
		err := srv.Serve(l)
		if err != nil && err != http.ErrServerClosed {
			log.Printf("server terminated with error: %v", err)

			*exitCode = 1
		}

		srv.Shutdown(ctx)
	}()

	return srv, nil
}

var (
	ns              = "jarvis"
	deskHeightGauge = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: ns,
		Name:      "desk_height_meters",
		Help:      "The current height of the desk, in meters.",
	})
	presetSelectedGauge = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: ns,
		Name:      "desk_preset_selected",
		Help:      "The last preset selected",
	})
	readErrorsCount = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: ns,
		Name:      "read_errors_total",
		Help:      "A count of read errors encountered. See logs for details.",
	})
)

func readLoop(ctx context.Context, portname string, jmetrics *jarvis.Metrics) error {
	options := serial.Config{Name: portname, Baud: 9600}

	// loop to reconnect to the serial port if it closes or we get a read
	// failure...
	for {
		if err := streamMessages(ctx, &options, jmetrics); err != nil {
			return err
		}
	}
}

func streamMessages(ctx context.Context, options *serial.Config, jmetrics *jarvis.Metrics) error {
	log.Println("opening serial port for streaming...")

	port, err := serial.OpenPort(options)
	if err != nil {
		return fmt.Errorf("port open: %w", err)
	}

	defer port.Close()

	desk := jarvis.New(port, jmetrics)

	errch := make(chan error)
	ch := make(chan *jarvis.Message)

	// start streaming messages into the given channel
	go desk.ReadMessages(ctx, errch, ch)

	for {
		select {
		case msg := <-ch:
			if msg == nil {
				break
			}

			switch msg.Type {
			case jarvis.Height:
				deskHeightGauge.Set(float64(msg.Height()) / 1000)
			case jarvis.Preset:
				log.Printf("moving to preset %#02x", msg.Preset())
				presetSelectedGauge.Set(float64(msg.Preset()))
			default:
				log.Printf("got message (type %#x): %s", msg.Type, msg)
			}
		case <-ctx.Done():
			return ctx.Err()
		case err = <-errch:
			readErrorsCount.Inc()

			log.Printf("read error: %v", err)

			// we count this as a non-fatal error - perhaps retrying will help...
			return nil
		}
	}
}

package serial

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"time"

	"go.bug.st/serial"
)

// Reader streams newline-delimited bytes from a serial port (or PTY slave).
type Reader struct {
	portPath string
	baud     int
}

// NewReader creates a serial reader for the given device path and baud rate.
func NewReader(portPath string, baud int) *Reader {
	return &Reader{portPath: portPath, baud: baud}
}

// Run opens the port (retrying on failure) and sends each line to out until ctx is cancelled.
func (r *Reader) Run(ctx context.Context, out chan<- []byte) error {
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		port, err := r.open()
		if err != nil {
			log.Printf("serial: open %s failed: %v; retrying in 2s", r.portPath, err)
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(2 * time.Second):
				continue
			}
		}
		err = r.readLoop(ctx, port, out)
		_ = port.Close()
		if ctx.Err() != nil {
			return ctx.Err()
		}
		log.Printf("serial: read loop ended: %v; reconnecting", err)
	}
}

func (r *Reader) open() (serial.Port, error) {
	mode := &serial.Mode{BaudRate: r.baud}
	port, err := serial.Open(r.portPath, mode)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", r.portPath, err)
	}
	return port, nil
}

func (r *Reader) readLoop(ctx context.Context, port serial.Port, out chan<- []byte) error {
	reader := bufio.NewReader(port)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		_ = port.SetReadTimeout(500 * time.Millisecond)
		line, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				return err
			}
			// Timeout is expected; keep reading.
			if len(line) == 0 {
				continue
			}
		}
		if len(line) == 0 {
			continue
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case out <- line:
		}
	}
}

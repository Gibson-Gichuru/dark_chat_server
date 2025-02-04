package pinger

import (
	"context"
	"darkchat/monitor"
	"io"
	"time"
)

const DEFAULTPINGINTERVAL = 30 * time.Second

var monitorLogger = monitor.New("pinger.log")

// Ping writes a "PING" message to the writer at regular intervals, given by the
// reset channel. If the reset channel is closed, Ping returns immediately.
// If the context is canceled, Ping returns immediately.
// If the writer returns an error, the error is logged to the monitorLogger.
//
// If the interval is zero, or becomes zero after a reset, the interval defaults
// to DEFAULTPINGINTERVAL.
//
// The function runs in its own goroutine, and does not block.
func Ping(ctx context.Context, w io.Writer, reset <-chan time.Duration) {
	var interval time.Duration

	select {
	case <-ctx.Done():
		return
	case interval = <-reset:
	default:
	}

	if interval <= 0 {
		interval = DEFAULTPINGINTERVAL
	}

	timer := time.NewTimer(interval)

	defer func() {
		if !timer.Stop() {
			<-timer.C
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case newInterval := <-reset:
			if !timer.Stop() {
				<-timer.C
			}
			if newInterval > 0 {
				interval = newInterval
			}
		case <-timer.C:
			if _, err := w.Write([]byte("PING")); err != nil {
				monitorLogger.Error(err.Error())
			}
		}
		_ = timer.Reset(interval)
	}
}

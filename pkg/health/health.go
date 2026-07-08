package health

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync/atomic"
	"time"
)

const (
	// LivenessPath is the HTTP path served for the liveness probe.
	LivenessPath = "/healthz"
	// ReadinessPath is the HTTP path served for the readiness probe.
	ReadinessPath = "/readyz"

	// shutdownTimeout bounds the graceful drain of the health server when
	// the context is cancelled.
	shutdownTimeout = 5 * time.Second
)

// Checker tracks a liveness heartbeat and an optional readiness gate.
type Checker struct {
	// lastBeat holds the timestamp of the most recent heartbeat.
	lastBeat atomic.Int64

	// staleAfter is how long the heartbeat may go un-refreshed before
	// liveness is reported as failed. It must be comfortably larger than the
	// heartbeat interval so ordinary scheduling jitter never trips it.
	staleAfter time.Duration

	// synced reports whether the provisioner's informer cache has completed
	// its initial sync. It gates readiness only; liveness never consults it.
	// A nil synced means "ready as soon as live".
	synced func() bool

	// now is overridable for tests; defaults to time.Now.
	now func() time.Time
}

// New returns a Checker whose liveness fails once the heartbeat is older than
// staleAfter, and whose readiness additionally requires synced() to be true.
// synced may be nil, in which case readiness tracks liveness.
//
// New records an initial heartbeat so the Checker reports live immediately;
// callers must still run StartHeartbeat to keep it fresh.
func New(staleAfter time.Duration, synced func() bool) *Checker {
	c := &Checker{
		staleAfter: staleAfter,
		synced:     synced,
		now:        time.Now,
	}
	c.Beat()
	return c
}

// Beat records a heartbeat at the current time.
func (c *Checker) Beat() {
	c.lastBeat.Store(c.now().UnixNano())
}

// StartHeartbeat beats immediately and then once every interval until ctx is
// cancelled. It is intended to run as its own goroutine.
func (c *Checker) StartHeartbeat(ctx context.Context, interval time.Duration) {
	c.Beat()
	t := time.NewTicker(interval)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			c.Beat()
		}
	}
}

// Live reports whether the heartbeat is fresh, plus a human-readable reason.
func (c *Checker) Live() (bool, string) {
	last := c.lastBeat.Load()
	if last == 0 {
		return false, "heartbeat not started"
	}
	age := c.now().Sub(time.Unix(0, last))
	if age > c.staleAfter {
		return false, fmt.Sprintf("heartbeat stale: last beat %s ago (threshold %s)",
			age.Truncate(time.Millisecond), c.staleAfter)
	}
	return true, "ok"
}

// Ready reports readiness: live, and (if configured) informer cache synced.
func (c *Checker) Ready() (bool, string) {
	if ok, reason := c.Live(); !ok {
		return false, reason
	}
	if c.synced != nil && !c.synced() {
		return false, "informer cache not synced"
	}
	return true, "ok"
}

// LivenessHandler serves the liveness probe: 200 when live, 503 otherwise.
func (c *Checker) LivenessHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		ok, reason := c.Live()
		writeProbe(w, ok, reason)
	}
}

// ReadinessHandler serves the readiness probe: 200 when ready, 503 otherwise.
func (c *Checker) ReadinessHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		ok, reason := c.Ready()
		writeProbe(w, ok, reason)
	}
}

func writeProbe(w http.ResponseWriter, ok bool, reason string) {
	if ok {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
	}
	// Body is advisory (kubelet ignores it); it aids `kubectl exec … wget`
	// and manual curl debugging.
	_, _ = fmt.Fprintln(w, reason)
}

// Handler returns an http.Handler exposing the liveness and readiness paths.
func (c *Checker) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.Handle(LivenessPath, c.LivenessHandler())
	mux.Handle(ReadinessPath, c.ReadinessHandler())
	return mux
}

// ListenAndServe runs the health HTTP server on addr until ctx is cancelled,
// then gracefully shuts it down. It blocks, so callers typically run it in a
// goroutine. A nil return means a clean shutdown; a non-nil return means the
// listener failed to start or serve.
func (c *Checker) ListenAndServe(ctx context.Context, addr string) error {
	srv := &http.Server{
		Addr:              addr,
		Handler:           c.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       30 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		err := srv.ListenAndServe()
		if errors.Is(err, http.ErrServerClosed) {
			err = nil
		}
		errCh <- err
	}()

	select {
	case err := <-errCh:
		// Server returned before context cancellation: startup/serve error.
		return err
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return <-errCh
	}
}

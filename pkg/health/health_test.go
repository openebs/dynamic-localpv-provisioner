package health

import (
	"context"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

// fakeClock lets tests advance time deterministically.
type fakeClock struct{ nanos atomic.Int64 }

func (f *fakeClock) Now() time.Time          { return time.Unix(0, f.nanos.Load()) }
func (f *fakeClock) Advance(d time.Duration) { f.nanos.Add(int64(d)) }

// newWithClock builds a Checker wired to a controllable clock.
func newWithClock(clk *fakeClock, staleAfter time.Duration, synced func() bool) *Checker {
	c := &Checker{staleAfter: staleAfter, synced: synced, now: clk.Now}
	c.Beat()
	return c
}

func TestLive_FreshHeartbeatIsLive(t *testing.T) {
	clk := &fakeClock{}
	clk.Advance(time.Hour) // move off the zero instant
	c := newWithClock(clk, 45*time.Second, nil)

	if ok, reason := c.Live(); !ok {
		t.Fatalf("expected live immediately after construction, got not-live: %s", reason)
	}

	// Just under the threshold: still live.
	clk.Advance(44 * time.Second)
	if ok, _ := c.Live(); !ok {
		t.Fatalf("expected live at 44s < 45s threshold")
	}
}

func TestLive_StaleHeartbeatFails(t *testing.T) {
	clk := &fakeClock{}
	clk.Advance(time.Hour)
	c := newWithClock(clk, 45*time.Second, nil)

	// Cross the threshold without a fresh beat.
	clk.Advance(46 * time.Second)
	ok, reason := c.Live()
	if ok {
		t.Fatalf("expected not-live once heartbeat is stale")
	}
	if reason == "ok" || reason == "" {
		t.Fatalf("expected a descriptive stale reason, got %q", reason)
	}

	// A fresh beat recovers liveness.
	c.Beat()
	if ok, _ := c.Live(); !ok {
		t.Fatalf("expected live again after a fresh beat")
	}
}

func TestReady_GatesOnSynced(t *testing.T) {
	clk := &fakeClock{}
	clk.Advance(time.Hour)

	synced := false
	c := newWithClock(clk, 45*time.Second, func() bool { return synced })

	// Live but not synced -> not ready.
	if ok, reason := c.Ready(); ok {
		t.Fatalf("expected not-ready while cache unsynced")
	} else if reason != "informer cache not synced" {
		t.Fatalf("unexpected reason: %q", reason)
	}

	// Synced -> ready.
	synced = true
	if ok, _ := c.Ready(); !ok {
		t.Fatalf("expected ready once synced")
	}
}

func TestReady_StaleHeartbeatOverridesSynced(t *testing.T) {
	clk := &fakeClock{}
	clk.Advance(time.Hour)
	c := newWithClock(clk, 45*time.Second, func() bool { return true })

	clk.Advance(2 * time.Minute) // heartbeat now stale
	if ok, _ := c.Ready(); ok {
		t.Fatalf("expected not-ready when heartbeat is stale even if synced")
	}
}

func TestReady_NilSyncedTracksLiveness(t *testing.T) {
	clk := &fakeClock{}
	clk.Advance(time.Hour)
	c := newWithClock(clk, 45*time.Second, nil)

	if ok, _ := c.Ready(); !ok {
		t.Fatalf("expected ready when synced is nil and heartbeat fresh")
	}
	clk.Advance(2 * time.Minute)
	if ok, _ := c.Ready(); ok {
		t.Fatalf("expected not-ready when heartbeat is stale")
	}
}

func TestHandlers_StatusCodes(t *testing.T) {
	clk := &fakeClock{}
	clk.Advance(time.Hour)
	synced := false
	c := newWithClock(clk, 45*time.Second, func() bool { return synced })

	// Liveness: 200 while fresh.
	if code := probe(t, c.Handler(), LivenessPath); code != http.StatusOK {
		t.Fatalf("liveness: want 200, got %d", code)
	}
	// Readiness: 503 while unsynced.
	if code := probe(t, c.Handler(), ReadinessPath); code != http.StatusServiceUnavailable {
		t.Fatalf("readiness (unsynced): want 503, got %d", code)
	}

	synced = true
	if code := probe(t, c.Handler(), ReadinessPath); code != http.StatusOK {
		t.Fatalf("readiness (synced): want 200, got %d", code)
	}

	// Liveness: 503 once heartbeat goes stale.
	clk.Advance(2 * time.Minute)
	if code := probe(t, c.Handler(), LivenessPath); code != http.StatusServiceUnavailable {
		t.Fatalf("liveness (stale): want 503, got %d", code)
	}
}

func probe(t *testing.T, h http.Handler, path string) int {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	_, _ = io.Copy(io.Discard, rec.Result().Body)
	_ = rec.Result().Body.Close()
	return rec.Code
}

// eventually polls cond until it returns true or the deadline elapses,
// reporting whether it succeeded. It replaces fixed sleeps in the real-clock
// tests: a slow scheduler (e.g. a loaded CI runner) only delays a success, it
// cannot turn a correct implementation into a failing test.
func eventually(within, tick time.Duration, cond func() bool) bool {
	deadline := time.Now().Add(within)
	for {
		if cond() {
			return true
		}
		if time.Now().After(deadline) {
			return false
		}
		time.Sleep(tick)
	}
}

func TestStartHeartbeat_RefreshesBeat(t *testing.T) {
	// staleAfter is huge and irrelevant here: this test observes the stored
	// heartbeat directly (same package) rather than through Live(), so it
	// proves the ticker refreshes the beat without depending on any real-time
	// staleness threshold. The stored beat only ever moves forward, so a
	// generous poll deadline absorbs scheduler delays without flaking.
	c := New(time.Hour, nil)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	constructorBeat := c.lastBeat.Load()
	go c.StartHeartbeat(ctx, 5*time.Millisecond)

	// The beat must advance past the constructor's beat...
	var firstTick int64
	if !eventually(2*time.Second, 5*time.Millisecond, func() bool {
		firstTick = c.lastBeat.Load()
		return firstTick > constructorBeat
	}) {
		t.Fatal("heartbeat never advanced the stored beat past construction")
	}
	// ...and keep advancing, proving it is periodic rather than a single beat.
	if !eventually(2*time.Second, 5*time.Millisecond, func() bool {
		return c.lastBeat.Load() > firstTick
	}) {
		t.Fatal("heartbeat advanced only once; expected periodic beats")
	}
}

func TestStartHeartbeat_StopsOnCancel(t *testing.T) {
	// Small staleAfter so post-cancel staleness is observable quickly. This is
	// not a tight assertion: once the heartbeat stops, staleness is monotonic,
	// so scheduler delays can only make Live() go false sooner, never later — a
	// generous poll deadline cannot flake.
	const staleAfter = 50 * time.Millisecond
	c := New(staleAfter, nil)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// First confirm the heartbeat is actually running (stored beat advances).
	constructorBeat := c.lastBeat.Load()
	go c.StartHeartbeat(ctx, 5*time.Millisecond)
	if !eventually(2*time.Second, 5*time.Millisecond, func() bool {
		return c.lastBeat.Load() > constructorBeat
	}) {
		t.Fatal("heartbeat never started")
	}

	// Stopping the heartbeat must let liveness go stale and stay stale.
	cancel()
	if !eventually(2*time.Second, 5*time.Millisecond, func() bool {
		ok, _ := c.Live()
		return !ok
	}) {
		t.Fatal("liveness never went stale after the heartbeat was cancelled")
	}
}

func TestListenAndServe_ServesThenShutsDown(t *testing.T) {
	// Reserve a real loopback port so the test can poll the endpoint until it
	// is serving, instead of sleeping a fixed interval to "wait for startup".
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("reserving a port: %v", err)
	}
	addr := ln.Addr().String()
	_ = ln.Close()

	c := New(45*time.Second, nil)
	ctx, cancel := context.WithCancel(context.Background())

	errCh := make(chan error, 1)
	go func() { errCh <- c.ListenAndServe(ctx, addr) }()

	// Poll until the server actually answers (no fixed startup sleep).
	url := "http://" + addr + LivenessPath
	if !eventually(2*time.Second, 10*time.Millisecond, func() bool {
		resp, err := http.Get(url)
		if err != nil {
			return false
		}
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
		return resp.StatusCode == http.StatusOK
	}) {
		cancel()
		t.Fatal("server did not start serving /healthz within the deadline")
	}

	// Cancellation must trigger a clean graceful shutdown.
	cancel()
	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("expected clean shutdown, got: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("ListenAndServe did not return after context cancellation")
	}
}

func TestListenAndServe_EndToEnd(t *testing.T) {
	c := New(45*time.Second, func() bool { return true })
	srv := httptest.NewServer(c.Handler())
	defer srv.Close()

	for _, tc := range []struct {
		path string
		want int
	}{
		{LivenessPath, http.StatusOK},
		{ReadinessPath, http.StatusOK},
	} {
		resp, err := http.Get(srv.URL + tc.path)
		if err != nil {
			t.Fatalf("GET %s: %v", tc.path, err)
		}
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
		if resp.StatusCode != tc.want {
			t.Fatalf("GET %s: want %d, got %d", tc.path, tc.want, resp.StatusCode)
		}
	}
}

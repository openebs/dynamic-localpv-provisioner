package logger

import (
	"context"
	"log"
	"time"

	"k8s.io/klog/v2"
)

// DefaultFlushInterval is klog's default flush interval.
const DefaultFlushInterval = 5 * time.Second

type KlogWriter struct{}

func (k KlogWriter) Write(data []byte) (n int, err error) {
	klog.Info(string(data))
	return len(data), nil
}

// InitLogging sets up logging with the specified flush interval.
// Redirects Go's log package to klog and attaches a logger to the context.
func InitLogging(ctx context.Context, flushInterval time.Duration) context.Context {
	log.SetOutput(KlogWriter{})
	log.SetFlags(0)

	klog.StartFlushDaemon(flushInterval)

	return klog.NewContext(ctx, klog.Background())
}

// FinishLogging flushes remaining logs and stops the flush daemon.
func FinishLogging() {
	klog.StopFlushDaemon()
	klog.Flush()
}

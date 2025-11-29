package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"k8s.io/klog/v2"

	"github.com/openebs/dynamic-localpv-provisioner/cmd/pvc-manager/app"
)

func main() {
	var (
		listenAddr = flag.String("listen-addr", ":8080", "The address to listen on for HTTP requests")
		logLevel   = flag.String("log-level", "info", "Log level (debug, info, warn, error)")
	)
	flag.Parse()

	// Initialize logger - klog doesn't have SetLevel, so we just note the desired level
	klog.Infof("Starting PVC Manager service with log level: %s", *logLevel)

	klog.Info("Starting PVC Manager service...")

	// Create and start the PVC manager server
	server := app.NewServer(*listenAddr)

	// Start server in a goroutine
	go func() {
		if err := server.Start(); err != nil {
			klog.Fatalf("Failed to start PVC Manager server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	klog.Info("Shutting down PVC Manager service...")
	if err := server.Stop(); err != nil {
		klog.Errorf("Error during server shutdown: %v", err)
	}
	klog.Info("PVC Manager service stopped")
}

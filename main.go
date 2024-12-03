package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// Create worker
	worker, err := NewWorker(Config{})
	if err != nil {
		log.Fatalf("failed to create worker: %v", err)
	}

	// Setup watching for exit signal
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-signalChan
		worker.Stop()
	}()

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Run worker (in a goroutine)
	// This will be stopped after either a termination signal or 10 seconds.
	go func() {
		err = worker.Start(ctx)
		if err != nil {
			log.Fatalf("failed to start worker: %v", err)
		}
	}()

	// Wait for worker to started and the stopped
	worker.WaitUntilStarted()
	worker.WaitUntilStopped()
}

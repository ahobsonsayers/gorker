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
	worker, err := NewWorker(Config{
		PollingPeriod: time.Second,
	})
	if err != nil {
		log.Fatalf("failed to create worker: %v", err)
	}

	// Setup signal handling
	signalChan := make(chan os.Signal, 1)
	signal.Notify(
		signalChan,
		syscall.SIGINT,  // Interrupt (Ctrl+C)
		syscall.SIGTERM, // Termination request
		syscall.SIGQUIT, // Quit (Ctrl+\)
	)
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

	time.Sleep(time.Millisecond)

	// Wait for worker to stop
	worker.WaitUntilStopped()
}

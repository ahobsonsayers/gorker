package main

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

type Config struct {
	PollingPeriod time.Duration
}

func (c *Config) Validate() error { // nolint: unparam
	if c.PollingPeriod <= 0 {
		return errors.New("polling period must be a valid positive duration")
	}

	return nil
}

type Result struct{}

type Worker struct {
	config Config

	// Result Channels
	resultChan chan Result
	errorChan  chan error

	// Synchronisation
	mu      *sync.Mutex
	startWg *sync.WaitGroup
	stopWg  *sync.WaitGroup
	cancel  context.CancelFunc
}

func (w *Worker) Results() <-chan Result { return w.resultChan }
func (w *Worker) Errors() <-chan error   { return w.errorChan }

func (w *Worker) IsRunning() bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.cancel != nil
}

func (w *Worker) WaitUntilStarted() { w.startWg.Wait() }
func (w *Worker) WaitUntilStopped() { w.stopWg.Wait() }

// setup can be used to add custom setup logic to the worker
func (*Worker) setup(_ context.Context) error { // nolint: unparam
	fmt.Println("Running setup")
	return nil
}

// process is your worker's main process
func (*Worker) process(_ context.Context) error { // nolint: unparam
	fmt.Println("Running process")

	currentTime := time.Now()
	fmt.Printf("Current Time: %v\n", currentTime)
	return nil
}

// Start the worker.
// This will block until the worker stopped.
func (w *Worker) Start(ctx context.Context) error {
	w.mu.Lock()

	// Check if already running
	if w.cancel != nil {
		w.mu.Unlock()
		return errors.New("worker already running")
	}

	// Run setup
	err := w.setup(ctx)
	if err != nil {
		w.mu.Unlock()
		return fmt.Errorf("setup failed: %w", err)
	}

	ticker := time.NewTicker(w.config.PollingPeriod)
	defer ticker.Stop()

	// Start worker go routine
	ctx, cancel := context.WithCancel(ctx)
	go func() {
		for {
			select {
			case <-ticker.C:
				err := w.process(ctx)
				if err != nil {
					fmt.Printf("process failed: %v\n", err)
				}

			case <-ctx.Done():
				w.mu.Lock()
				w.startWg.Add(1)
				w.stopWg.Done()
				w.cancel = nil
				w.mu.Unlock()
				return
			}
		}
	}()

	// Update fields
	w.cancel = cancel
	w.stopWg.Add(1)
	w.startWg.Done()

	w.mu.Unlock()

	w.stopWg.Wait()
	return nil
}

// Stop the worker.
// This will block until worker is fully stopped.
func (w *Worker) Stop() {
	w.mu.Lock()
	w.cancel()
	w.mu.Unlock()
	w.stopWg.Wait()
}

func NewWorker(config Config) (*Worker, error) {
	err := config.Validate()
	if err != nil {
		return nil, err
	}

	var startWg sync.WaitGroup
	startWg.Add(1)

	return &Worker{
		config: config,
		// Result Channels
		resultChan: make(chan Result),
		errorChan:  make(chan error),
		// Synchronisation
		mu:      new(sync.Mutex),
		startWg: &startWg,
		stopWg:  new(sync.WaitGroup),
	}, nil
}

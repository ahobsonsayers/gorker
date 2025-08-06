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

func (c *Config) Validate() error { //nolint:revive,unparam // we are using placeholders
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

func (s *Worker) Results() <-chan Result { return s.resultChan }
func (s *Worker) Errors() <-chan error   { return s.errorChan }

func (s *Worker) IsRunning() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.cancel != nil
}

func (s *Worker) WaitUntilStarted() { s.startWg.Wait() }
func (s *Worker) WaitUntilStopped() { s.stopWg.Wait() }

// setup can be used to add custom setup logic to the worker
func (s *Worker) setup(ctx context.Context) error { //nolint:revive,unparam // we are using placeholders
	fmt.Println("setup")
	return nil
}

// process is your workers main process
func (s *Worker) process(ctx context.Context) error { //nolint:revive,unparam // we are using placeholders
	fmt.Println("process")
	return nil
}

// Start the worker.
// This will block until the worker stopped.
func (s *Worker) Start(ctx context.Context) error {
	s.mu.Lock()

	// Check if already running
	if s.cancel != nil {
		s.mu.Unlock()
		return errors.New("worker already running")
	}

	// Run setup
	err := s.setup(ctx)
	if err != nil {
		s.mu.Unlock()
		return fmt.Errorf("setup failed: %w", err)
	}

	ticker := time.NewTicker(s.config.PollingPeriod)
	defer ticker.Stop()

	// Start worker go routine
	ctx, cancel := context.WithCancel(ctx)
	go func() {
		for {
			select {
			case <-ticker.C:
				fmt.Println("tick")

			case <-ctx.Done():
				s.mu.Lock()
				s.startWg.Add(1)
				s.stopWg.Done()
				s.cancel = nil
				s.mu.Unlock()
				return
			}
		}
	}()

	// Update fields
	s.cancel = cancel
	s.stopWg.Add(1)
	s.startWg.Done()

	s.mu.Unlock()

	s.stopWg.Wait()
	return nil
}

// Stop the worker.
// This will block until worker is fully stopped.
func (s *Worker) Stop() {
	s.mu.Lock()
	s.cancel()
	s.mu.Unlock()
	s.stopWg.Wait()
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

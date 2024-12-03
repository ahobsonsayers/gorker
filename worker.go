package main

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

var ErrRunning = errors.New("service already running")

type Config struct{}

func (c *Config) Validate() error { //nolint:revive,unparam // we are using placeholders
	return nil
}

type Result struct{}

type Service struct {
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

func (s *Service) Results() <-chan Result { return s.resultChan }
func (s *Service) Errors() <-chan error   { return s.errorChan }

func (s *Service) IsRunning() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.cancel != nil
}

func (s *Service) WaitUntilStarted() { s.startWg.Wait() }
func (s *Service) WaitUntilStopped() { s.stopWg.Wait() }

// setup can be used to add custom setup logic to the worker
func (s *Service) setup(ctx context.Context) error { //nolint:revive,unparam // we are using placeholders
	return nil
}

// Start the worker.
// This will block until the worker stopped.
func (s *Service) Start(ctx context.Context) error {
	s.mu.Lock()

	// Check if already running
	if s.cancel != nil {
		s.mu.Unlock()
		return ErrRunning
	}

	// Run setup
	err := s.setup(ctx)
	if err != nil {
		s.mu.Unlock()
		return fmt.Errorf("setup failed: %w", err)
	}

	ticker := time.NewTicker(time.Second)
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
func (s *Service) Stop() {
	s.mu.Lock()
	s.cancel()
	s.mu.Unlock()
	s.stopWg.Wait()
}

func NewWorker(config Config) (*Service, error) {
	err := config.Validate()
	if err != nil {
		return nil, err
	}

	var startWg sync.WaitGroup
	startWg.Add(1)

	return &Service{
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

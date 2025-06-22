package services

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/svbnbyrk/go-message-dispatcher/internal/core/ports/usecases"
)

// ProcessingScheduler handles automatic message processing at regular intervals
type ProcessingScheduler struct {
	messageProcessing usecases.MessageProcessingUseCase
	interval          time.Duration
	batchSize         int
	isRunning         bool
	stopChan          chan struct{}
	doneChan          chan struct{}
	mu                sync.RWMutex
	stats             *ProcessingStats
}

// ProcessingStats tracks processing statistics
type ProcessingStats struct {
	mu                    sync.RWMutex
	TotalProcessed        int64
	TotalSuccessful       int64
	TotalFailed           int64
	LastProcessingTime    time.Time
	LastProcessingResult  *usecases.ProcessingResult
	IsCurrentlyProcessing bool
}

// SchedulerConfig contains configuration for the processing scheduler
type SchedulerConfig struct {
	Interval  time.Duration
	BatchSize int
}

// NewProcessingScheduler creates a new processing scheduler
func NewProcessingScheduler(
	messageProcessing usecases.MessageProcessingUseCase,
	config SchedulerConfig,
) *ProcessingScheduler {
	return &ProcessingScheduler{
		messageProcessing: messageProcessing,
		interval:          config.Interval,
		batchSize:         config.BatchSize,
		stopChan:          make(chan struct{}),
		doneChan:          make(chan struct{}),
		stats: &ProcessingStats{
			LastProcessingTime: time.Now(),
		},
	}
}

// Start begins the background processing scheduler
func (s *ProcessingScheduler) Start(ctx context.Context) error {
	s.mu.Lock()
	if s.isRunning {
		s.mu.Unlock()
		return fmt.Errorf("scheduler is already running")
	}
	s.isRunning = true
	s.mu.Unlock()

	log.Printf("⏰ Starting message processing scheduler (interval: %v, batch size: %d)", s.interval, s.batchSize)

	go s.run(ctx)
	return nil
}

// Stop gracefully stops the processing scheduler
func (s *ProcessingScheduler) Stop() error {
	s.mu.Lock()
	if !s.isRunning {
		s.mu.Unlock()
		return fmt.Errorf("scheduler is not running")
	}
	s.mu.Unlock()

	log.Println("🛑 Stopping message processing scheduler...")
	close(s.stopChan)

	// Wait for processing to complete with timeout
	select {
	case <-s.doneChan:
		log.Println("✅ Processing scheduler stopped gracefully")
		return nil
	case <-time.After(30 * time.Second):
		log.Println("⚠️ Processing scheduler stop timed out")
		return fmt.Errorf("scheduler stop timed out")
	}
}

// IsRunning returns whether the scheduler is currently running
func (s *ProcessingScheduler) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.isRunning
}

// GetStats returns current processing statistics
func (s *ProcessingScheduler) GetStats() ProcessingStats {
	s.stats.mu.RLock()
	defer s.stats.mu.RUnlock()
	return *s.stats
}

// run is the main processing loop
func (s *ProcessingScheduler) run(ctx context.Context) {
	defer close(s.doneChan)
	defer func() {
		s.mu.Lock()
		s.isRunning = false
		s.mu.Unlock()
	}()

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	// Process immediately on start
	s.processMessages(ctx)

	for {
		select {
		case <-ctx.Done():
			log.Println("📤 Processing scheduler stopped due to context cancellation")
			return
		case <-s.stopChan:
			log.Println("📤 Processing scheduler received stop signal")
			return
		case <-ticker.C:
			s.processMessages(ctx)
		}
	}
}

// processMessages handles a single processing cycle
func (s *ProcessingScheduler) processMessages(ctx context.Context) {
	// Update processing status
	s.stats.mu.Lock()
	s.stats.IsCurrentlyProcessing = true
	s.stats.mu.Unlock()

	defer func() {
		s.stats.mu.Lock()
		s.stats.IsCurrentlyProcessing = false
		s.stats.LastProcessingTime = time.Now()
		s.stats.mu.Unlock()
	}()

	log.Printf("🔄 Processing pending messages (batch size: %d)...", s.batchSize)

	// Process messages
	result, err := s.messageProcessing.ProcessPendingMessages(ctx, s.batchSize)
	if err != nil {
		log.Printf("❌ Error processing messages: %v", err)
		return
	}

	// Update statistics
	s.stats.mu.Lock()
	s.stats.TotalProcessed += int64(result.ProcessedCount)
	s.stats.TotalSuccessful += int64(result.SuccessCount)
	s.stats.TotalFailed += int64(result.FailedCount)
	s.stats.LastProcessingResult = result
	s.stats.mu.Unlock()

	// Log results
	if result.ProcessedCount > 0 {
		log.Printf("✅ Processed %d messages: %d successful, %d failed",
			result.ProcessedCount, result.SuccessCount, result.FailedCount)

		if len(result.Errors) > 0 {
			log.Printf("⚠️ Processing errors:")
			for i, err := range result.Errors {
				log.Printf("   %d. %v", i+1, err)
			}
		}
	} else {
		log.Printf("💤 No pending messages to process")
	}
}

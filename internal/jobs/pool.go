package jobs

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"time"

	"github.com/sraj/everest/internal/apperror"
)

// Job is a background task that receives a context and returns an error.
type Job func(ctx context.Context) error

// Pool manages a fixed number of workers that process jobs from a queue.
// It supports graceful shutdown with a configurable drain timeout.
type Pool struct {
	queue       chan Job
	wg          sync.WaitGroup
	ctx         context.Context
	cancel      context.CancelFunc
	log         *slog.Logger
	maxAttempts int
}

// Config holds pool settings.
type Config struct {
	Workers    int // number of goroutines processing jobs
	QueueSize  int // buffered queue capacity
	MaxAttempts int // retry count for failed jobs (0 = no retry)
	Log        *slog.Logger
}

// DefaultConfig returns sensible defaults.
func DefaultConfig(log *slog.Logger) Config {
	return Config{
		Workers:     4,
		QueueSize:   100,
		MaxAttempts: 2,
		Log:         log,
	}
}

// New creates a worker pool and starts the configured number of goroutines.
func New(cfg Config) *Pool {
	ctx, cancel := context.WithCancel(context.Background())
	p := &Pool{
		queue:       make(chan Job, cfg.QueueSize),
		ctx:         ctx,
		cancel:      cancel,
		log:         cfg.Log,
		maxAttempts: cfg.MaxAttempts,
	}
	for i := 0; i < cfg.Workers; i++ {
		p.wg.Add(1)
		go p.worker(i)
	}
	return p
}

// Submit adds a job to the queue. If the pool is shutting down, the job is dropped.
func (p *Pool) Submit(job Job) {
	select {
	case p.queue <- job:
	case <-p.ctx.Done():
		p.log.Warn("job dropped — pool is shutting down")
	}
}

// Shutdown stops accepting new jobs and waits for all in-flight jobs to finish.
// Returns when all workers have exited or after the drain timeout.
func (p *Pool) Shutdown(timeout time.Duration) {
	p.cancel()
	done := make(chan struct{})
	go func() {
		p.wg.Wait()
		close(done)
	}()
	select {
	case <-done:
		p.log.Info("job pool shut down gracefully")
	case <-time.After(timeout):
		p.log.Warn("job pool shutdown timed out — some jobs may be incomplete")
	}
}

func (p *Pool) worker(id int) {
	defer p.wg.Done()
	for {
		select {
		case job := <-p.queue:
			p.runJob(job, 0)
		case <-p.ctx.Done():
			// Drain remaining jobs before exiting.
			for {
				select {
				case job := <-p.queue:
					p.runJob(job, 0)
				default:
					return
				}
			}
		}
	}
}

func (p *Pool) runJob(job Job, attempt int) {
	ctx, cancel := context.WithTimeout(p.ctx, 120*time.Second)
	defer cancel()

	err := job(ctx)
	if err == nil {
		return
	}

	// Don't retry validation/user errors.
	var ae *apperror.AppError
	if errors.As(err, &ae) && ae.Status >= 400 && ae.Status < 500 {
		return
	}

	if attempt < p.maxAttempts {
		p.log.Warn("job failed, retrying", "attempt", attempt+1, "error", err.Error())
		time.Sleep(time.Duration(attempt+1) * 500 * time.Millisecond)
		p.runJob(job, attempt+1)
	} else {
		p.log.Error("job failed after max attempts", "attempts", attempt, "error", err.Error())
	}
}

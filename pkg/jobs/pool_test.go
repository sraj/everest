package jobs

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"sync/atomic"
	"testing"
	"time"
)

func testLog() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
}

func TestSubmitAndExecute(t *testing.T) {
	var called atomic.Bool
	pool := New(Config{Workers: 1, QueueSize: 1, MaxAttempts: 0, Log: testLog()})
	pool.Submit(func(ctx context.Context) error { called.Store(true); return nil })
	pool.Shutdown(5 * time.Second)
	if !called.Load() {
		t.Fatal("job was not executed")
	}
}

func TestConcurrentWorkers(t *testing.T) {
	var count atomic.Int32
	pool := New(Config{Workers: 4, QueueSize: 100, MaxAttempts: 0, Log: testLog()})
	for i := 0; i < 20; i++ {
		pool.Submit(func(ctx context.Context) error {
			count.Add(1)
			time.Sleep(10 * time.Millisecond)
			return nil
		})
	}
	pool.Shutdown(5 * time.Second)
	if count.Load() != 20 {
		t.Fatalf("expected 20 executions, got %d", count.Load())
	}
}

func TestShutdownDrains(t *testing.T) {
	var count atomic.Int32
	pool := New(Config{Workers: 1, QueueSize: 100, MaxAttempts: 0, Log: testLog()})

	// Submit 5 fast jobs
	for i := 0; i < 5; i++ {
		pool.Submit(func(ctx context.Context) error { count.Add(1); return nil })
	}

	// Shutdown — should drain all 5
	pool.Shutdown(5 * time.Second)

	// Try to submit after shutdown — should be dropped
	pool.Submit(func(ctx context.Context) error { count.Add(1); return nil })

	if count.Load() != 5 {
		t.Fatalf("expected 5 jobs executed, got %d", count.Load())
	}
}

func TestShutdownTimeout(t *testing.T) {
	pool := New(Config{Workers: 1, QueueSize: 1, MaxAttempts: 0, Log: testLog()})
	pool.Submit(func(ctx context.Context) error {
		time.Sleep(10 * time.Second)
		return nil
	})
	start := time.Now()
	pool.Shutdown(50 * time.Millisecond)
	if elapsed := time.Since(start); elapsed > 500*time.Millisecond {
		t.Fatalf("shutdown took too long: %v", elapsed)
	}
}

func TestRetryOnError(t *testing.T) {
	var attempts atomic.Int32
	pool := New(Config{Workers: 1, QueueSize: 1, MaxAttempts: 2, ShouldRetry: func(error) bool { return true }, Log: testLog()})
	pool.Submit(func(ctx context.Context) error {
		attempts.Add(1)
		return errors.New("boom")
	})
	pool.Shutdown(5 * time.Second)
	// 1 initial + 2 retries = 3 total
	if attempts.Load() != 3 {
		t.Fatalf("expected 3 attempts, got %d", attempts.Load())
	}
}

func TestShouldRetryFalse(t *testing.T) {
	var attempts atomic.Int32
	pool := New(Config{Workers: 1, QueueSize: 1, MaxAttempts: 2, ShouldRetry: func(error) bool { return false }, Log: testLog()})
	pool.Submit(func(ctx context.Context) error {
		attempts.Add(1)
		return errors.New("boom")
	})
	pool.Shutdown(5 * time.Second)
	if attempts.Load() != 1 {
		t.Fatalf("expected 1 attempt (no retry), got %d", attempts.Load())
	}
}

func TestNoRetryWhenZero(t *testing.T) {
	var attempts atomic.Int32
	pool := New(Config{Workers: 1, QueueSize: 1, MaxAttempts: 0, Log: testLog()})
	pool.Submit(func(ctx context.Context) error {
		attempts.Add(1)
		return errors.New("boom")
	})
	pool.Shutdown(5 * time.Second)
	if attempts.Load() != 1 {
		t.Fatalf("expected 1 attempt (MaxAttempts=0), got %d", attempts.Load())
	}
}

func TestContextDeadline(t *testing.T) {
	pool := New(Config{Workers: 1, QueueSize: 1, MaxAttempts: 0, Log: testLog()})
	var deadline time.Time
	pool.Submit(func(ctx context.Context) error {
		var ok bool
		deadline, ok = ctx.Deadline()
		if !ok {
			t.Error("context has no deadline")
		}
		return nil
	})
	pool.Shutdown(5 * time.Second)
	if time.Until(deadline) > 120*time.Second || time.Until(deadline) < 110*time.Second {
		t.Fatalf("expected ~120s deadline, got %v", time.Until(deadline))
	}
}

func TestSubmitAfterShutdown(t *testing.T) {
	pool := New(Config{Workers: 1, QueueSize: 1, MaxAttempts: 0, Log: testLog()})
	pool.Shutdown(5 * time.Second)

	var called atomic.Bool
	pool.Submit(func(ctx context.Context) error { called.Store(true); return nil })
	time.Sleep(100 * time.Millisecond)

	if called.Load() {
		t.Fatal("job submitted after shutdown was unexpectedly executed")
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig(testLog())
	if cfg.Workers != 4 {
		t.Fatalf("expected 4 workers, got %d", cfg.Workers)
	}
	if cfg.QueueSize != 100 {
		t.Fatalf("expected 100 queue size, got %d", cfg.QueueSize)
	}
	if cfg.MaxAttempts != 2 {
		t.Fatalf("expected 2 max attempts, got %d", cfg.MaxAttempts)
	}
}

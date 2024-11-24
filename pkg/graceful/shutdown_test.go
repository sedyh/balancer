package graceful

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShutdownOrder(t *testing.T) {
	s := NewShutdown(React(ReactNope))

	closed := make([]string, 0, 3)

	s.Add(func(_ context.Context) error {
		closed = append(closed, "first")
		return nil
	})

	s.Add(func(_ context.Context) error {
		closed = append(closed, "second")
		return nil
	})

	s.Add(func(_ context.Context) error {
		closed = append(closed, "third")
		return nil
	})

	s.Stop()
	s.Wait()

	assert.Equal(t, closed, []string{"third", "second", "first"})
}

func TestShutdownOnError(t *testing.T) {
	var e error
	s := NewShutdown(
		React(func(err error) {
			e = err
		}),
	)

	handled := errors.New("fail to init connection")

	s.Add(func(_ context.Context) error {
		return handled
	})

	s.Stop()
	s.Wait()

	assert.Equal(t, e, handled)
}

func TestShutdownCloseTimeout(t *testing.T) {
	screen := 10 * time.Millisecond
	duration := 120 * time.Millisecond

	s := NewShutdown(
		CloseTimeout(duration),
		CancelTimeout(0),
		React(ReactNope),
	)

	s.Add(func(_ context.Context) error {
		<-time.After(300 * time.Second)
		return nil
	})

	now := time.Now()

	s.Stop()
	s.Wait()
	passed := time.Since(now)

	expected := float64(duration.Milliseconds())
	actual := float64(passed.Milliseconds())
	delta := float64(screen.Milliseconds())
	assert.InDelta(t, expected, actual, delta)
}

func TestShutdownCancelTimeout(t *testing.T) {
	screen := 10 * time.Millisecond
	duration := 100 * time.Millisecond
	cancel := 50 * time.Millisecond

	s := NewShutdown(
		CloseTimeout(duration),
		CancelTimeout(cancel),
		React(ReactNope),
	)

	var done int64

	s.Add(func(ctx context.Context) error {
		<-time.After(duration)
		<-ctx.Done()
		atomic.AddInt64(&done, 1)
		return nil
	})

	now := time.Now()

	s.Stop()
	s.Wait()
	passed := time.Since(now)

	expected := float64(duration.Milliseconds()) + float64(cancel.Milliseconds())
	actual := float64(passed.Milliseconds())
	delta := float64(screen.Milliseconds())
	assert.InDelta(t, expected, actual, delta)
	assert.Equal(t, atomic.LoadInt64(&done), int64(1))
}

func TestShutdownEnsureRunning(t *testing.T) {
	s := NewShutdown(
		CloseTimeout(120*time.Millisecond),
		React(ReactNope),
	)

	s.Add(func(_ context.Context) error {
		<-time.After(300 * time.Second)
		return nil
	})

	s.Stop()
	require.True(t, s.Done())

	s.Wait()
	require.False(t, s.Done())
}

func TestShutdownStopCause(t *testing.T) {
	var e error
	s := NewShutdown(
		React(func(err error) {
			e = err
		}),
	)

	s.Add(func(_ context.Context) error {
		<-time.After(300 * time.Second)
		return nil
	})

	cause := errors.New("fail to init connection")

	s.Stop(cause)
	require.Equal(t, cause, e)
}

package graceful

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

const (
	DefaultClosers       = 100
	DefaultCloseTimeout  = 1 * time.Second
	DefaultCancelTimeout = 200 * time.Millisecond
)

var (
	DefaultShutdown    = NewShutdown()
	DefaultSignals     = []os.Signal{syscall.SIGINT, syscall.SIGTERM}
	ErrShutdownTimeout = errors.New("closer completed with timeout")
)

type (
	ContextErrorCloser = func(ctx context.Context) error
	ContextCloser      = func(ctx context.Context)
	ErrorCloser        = func() error
	Closer             = func()
	Handler            = func(err error)
	Option             = func(s *Shutdown)
)

type closer struct {
	close         ContextErrorCloser
	closeTimeout  time.Duration
	cancelTimeout time.Duration
}

func (c closer) clone(close ContextErrorCloser) closer {
	return closer{
		close:         close,
		closeTimeout:  c.closeTimeout,
		cancelTimeout: c.cancelTimeout,
	}
}

type Shutdown struct {
	closers []closer
	parent  closer
	react   Handler
	counter atomic.Int64
	notify  chan os.Signal
	done    atomic.Bool
	stop    chan bool
	mu      sync.Mutex
}

func NewShutdown(options ...Option) *Shutdown {
	s := &Shutdown{
		closers: make([]closer, 0, DefaultClosers),
		parent: closer{
			closeTimeout:  DefaultCloseTimeout,
			cancelTimeout: DefaultCancelTimeout,
		},
		react:  ReactSlog,
		notify: make(chan os.Signal, 1),
		stop:   make(chan bool, 1),
	}
	signal.Notify(s.notify, DefaultSignals...)

	for _, op := range options {
		op(s)
	}

	return s
}

func (s *Shutdown) Stop(cause ...error) {
	if len(cause) > 0 {
		s.react(cause[0])
	}

	s.counter.Store(0)
	s.done.Store(true)
	s.stop <- true
}

func (s *Shutdown) Add(closer any) {
	s.mu.Lock()
	defer s.mu.Unlock()
	defer s.counter.Add(1)

	switch v := closer.(type) {
	case ContextErrorCloser:
		s.closers = append(s.closers, s.parent.clone(v))
	case ContextCloser:
		s.closers = append(s.closers, s.parent.clone(
			func(ctx context.Context) error {
				v(ctx)
				return nil
			},
		))
	case ErrorCloser:
		s.closers = append(s.closers, s.parent.clone(
			func(_ context.Context) error {
				return v()
			},
		))
	case Closer:
		s.closers = append(s.closers, s.parent.clone(
			func(_ context.Context) error {
				v()
				return nil
			},
		))
	default:
		s.react(fmt.Errorf("passed closer #%d will not be handler", s.counter.Load()))
	}
}

func (s *Shutdown) Wait() {
	select {
	case <-s.notify:
	case <-s.stop:
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for len(s.closers) > 0 {
		n := len(s.closers) - 1
		c := s.closers[n]
		s.close(c)
		s.closers = s.closers[:n]
	}

	s.closers = s.closers[:0]
	s.done.Store(false)
}

func (s *Shutdown) Done() bool {
	return s.done.Load()
}

func (s *Shutdown) Ensure() {
	if s.Done() {
		os.Exit(1)
	}
}

func (s *Shutdown) close(c closer) {
	ctx, cancel := context.WithTimeout(context.Background(), c.closeTimeout)
	defer cancel()

	done := make(chan bool, 1)
	go func() {
		defer close(done)

		if err := c.close(ctx); err != nil {
			s.react(err)
		}
	}()

	select {
	case <-ctx.Done():
		timer := time.NewTimer(c.cancelTimeout)
		defer timer.Stop()

		// Additional time for reaction after context cancel
		<-timer.C

		// Check again.
		select {
		case <-done:
			return
		default:
		}

		s.react(ErrShutdownTimeout)
	case <-done:
	}
}

func ReactSlog(err error) {
	slog.Error("shutdown", "error", err)
}

func ReactNope(_ error) {}

func Add(closer any) {
	DefaultShutdown.Add(closer)
}

func Wait() {
	DefaultShutdown.Wait()
}

func Stop(cause error) {
	DefaultShutdown.Stop(cause)
}

func Check(err error) {
	if err == nil {
		return
	}

	DefaultShutdown.Stop(err)
	DefaultShutdown.Ensure()
}

func Done() bool {
	return DefaultShutdown.Done()
}

func Ensure() {
	DefaultShutdown.Ensure()
}

func React(h Handler) Option {
	return func(s *Shutdown) {
		s.react = h
	}
}

func CloseTimeout(d time.Duration) Option {
	return func(s *Shutdown) {
		s.parent.closeTimeout = d
	}
}

func CancelTimeout(d time.Duration) Option {
	return func(s *Shutdown) {
		s.parent.cancelTimeout = d
	}
}

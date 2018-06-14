package circuit

import (
	"errors"
	"sync/atomic"
	"time"
)

var (
	ErrBreakerOpen    = errors.New("breaker open")
	ErrBreakerTimeout = errors.New("breaker time out")
)

// SimpleCircuitBreaker is window sliding CircuitBreaker with failure threshold.
type SimpleCircuitBreaker struct {
	lastFailureTime  time.Time
	failures         uint64
	failureThreshold uint64
	window           time.Duration
}

func NewSimpleCircuitBreaker(failureThreshold uint64, window time.Duration) *SimpleCircuitBreaker {
	return &SimpleCircuitBreaker{
		failureThreshold: failureThreshold,
		window:           window,
	}
}

// Call Circuit function
func (cb *SimpleCircuitBreaker) Call(fn func() error, d time.Duration) error {
	var err error

	if !cb.ready() {
		return ErrBreakerOpen
	}

	if d == 0 {
		err = fn()
	} else {
		c := make(chan error, 1)
		go func() {
			c <- fn()
			close(c)
		}()

		t := time.NewTimer(d)
		select {
		case e := <-c:
			err = e
		case <-t.C:
			err = ErrBreakerTimeout
		}
		t.Stop()
	}

	if err == nil {
		cb.success()
	} else {
		cb.fail()
	}

	return err
}

func (cb *SimpleCircuitBreaker) ready() bool {
	if time.Since(cb.lastFailureTime) > cb.window {
		cb.reset()
		return true
	}

	failures := atomic.LoadUint64(&cb.failures)
	return failures < cb.failureThreshold
}

func (cb *SimpleCircuitBreaker) success() {
	cb.reset()
}
func (cb *SimpleCircuitBreaker) fail() {
	atomic.AddUint64(&cb.failures, 1)
	cb.lastFailureTime = time.Now()
}

func (cb *SimpleCircuitBreaker) reset() {
	atomic.StoreUint64(&cb.failures, 0)
	cb.lastFailureTime = time.Now()
}

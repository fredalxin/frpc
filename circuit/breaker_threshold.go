package circuit

import (
	cir "github.com/rubyist/circuitbreaker"
	"time"
)

// SimpleCircuitBreaker is window sliding CircuitBreaker with failure threshold.
type ThresholdBreaker struct {
	breaker *cir.Breaker
}

// NewThresholdBreaker creates a Breaker with a ThresholdTripFunc.
func NewThresholdBreaker(threshold int64) Breaker {
	return &ThresholdBreaker{
		cir.NewThresholdBreaker(threshold),
	}
}

//// Call Circuit function
func (cb *ThresholdBreaker) Call(fn func() error, d time.Duration) error {
	return cb.breaker.Call(fn, d)
}


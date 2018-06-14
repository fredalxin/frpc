package circuit

import (
	cir "github.com/rubyist/circuitbreaker"
	"time"
)

// SimpleCircuitBreaker is window sliding CircuitBreaker with failure threshold.
type ConsecutiveBreaker struct {
	breaker *cir.Breaker
}

// NewThresholdBreaker creates a Breaker with a ThresholdTripFunc.
func NewConsecutiveBreaker(threshold int64) Breaker {
	return &ConsecutiveBreaker{
		cir.NewConsecutiveBreaker(threshold),
	}
}

//// Call Circuit function
func (cb *ConsecutiveBreaker) Call(fn func() error, d time.Duration) error {
	return cb.breaker.Call(fn, d)
}


package circuit

import (
	cir "github.com/rubyist/circuitbreaker"
	"time"
)

// SimpleCircuitBreaker is window sliding CircuitBreaker with failure threshold.
type RateBreaker struct {
	breaker *cir.Breaker
}

// NewThresholdBreaker creates a Breaker with a ThresholdTripFunc.
func NewRateBreaker(rate float64, minSamples int64) Breaker {
	return &RateBreaker{
		cir.NewRateBreaker(rate, minSamples),
	}
}

//// Call Circuit function
func (cb *RateBreaker) Call(fn func() error, d time.Duration) error {
	return cb.breaker.Call(fn, d)
}


package circuit

import "time"

type Breaker interface {
	Call(func() error, time.Duration) error
}

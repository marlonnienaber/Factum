package shared

import "time"

type Clock interface {
	Now() time.Time
}

type DefaultClock struct{}

func NewDefaultClock() *DefaultClock {
	return &DefaultClock{}
}

func (clock *DefaultClock) Now() time.Time {
	return time.Now().UTC()
}

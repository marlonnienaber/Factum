package support

import "time"

type TestingClock struct {
	currentTime time.Time
}

func NewTestingClock(currentTime time.Time) *TestingClock {
	return &TestingClock{currentTime: currentTime.UTC()}
}

func (clock *TestingClock) Now() time.Time {
	return clock.currentTime
}

func (clock *TestingClock) Advance(duration time.Duration) {
	clock.currentTime = clock.currentTime.Add(duration)
}

func (clock *TestingClock) Set(currentTime time.Time) {
	clock.currentTime = currentTime.UTC()
}

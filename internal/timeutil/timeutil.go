package timeutil

import (
	"sync"
	"time"
)

type Clock interface {
	Now() time.Time
}

func NewRealClock() Clock { return &realTimeClock{} }

type realTimeClock struct{}

func (_ *realTimeClock) Now() time.Time { return time.Now() }

type SimulatedClock struct {
	mu sync.Mutex
	t  time.Time // guarded by mu
}

func NewSimulatedClock(t time.Time) *SimulatedClock {
	return &SimulatedClock{t: t}
}

func (c *SimulatedClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.t
}

func (c *SimulatedClock) SetTime(t time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.t = t
}

func (c *SimulatedClock) AdvanceTime(d time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.t = c.t.Add(d)
}

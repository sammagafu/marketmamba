package decision

import (
	"sync"
	"time"
)

// CooldownTracker enforces sniper spacing per symbol (process-wide).
type CooldownTracker struct {
	mu       sync.Mutex
	lastTake map[string]time.Time
	cooldown time.Duration
}

func NewCooldownTracker(cooldown time.Duration) *CooldownTracker {
	if cooldown < time.Minute {
		cooldown = 45 * time.Minute
	}
	return &CooldownTracker{
		lastTake: make(map[string]time.Time),
		cooldown: cooldown,
	}
}

func (c *CooldownTracker) CanTake(symbol string) (bool, time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	t, ok := c.lastTake[symbol]
	if !ok {
		return true, 0
	}
	remaining := c.cooldown - time.Since(t)
	if remaining <= 0 {
		return true, 0
	}
	return false, remaining
}

func (c *CooldownTracker) RecordTake(symbol string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.lastTake[symbol] = time.Now()
}

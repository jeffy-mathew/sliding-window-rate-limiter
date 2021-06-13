package counter

import (
	"request-window-counter/internal/models"
	"request-window-counter/internal/persistence"
	"sync"
	"time"
)

// Counter service handles the hit counter
type Counter struct {
	// persistence is to load and dump the counter window to a csv file
	persistence persistence.Persistence
	mu          sync.Mutex
	// store the request window in the past 60 seconds
	window     []models.Entry
	hitCounter int64
}

// NewCounterService load data from csv file to window and returns a new counter service
func NewCounterService(persistence persistence.Persistence) (*Counter, error) {
	entries, totalHits, err := persistence.Load()
	if err != nil {
		return nil, err
	}
	// discard entries if last entry is before 60 seconds
	now := time.Now().Unix()
	entriesLength := len(entries)
	if entriesLength > 0 && entries[entriesLength-1].EpochTimestamp < now-60 {
		entries = []models.Entry{}
		totalHits = 0
	}
	return &Counter{
		persistence: persistence,
		mu:          sync.Mutex{},
		window:      entries,
		hitCounter:  totalHits,
	}, nil
}

// Hit handles the counter and returns the total number of hits received in the past 60 seconds
func (c *Counter) Hit() int64 {
	now := time.Now().Unix()
	sixtySecondsAgo := now - 60
	c.mu.Lock()
	var diff int64
	windowLength := len(c.window)
	discardCount := 0
	for i := 0; i < windowLength; i++ {
		if c.window[i].EpochTimestamp < sixtySecondsAgo {
			discardCount++
			diff -= c.window[i].Hits
		}
	}
	if windowLength > 0 && c.window[windowLength-1].EpochTimestamp == now {
		c.window[windowLength-1].Hits += 1
	} else {
		c.window = append(c.window[discardCount:windowLength], models.Entry{EpochTimestamp: now, Hits: 1})
	}
	diff++
	c.hitCounter = c.hitCounter + diff
	c.mu.Unlock()
	return c.hitCounter
}

// Dump dump the window to persistence
func (c *Counter) Dump() error {
	return c.persistence.Dump(c.window)
}

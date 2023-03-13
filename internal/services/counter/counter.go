package counter

import (
	"sync"
	"time"

	"github.com/jeffy-mathew/sliding-window-rate-limiter/internal/models"
)

// Counter service handles the hit counter
type Counter struct {
	windowSize int64
	mu         *sync.Mutex
	// store the request window in the past windowSize seconds
	window     []models.Entry
	hitCounter int64
}

// NewCounterService load data from csv file to window and returns a new counter service
func NewCounterService(windowSize int, entries []models.Entry) *Counter {
	// discard entries if last entry is before windowSize seconds
	now := time.Now().Unix()
	var totalHits int64 = 0
	entriesLength := len(entries)
	if entriesLength > 0 && entries[entriesLength-1].EpochTimestamp < now-int64(windowSize) {
		entries = []models.Entry{}
	} else {
		for _, entry := range entries {
			totalHits += entry.Hits
		}
	}

	// default value setting
	if windowSize == 0 {
		windowSize = 60
	}

	return &Counter{
		windowSize: int64(windowSize),
		mu:         &sync.Mutex{},
		window:     entries,
		hitCounter: totalHits,
	}
}

// Hit handles the counter and returns the total number of hits received in the past windowSize seconds
func (c *Counter) Hit() int64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	now := time.Now().Unix()
	c.discard(now)
	windowLength := len(c.window)
	if windowLength > 0 && c.window[windowLength-1].EpochTimestamp == now {
		c.window[windowLength-1].Hits += 1
	} else {
		c.window = append(c.window, models.Entry{EpochTimestamp: now, Hits: 1})
	}
	c.hitCounter = c.hitCounter + 1
	return c.hitCounter
}

func (c *Counter) Count() int64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	now := time.Now().Unix()
	c.discard(now)
	return c.hitCounter
}

func (c *Counter) Window() []models.Entry {
	c.mu.Lock()
	defer c.mu.Unlock()
	now := time.Now().Unix()
	c.discard(now)
	return c.window
}

func (c *Counter) discard(now int64) {
	windowLength := len(c.window)
	windowStart := now - c.windowSize
	var diff int64
	discardCount := 0
	for i := 0; i < windowLength; i++ {
		if c.window[i].EpochTimestamp < windowStart {
			discardCount++
			diff += c.window[i].Hits
		}
	}
	c.window = c.window[discardCount:windowLength]
	c.hitCounter = c.hitCounter - diff
}

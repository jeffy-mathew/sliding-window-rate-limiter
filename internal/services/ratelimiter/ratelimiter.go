package ratelimiter

import (
	"sync"

	"github.com/jeffy-mathew/sliding-window-rate-limiter/internal/models"
	"github.com/jeffy-mathew/sliding-window-rate-limiter/internal/persistence"
	"github.com/jeffy-mathew/sliding-window-rate-limiter/internal/services"
	"github.com/jeffy-mathew/sliding-window-rate-limiter/internal/services/counter"
)

// GlobalCounterKey is the key to store global rate counter.
const GlobalCounterKey = "GLOBAL"

// RateLimiter is the rate limiter, it decides whether to discard a request or not.
type RateLimiter struct {
	mu               sync.Mutex
	counters         map[string]services.CounterServiceInterface
	allowedRate      int64
	ipWindowSize     int
	globalWindowSize int
	// persistence is to load and dump the counter window to a json file
	persistence persistence.Persistence
}

// NewRateLimiter returns a RateLimiter with the provided configurations.
// globalWindowSize is windowSize for the global counter
// ipWindowSize is the windowSize for each IP counter.
// dataPersistence is the persistent storage.
func NewRateLimiter(globalWindowSize, ipWindowSize int, allowedRate int64, dataPersistence persistence.Persistence) (*RateLimiter, error) {
	ipCounterEntries, err := dataPersistence.Load()
	if err != nil {
		return nil, err
	}
	var counters = make(map[string]services.CounterServiceInterface)
	for ipAddr, entries := range ipCounterEntries {
		if ipAddr == GlobalCounterKey {
			counters[GlobalCounterKey] = counter.NewCounterService(globalWindowSize, entries)
			continue
		}
		counters[ipAddr] = counter.NewCounterService(ipWindowSize, entries)
	}

	// initialize global counter if not found
	if _, ok := ipCounterEntries[GlobalCounterKey]; !ok {
		counters[GlobalCounterKey] = counter.NewCounterService(globalWindowSize, []models.Entry{})
	}

	return &RateLimiter{
		mu:               sync.Mutex{},
		counters:         counters,
		allowedRate:      allowedRate,
		ipWindowSize:     ipWindowSize,
		globalWindowSize: globalWindowSize,
		persistence:      dataPersistence,
	}, nil
}

// Hit records a request and increments global counter and IP counter.
func (r *RateLimiter) Hit(ipAddr string) (int64, int64, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	globalHits := r.counters[GlobalCounterKey].Hit()
	ipHitCounter, ok := r.counters[ipAddr]
	if !ok {
		ipHitCounter = counter.NewCounterService(r.ipWindowSize, []models.Entry{})
		r.counters[ipAddr] = ipHitCounter
		ipHit := ipHitCounter.Hit()
		return globalHits, ipHit, false
	}

	ipHitSoFar := ipHitCounter.Count()
	if ipHitSoFar >= r.allowedRate {
		return globalHits, ipHitSoFar, true
	}

	return globalHits, ipHitCounter.Hit(), false
}

// Dump dumps current counter information to the underlying persistence storage.
func (r *RateLimiter) Dump() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	var counterEntries = make(map[string][]models.Entry)
	for ipAddr, ipCounter := range r.counters {
		entries := ipCounter.Window()
		if len(entries) > 0 {
			counterEntries[ipAddr] = entries
		}
	}

	return r.persistence.Dump(counterEntries)
}

package ratelimiter

import (
	"sliding-window-rate-limiter/internal/models"
	"sliding-window-rate-limiter/internal/persistence"
	"sliding-window-rate-limiter/internal/services"
	"sliding-window-rate-limiter/internal/services/counter"
	"sync"
)

const GlobalCounterKey = "GLOBAL"

type RateLimiter struct {
	mu               sync.Mutex
	counters         map[string]services.CounterServiceInterface
	allowedRate      int64
	ipWindowSize     int
	globalWindowSize int
	// persistence is to load and dump the counter window to a csv file
	persistence persistence.Persistence
}

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

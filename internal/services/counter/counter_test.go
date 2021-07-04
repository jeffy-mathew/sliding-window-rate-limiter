package counter

import (
	"github.com/stretchr/testify/assert"
	"request-window-counter/internal/models"
	"sync"
	"testing"
	"time"
)

func TestNewCounterService(t *testing.T) {
	t.Run("should return counter service successfully", func(t *testing.T) {
		counterService := NewCounterService(60, []models.Entry{})
		assert.NotEmpty(t, counterService)
	})
	t.Run("should return counter service with default window size 60", func(t *testing.T) {
		counterService := NewCounterService(0, []models.Entry{})
		assert.Equal(t, counterService.windowSize, int64(60))
	})
	t.Run("should set window to be empty slice when the latest entry from loaded entry is more than 60s ago -when windowsize is 60", func(t *testing.T) {
		epochNow := time.Now().Unix()
		counterService := NewCounterService(60, []models.Entry{{EpochTimestamp: epochNow - 70, Hits: int64(50)}})
		assert.Empty(t, counterService.window)
		assert.Zero(t, counterService.hitCounter)
	})
}

func TestCounter_Hit(t *testing.T) {
	t.Run("should include loaded count 50 seconds ago with latest hit", func(t *testing.T) {
		epochNow := time.Now().Unix()
		counterService := NewCounterService(60, []models.Entry{{EpochTimestamp: epochNow - 50, Hits: int64(50)}})
		newHit := counterService.Hit()
		var expectedHits int64 = 51
		assert.Equal(t, expectedHits, newHit)
	})
	t.Run("should discard loaded count 60 seconds ago", func(t *testing.T) {
		epochNow := time.Now().Unix()
		counterService := NewCounterService(60, []models.Entry{{EpochTimestamp: epochNow - 70, Hits: int64(50)}})
		newHit := counterService.Hit()
		var expectedHits int64 = 1
		assert.Equal(t, expectedHits, newHit)
	})
	t.Run("do concurrent requests and ensure the count is valid", func(t *testing.T) {
		epochNow := time.Now().Unix()
		counterService := NewCounterService(60, []models.Entry{{EpochTimestamp: epochNow - 30, Hits: int64(50)}})
		wg := sync.WaitGroup{}
		for i := 0; i < 20; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				counterService.Hit()
				counterService.Hit()
				counterService.Hit()
			}()
		}
		wg.Wait()
		count := counterService.Count()
		var expectedHits int64 = 110
		assert.Equal(t, expectedHits, count)
	})
}

func TestCounter_Window(t *testing.T) {
	t.Run("should discard old entries and return newer entries, reducing hit count", func(t *testing.T) {
		now := time.Now().Unix()
		counterService := Counter{
			windowSize: 20,
			mu:         &sync.Mutex{},
			window: []models.Entry{
				{EpochTimestamp: now - 40, Hits: 20},
				{EpochTimestamp: now - 35, Hits: 20},
				{EpochTimestamp: now - 18, Hits: 20},
				{EpochTimestamp: now - 15, Hits: 20},
				{EpochTimestamp: now - 10, Hits: 20},
			},
			hitCounter: 100,
		}
		entries := counterService.Window()
		expectedEntries := []models.Entry{
			{EpochTimestamp: now - 18, Hits: 20},
			{EpochTimestamp: now - 15, Hits: 20},
			{EpochTimestamp: now - 10, Hits: 20},
		}
		expectedHits := int64(60)
		assert.Equal(t, expectedEntries, entries)
		assert.Equal(t, expectedHits, counterService.hitCounter)
	})
}

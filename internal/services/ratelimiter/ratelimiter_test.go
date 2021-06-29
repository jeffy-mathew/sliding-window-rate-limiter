package ratelimiter

import (
	"errors"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"request-window-counter/internal/models"
	"request-window-counter/internal/persistence/persistence_mock"
	"request-window-counter/internal/services"
	"request-window-counter/internal/services/services_mock"

	//"request-window-counter/internal/services"
	//"request-window-counter/internal/services/services_mock"
	"sync"
	"testing"
	"time"
)

func TestNewRateLimiter(t *testing.T) {
	t.Run("should return error when persistence load fails", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockPersistence := persistence_mock.NewMockPersistence(ctrl)
		mockPersistence.EXPECT().Load().Return(nil, errors.New("something failed while loading persisted file"))
		rateLimiterService, err := NewRateLimiter(60, 20, 15, mockPersistence)
		assert.Error(t, err)
		assert.Nil(t, rateLimiterService)
	})

	t.Run("should return error when rateLimiter service successfully when entries are empty", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockPersistence := persistence_mock.NewMockPersistence(ctrl)
		mockPersistence.EXPECT().Load().Return(map[string][]models.Entry{}, nil)
		rateLimiterService, err := NewRateLimiter(60, 20, 15, mockPersistence)
		assert.Nil(t, err)
		assert.Equal(t, 60, rateLimiterService.globalWindowSize)
		assert.Equal(t, 20, rateLimiterService.ipWindowSize)
		assert.Equal(t, int64(15), rateLimiterService.allowedRate)
		assert.NotNil(t, rateLimiterService.counters)
	})
}

func TestRateLimiter_Hit(t *testing.T) {
	t.Run("should include loaded global count and discard loaded ip count  50 seconds ago with latest hit", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		epochNow := time.Now().Unix()
		ipAddr := "10.0.0.1"
		mockPersistence := persistence_mock.NewMockPersistence(ctrl)
		mockEntries := map[string][]models.Entry{GlobalCounterKey: {{EpochTimestamp: epochNow - 50, Hits: int64(50)}}, ipAddr: {{EpochTimestamp: epochNow - 50, Hits: int64(50)}}}
		mockPersistence.EXPECT().Load().Return(mockEntries, nil)
		rateLimiterService, err := NewRateLimiter(60, 20, 15, mockPersistence)
		assert.NoError(t, err)
		globalHit, ipHit, shouldDiscard := rateLimiterService.Hit(ipAddr)
		var expectedGlobalHits, ipHits int64 = 51, 1
		assert.Equal(t, expectedGlobalHits, globalHit)
		assert.Equal(t, ipHits, ipHit)
		assert.False(t, shouldDiscard)
	})

	t.Run("should include loaded ip count 15 seconds ago", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ipAddr := "10.0.0.1"
		epochNow := time.Now().Unix()
		mockPersistence := persistence_mock.NewMockPersistence(ctrl)
		mockEntries := map[string][]models.Entry{GlobalCounterKey: {{EpochTimestamp: epochNow - 50, Hits: int64(50)}}, ipAddr: {{EpochTimestamp: epochNow - 15, Hits: int64(10)}}}
		mockPersistence.EXPECT().Load().Return(mockEntries, nil)
		rateLimiterService, err := NewRateLimiter(60, 20, 15, mockPersistence)
		assert.NoError(t, err)
		globalHit, ipHit, shouldDiscard := rateLimiterService.Hit(ipAddr)
		var expectedGlobalHits, ipHits int64 = 51, 11
		assert.Equal(t, expectedGlobalHits, globalHit)
		assert.Equal(t, ipHits, ipHit)
		assert.False(t, shouldDiscard)
	})

	t.Run("should rateLimit with loaded ip count 15 seconds ago", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ipAddr := "10.0.0.1"
		epochNow := time.Now().Unix()
		mockPersistence := persistence_mock.NewMockPersistence(ctrl)
		mockEntries := map[string][]models.Entry{GlobalCounterKey: {{EpochTimestamp: epochNow - 50, Hits: int64(50)}}, ipAddr: {{EpochTimestamp: epochNow - 15, Hits: int64(15)}}}
		mockPersistence.EXPECT().Load().Return(mockEntries, nil)
		rateLimiterService, err := NewRateLimiter(60, 20, 15, mockPersistence)
		assert.NoError(t, err)
		globalHit, ipHit, shouldDiscard := rateLimiterService.Hit(ipAddr)
		var expectedGlobalHits, ipHits int64 = 51, 15
		assert.Equal(t, expectedGlobalHits, globalHit)
		assert.Equal(t, ipHits, ipHit)
		assert.True(t, shouldDiscard)
	})

	t.Run("do concurrent requests and ensure the rate limited for IP and valid count for global counter", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ipAddr := "10.0.0.1"
		mockPersistence := persistence_mock.NewMockPersistence(ctrl)
		mockEntries := map[string][]models.Entry{}
		mockPersistence.EXPECT().Load().Return(mockEntries, nil)
		rateLimiterService, err := NewRateLimiter(60, 20, 15, mockPersistence)
		assert.NoError(t, err)
		wg := sync.WaitGroup{}
		for i := 0; i < 20; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				rateLimiterService.Hit(ipAddr)
				rateLimiterService.Hit(ipAddr)
				rateLimiterService.Hit(ipAddr)
			}()
		}
		wg.Wait()
		globalHit, ipHit, shouldDiscard := rateLimiterService.Hit(ipAddr)
		var expectedGlobalHits, ipHits int64 = 61, 15
		assert.Equal(t, expectedGlobalHits, globalHit)
		assert.Equal(t, ipHits, ipHit)
		assert.True(t, shouldDiscard)
	})

	t.Run("do concurrent requests and ensure the rate limited for two IP addresses valid count for global counter", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ipAddr1 := "10.0.0.1"
		ipAddr2 := "10.0.0.2"
		mockPersistence := persistence_mock.NewMockPersistence(ctrl)
		mockEntries := map[string][]models.Entry{}
		mockPersistence.EXPECT().Load().Return(mockEntries, nil)
		rateLimiterService, err := NewRateLimiter(60, 20, 15, mockPersistence)
		assert.NoError(t, err)
		wg := sync.WaitGroup{}
		for i := 0; i < 20; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				rateLimiterService.Hit(ipAddr1)
				rateLimiterService.Hit(ipAddr2)
			}()
		}
		wg.Wait()
		globalHit, ipHit, shouldDiscard := rateLimiterService.Hit(ipAddr1)
		var expectedGlobalHits, ipHits int64 = 41, 15
		assert.Equal(t, expectedGlobalHits, globalHit)
		assert.Equal(t, ipHits, ipHit)
		assert.True(t, shouldDiscard)
		globalHit, ipHit, shouldDiscard = rateLimiterService.Hit(ipAddr2)
		expectedGlobalHits, ipHits = 42, 15
		assert.Equal(t, expectedGlobalHits, globalHit)
		assert.Equal(t, ipHits, ipHit)
		assert.True(t, shouldDiscard)
	})

	t.Run("do concurrent requests and ensure the not rate limited for two IP addresses when requested below threshold valid count for global counter", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ipAddr1 := "10.0.0.1"
		ipAddr2 := "10.0.0.2"
		mockPersistence := persistence_mock.NewMockPersistence(ctrl)
		mockEntries := map[string][]models.Entry{}
		mockPersistence.EXPECT().Load().Return(mockEntries, nil)
		rateLimiterService, err := NewRateLimiter(60, 20, 15, mockPersistence)
		assert.NoError(t, err)
		wg := sync.WaitGroup{}
		for i := 0; i < 5; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				rateLimiterService.Hit(ipAddr1)
				rateLimiterService.Hit(ipAddr2)
			}()
		}
		wg.Wait()
		globalHit, ipHit, shouldDiscard := rateLimiterService.Hit(ipAddr1)
		var expectedGlobalHits, ipHits int64 = 11, 6
		assert.Equal(t, expectedGlobalHits, globalHit)
		assert.Equal(t, ipHits, ipHit)
		assert.False(t, shouldDiscard)
		globalHit, ipHit, shouldDiscard = rateLimiterService.Hit(ipAddr2)
		expectedGlobalHits, ipHits = 12, 6
		assert.Equal(t, expectedGlobalHits, globalHit)
		assert.Equal(t, ipHits, ipHit)
		assert.False(t, shouldDiscard)
	})
}

func TestRateLimiter_Dump(t *testing.T) {
	t.Run("should empty map when there is no entries in each key", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockCounterService := services_mock.NewMockCounterServiceInterface(ctrl)
		mockCounterService.EXPECT().Window().Return([]models.Entry{})
		mockCounterEntries := map[string][]models.Entry{}
		mockPersistence := persistence_mock.NewMockPersistence(ctrl)
		mockPersistence.EXPECT().Dump(mockCounterEntries).Return(nil)
		rateLimiterService := RateLimiter{
			allowedRate:      15,
			globalWindowSize: 60,
			counters:         map[string]services.CounterServiceInterface{GlobalCounterKey: mockCounterService},
			ipWindowSize:     20,
			persistence:      mockPersistence,
			mu:               sync.Mutex{},
		}
		err := rateLimiterService.Dump()
		assert.NoError(t, err)
	})

	t.Run("should properly dump entries", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		now := time.Now().Unix()
		ipAddr := "10.0.0.1"
		mockCounterServiceGlobal := services_mock.NewMockCounterServiceInterface(ctrl)
		mockGlobalEntries := []models.Entry{{EpochTimestamp: now - 10, Hits: 60}}
		mockCounterServiceGlobal.EXPECT().Window().Return(mockGlobalEntries)
		mockCounterServiceIP := services_mock.NewMockCounterServiceInterface(ctrl)
		mockIPEntries := []models.Entry{{EpochTimestamp: now - 10, Hits: 15}}
		mockCounterServiceIP.EXPECT().Window().Return(mockIPEntries)
		mockCounterEntries := map[string][]models.Entry{GlobalCounterKey: mockGlobalEntries, ipAddr: mockIPEntries}
		mockPersistence := persistence_mock.NewMockPersistence(ctrl)
		mockPersistence.EXPECT().Dump(mockCounterEntries).Return(nil)
		rateLimiterService := RateLimiter{
			allowedRate:      15,
			globalWindowSize: 60,
			counters:         map[string]services.CounterServiceInterface{GlobalCounterKey: mockCounterServiceGlobal, ipAddr: mockCounterServiceIP},
			ipWindowSize:     20,
			persistence:      mockPersistence,
			mu:               sync.Mutex{},
		}
		err := rateLimiterService.Dump()
		assert.NoError(t, err)
	})

	t.Run("should return error when persistence returns error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		now := time.Now().Unix()
		ipAddr := "10.0.0.1"
		mockCounterServiceGlobal := services_mock.NewMockCounterServiceInterface(ctrl)
		mockGlobalEntries := []models.Entry{{EpochTimestamp: now - 10, Hits: 60}}
		mockCounterServiceGlobal.EXPECT().Window().Return(mockGlobalEntries)
		mockCounterServiceIP := services_mock.NewMockCounterServiceInterface(ctrl)
		mockIPEntries := []models.Entry{{EpochTimestamp: now - 10, Hits: 15}}
		mockCounterServiceIP.EXPECT().Window().Return(mockIPEntries)
		mockCounterEntries := map[string][]models.Entry{GlobalCounterKey: mockGlobalEntries, ipAddr: mockIPEntries}
		mockPersistence := persistence_mock.NewMockPersistence(ctrl)
		mockPersistence.EXPECT().Dump(mockCounterEntries).Return(errors.New("some error occurred while dumping"))
		rateLimiterService := RateLimiter{
			allowedRate:      15,
			globalWindowSize: 60,
			counters:         map[string]services.CounterServiceInterface{GlobalCounterKey: mockCounterServiceGlobal, ipAddr: mockCounterServiceIP},
			ipWindowSize:     20,
			persistence:      mockPersistence,
			mu:               sync.Mutex{},
		}
		err := rateLimiterService.Dump()
		assert.EqualError(t, err, "some error occurred while dumping")
	})

}

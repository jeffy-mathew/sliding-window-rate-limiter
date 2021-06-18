package counter

import (
	"errors"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"request-window-counter/internal/models"
	"request-window-counter/internal/persistence/persistence_mock"
	"sync"
	"testing"
	"time"
)

func TestNewCounterService(t *testing.T) {
	t.Run("should return error if loading persistence returned error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockPersistence := persistence_mock.NewMockPersistence(ctrl)
		defer ctrl.Finish()
		mockPersistence.EXPECT().Load().Return(nil, int64(0), errors.New("error while loading persistence"))
		counterService, err := NewCounterService(mockPersistence)
		assert.Error(t, err)
		assert.Nil(t, counterService)
	})
	t.Run("should return counter service when persistence load is successful", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockPersistence := persistence_mock.NewMockPersistence(ctrl)
		defer ctrl.Finish()
		mockPersistence.EXPECT().Load().Return([]models.Entry{}, int64(0), nil)
		counterService, err := NewCounterService(mockPersistence)
		assert.NoError(t, err)
		assert.NotEmpty(t, counterService)
	})
	t.Run("should set window to be empty slice when the latest entry from loaded entry is more than 60s ago", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockPersistence := persistence_mock.NewMockPersistence(ctrl)
		defer ctrl.Finish()
		epochNow := time.Now().Unix()
		mockEntries := []models.Entry{{EpochTimestamp: epochNow - 70, Hits: int64(50)}}
		mockPersistence.EXPECT().Load().Return(mockEntries, int64(50), nil)
		counterService, err := NewCounterService(mockPersistence)
		assert.NoError(t, err)
		assert.Empty(t, counterService.window)
		assert.Zero(t, counterService.hitCounter)
	})
}

func TestCounter_Hit(t *testing.T) {
	t.Run("should include loaded count 50 seconds ago with latest hit", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockPersistence := persistence_mock.NewMockPersistence(ctrl)
		defer ctrl.Finish()
		epochNow := time.Now().Unix()
		mockEntries := []models.Entry{{EpochTimestamp: epochNow - 50, Hits: int64(50)}}
		mockPersistence.EXPECT().Load().Return(mockEntries, int64(50), nil)
		counterService, err := NewCounterService(mockPersistence)
		assert.NoError(t, err)
		newHit := counterService.Hit()
		var expectedHits int64 = 51
		assert.Equal(t, expectedHits, newHit)
	})
	t.Run("should discard loaded count 60 seconds ago", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockPersistence := persistence_mock.NewMockPersistence(ctrl)
		defer ctrl.Finish()
		epochNow := time.Now().Unix()
		mockEntries := []models.Entry{{EpochTimestamp: epochNow - 70, Hits: int64(50)}}
		mockPersistence.EXPECT().Load().Return(mockEntries, int64(50), nil)
		counterService, err := NewCounterService(mockPersistence)
		assert.NoError(t, err)
		newHit := counterService.Hit()
		var expectedHits int64 = 1
		assert.Equal(t, expectedHits, newHit)
	})
	t.Run("do concurrent requests and ensure the count is valid", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockPersistence := persistence_mock.NewMockPersistence(ctrl)
		defer ctrl.Finish()
		epochNow := time.Now().Unix()
		mockEntries := []models.Entry{{EpochTimestamp: epochNow - 30, Hits: int64(50)}}
		mockPersistence.EXPECT().Load().Return(mockEntries, int64(50), nil)
		counterService, err := NewCounterService(mockPersistence)
		assert.NoError(t, err)
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
		newHit := counterService.Hit()
		var expectedHits int64 = 111
		assert.Equal(t, expectedHits, newHit)
	})
}

func TestCounter_Dump(t *testing.T) {
	t.Run("should call persistence dump with empty window", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockPersistence := persistence_mock.NewMockPersistence(ctrl)
		defer ctrl.Finish()
		mockPersistence.EXPECT().Load().Return([]models.Entry{}, int64(0), nil)
		mockPersistence.EXPECT().Dump([]models.Entry{})
		counterService, err := NewCounterService(mockPersistence)
		assert.NoError(t, err)
		err = counterService.Dump()
		assert.NoError(t, err)
	})
	t.Run("should call persistence dump with loaded entries immediately", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockPersistence := persistence_mock.NewMockPersistence(ctrl)
		defer ctrl.Finish()
		epochNow := time.Now().Unix()
		mockEntries := []models.Entry{{EpochTimestamp: epochNow - 50, Hits: int64(50)}}
		mockPersistence.EXPECT().Load().Return(mockEntries, int64(0), nil)
		mockPersistence.EXPECT().Dump(mockEntries)
		counterService, err := NewCounterService(mockPersistence)
		assert.NoError(t, err)
		err = counterService.Dump()
		assert.NoError(t, err)
	})
}

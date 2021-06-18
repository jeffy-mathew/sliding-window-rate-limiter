package app

import (
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"request-window-counter/internal/services/services_mock"
	"testing"
)

func TestApp_Dump(t *testing.T) {
	t.Run("should call service dump when app dump is called", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockService := services_mock.NewMockCounterServiceInterface(ctrl)
		mockService.EXPECT().Dump()
		counterApp := NewApp(mockService)
		err := counterApp.Dump()
		assert.NoError(t, err)
	})
}

func TestApp_Hit(t *testing.T) {
	t.Run("should call service hit on request", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockService := services_mock.NewMockCounterServiceInterface(ctrl)
		mockService.EXPECT().Hit().Return(int64(100))
		counterApp := NewApp(mockService)
		ts := httptest.NewServer(http.HandlerFunc(counterApp.Hit))
		defer ts.Close()
		resp, err := http.Get(ts.URL)
		assert.NoError(t, err)
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		assert.NoError(t, err)
		assert.Equal(t, "100", string(body))
	})
}

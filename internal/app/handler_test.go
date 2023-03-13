package app

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/jeffy-mathew/sliding-window-rate-limiter/internal/services/services_mock"
	"github.com/stretchr/testify/assert"
)

func TestApp_Dump(t *testing.T) {
	t.Run("should call service dump when app dump is called", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockService := services_mock.NewMockRateLimiterInterface(ctrl)
		mockService.EXPECT().Dump()
		counterApp := NewApp(mockService)
		err := counterApp.Dump()
		assert.NoError(t, err)
	})
}

func TestApp_Hit(t *testing.T) {
	t.Run("should call service hit on request", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		ipAddr := "10.0.0.1"
		mockService := services_mock.NewMockRateLimiterInterface(ctrl)
		mockService.EXPECT().Hit(ipAddr).Return(int64(100), int64(12), false)
		counterApp := NewApp(mockService)
		ts := httptest.NewServer(http.HandlerFunc(counterApp.Hit))
		defer ts.Close()
		req, err := http.NewRequest(http.MethodGet, ts.URL, nil)
		assert.NoError(t, err)
		req.Header.Add(IpAddrKey, ipAddr)
		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		assert.NoError(t, err)
		assert.Equal(t, fmt.Sprintf("global counter - %d, IP Counter - %d, rateLimited - %t", 100, 12, false), string(body))
		assert.Equal(t, 200, resp.StatusCode)
	})

	t.Run("should call service hit on request and return status code 429(too many requests) on rate limited requests", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		ipAddr := "10.0.0.1"
		mockService := services_mock.NewMockRateLimiterInterface(ctrl)
		mockService.EXPECT().Hit(ipAddr).Return(int64(100), int64(15), true)
		counterApp := NewApp(mockService)
		ts := httptest.NewServer(http.HandlerFunc(counterApp.Hit))
		defer ts.Close()
		req, err := http.NewRequest(http.MethodGet, ts.URL, nil)
		assert.NoError(t, err)
		req.Header.Add(IpAddrKey, ipAddr)
		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		assert.NoError(t, err)
		assert.Equal(t, fmt.Sprintf("global counter - %d, IP Counter - %d, rateLimited - %t", 100, 15, true), string(body))
		assert.Equal(t, 429, resp.StatusCode)
	})
}

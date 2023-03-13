package app

import (
	"fmt"
	"net/http"

	"github.com/jeffy-mathew/sliding-window-rate-limiter/internal/services"
)

// IpAddrKey is the header name in which IP address would be present.
const IpAddrKey = "IP_ADDR"

// App handles the hit and dump from high level
type App struct {
	rateLimiterService services.RateLimiterInterface
}

// NewApp returns app configured with passed counterService
func NewApp(rateLimiterService services.RateLimiterInterface) *App {
	return &App{
		rateLimiterService: rateLimiterService,
	}
}

// Hit is the http handler function for handling the request
func (a *App) Hit(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("panic recovered ", err)
		}
	}()
	ipAddr := r.Header.Get(IpAddrKey)
	globalCounter, ipCounter, discardRequest := a.rateLimiterService.Hit(ipAddr)
	if discardRequest {
		w.WriteHeader(http.StatusTooManyRequests)
	}
	fmt.Fprintf(w, "global counter - %d, IP Counter - %d, rateLimited - %t", globalCounter, ipCounter, discardRequest)
}

// Dump calls service dump to dump the window
func (a *App) Dump() error {
	return a.rateLimiterService.Dump()
}

package app

import (
	"fmt"
	"net/http"
	"request-window-counter/internal/services"
)

// App handles the hit and dump from high level
type App struct {
	counterService services.CounterServiceInterface
}

// NewApp returns app configured with passed counterService
func NewApp(counterService services.CounterServiceInterface) *App {
	return &App{
		counterService: counterService,
	}
}

// Hit is the http handler function for handling the request
func (a *App) Hit(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("panic recovered ", err)
		}
	}()
	fmt.Fprintf(w, "%d", a.counterService.Hit())
}

// Dump calls service dump to dump the window
func (a *App) Dump() error {
	return a.counterService.Dump()
}

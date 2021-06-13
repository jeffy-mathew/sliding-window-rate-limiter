// main package contains the driver code for running the application
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"request-window-counter/internal/app"
	"request-window-counter/internal/persistence/csvpersistence"
	"request-window-counter/internal/services/counter"
	"syscall"
	"time"
)

const (
	AppPortEnv = "APP_PORT"
	AppPort    = ":8000"
)

// serve handles the logic of running  server in a goroutine and waiting for signal to gracefully stop the server
// on ctx.Done signal a request to shutdown the server is sent, so that no new requests will be served
// after that the window is dumped to the file
func serve(ctx context.Context, counterApp *app.App) (err error) {
	mux := http.NewServeMux()
	mux.Handle("/", http.HandlerFunc(counterApp.Hit))
	port := os.Getenv(AppPortEnv)
	if port == "" {
		port = AppPort
	}
	srv := &http.Server{Addr: port, Handler: mux}
	go func() {
		if err = srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen:%s\n", err)
		}
	}()

	log.Printf("server started")

	<-ctx.Done()

	log.Printf("gracefull shutdown request received")

	ctxShutDown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		cancel()
	}()

	if err = srv.Shutdown(ctxShutDown); err != nil {
		log.Fatalf("server Shutdown Failed:%s", err.Error())
	}
	log.Println("application stopped accepting requests, dumping window")

	err = counterApp.Dump()
	if err != nil {
		log.Fatalln("dumping window failed", err.Error())
	}
	log.Println("dumping window complete. app exiting!!")

	return
}

// main initiates new app and calls serve to start the server
// it also spawns a goroutine to listen to os signals SIGINT or SIGTERM
// once the os signal is received the cancel func of ctx passed to serve is called
// notifying it to initiate a graceful shutdown
func main() {
	persistence, err := csvpersistence.NewPersistence()
	if err != nil {
		log.Fatalf("error while initilizing persistence %s", err.Error())
	}
	counterService, err := counter.NewCounterService(persistence)
	if err != nil {
		log.Fatalf("error while initilizing counter service %s", err.Error())
	}
	counterApp := app.NewApp(counterService)
	defer func() {
		if err := recover(); err != nil {
			log.Println("recovering from panic, dumping window")
			dumpErr := counterApp.Dump()
			if dumpErr != nil {
				log.Fatalln("dumping window failed", dumpErr.Error())
			}
			log.Println("dumping window complete. app exiting!!")
		}
	}()
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		<-c
		log.Printf("system call received")
		cancel()
	}()

	if err := serve(ctx, counterApp); err != nil {
		log.Printf("failed to serve:+%v\n", err)
	}
}

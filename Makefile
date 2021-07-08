all: prepare test build run

prepare:
	go mod download

test:
	go test -cover ./...

build:
	go build -o sliding-window-rate-limiter cmd/main.go

run:
	./sliding-window-rate-limiter
all: prepare test build run

prepare:
	go mod download

test:
	go test -cover ./...

build:
	go build -o request-window-counter cmd/main.go

run:
	./request-window-counter
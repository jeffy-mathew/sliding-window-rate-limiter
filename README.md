# Window rate counter

## Prerequisites
1. [Go 1.16](https://golang.org/dl/)

## How To Run

From project root directory run:

```sh
$ go build -o window-rate-counter cmd/main.go
$ ./window-rate-counter
```

The route is configured to `/` of the server

```sh
curl http://localhost:{APP_PORT}
```

The default value for `APP_PORT` is 8000.
It can be overridden by setting environment variable `APP_PORT` to the required port.
Make requests to this API to get the count of request received in the server in last 60 seconds

## How to Test
From project root directory run:
```sh
$ go test -cover ./...
```
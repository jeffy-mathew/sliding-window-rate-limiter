# Window rate counter

## Prerequisites
1. [Go 1.16](https://golang.org/dl/)

## How To Run

From project root directory run:

```sh
$ go build -o window-rate-counter cmd/main.go
$ ./window-rate-counter
```

Alternatively, you can run the application after running tests with a single command 
if [GNU Make](https://www.gnu.org/software/make/) is installed
```sh
$ make all
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

## Running with Docker & docker-compose

### Prerequisites
1. [docker](https://docs.docker.com/engine/install/)
2. [docker-compose](https://docs.docker.com/compose/install/)

### Instructions

Run
```sh
$ docker-compose up
```

This will build docker and run application in a docker container.
Port mapping is done from `8000:8000`
In case need to change the port on host, change the first argument to the required port, like `9000:8000`

# Sliding Window rate Limiter

This is an application implementing a sliding window rate limiter.

For the sake of simplicity it is considered that the source IP address is present as an HTTP request header `IP_ADDR`.

The application maintains a counter for requests on a global level and per IP level, though currently rate limiting is implemented only on IP level.

Global window is hardcoded to 60 and IP rate limit is set to 15 requests per 20 seconds - this could be environment variables to make the application flexible.

It has a persistence storage, so on the event of stopping the application, the current hit rates are persisted to a json file from `DUMP_FILE` environment variable, if it's not set it is defaulted to `./dump.json`. 
When the application is back up, the hit counter information are reloaded back to memory and the rate limiter can continue working. If the loaded data are too old(i.e. before the window length), the data is discarded.

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

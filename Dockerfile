FROM golang:1.16-alpine as builder
RUN cd ..
RUN mkdir sliding-window-rate-limiter
WORKDIR sliding-window-rate-limiter
COPY . ./
RUN CGO_ENABLED=0 GOOS=linux go build -mod=vendor -o sliding-window-rate-limiter ./cmd/main.go

FROM alpine
WORKDIR app
COPY --from=builder /go/sliding-window-rate-limiter/sliding-window-rate-limiter /app/
ENTRYPOINT ["/app/sliding-window-rate-limiter"]

FROM golang:1.16-alpine as builder
RUN cd ..
RUN mkdir request-window-counter
WORKDIR request-window-counter
COPY . ./
RUN CGO_ENABLED=0 GOOS=linux go build -mod=vendor -o request-window-counter ./cmd/main.go
RUN ls

FROM alpine
RUN mkdir request-window-counter
WORKDIR app
COPY --from=builder /go/request-window-counter/request-window-counter /app/
ENTRYPOINT ["/app/request-window-counter"]

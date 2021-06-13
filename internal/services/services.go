package services

//go:generate mockgen -source=services.go -destination=./services_mock/services_mock.go -package=services_mock CounterServiceInterface

// CounterServiceInterface handles the counter and dump for the application
type CounterServiceInterface interface {
	Hit() int64
	Dump() error
}

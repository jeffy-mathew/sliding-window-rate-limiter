package persistence

//go:generate mockgen -source=persistence.go -destination=./persistence_mock/persistence_mock.go -package=persistence_mock Persistence
import "github.com/jeffy-mathew/sliding-window-rate-limiter/internal/models"

// Persistence persists the entries, it also can load entries from persistence
type Persistence interface {
	Dump(counters map[string][]models.Entry) error
	Load() (map[string][]models.Entry, error)
}

package models

// Entry is each entry in the window
// Entry represents epoch timestamp and number of hits received in that second
type Entry struct {
	EpochTimestamp int64
	Hits           int64
}

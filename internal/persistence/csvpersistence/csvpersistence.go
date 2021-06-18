package csvpersistence

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"request-window-counter/internal/models"
	"strconv"
)

const (
	DumpFileEnv             = "DUMP_FILE"
	DefaultDumpFileLocation = "./dump.csv"
)

// CSVPersistence persists the entries to a csv file, mentioned in DumpFileEnv location
type CSVPersistence struct {
	file *os.File
}

func NewPersistence() (*CSVPersistence, error) {
	dumpFile := os.Getenv(DumpFileEnv)
	if dumpFile == "" {
		dumpFile = DefaultDumpFileLocation
	}
	file, err := os.OpenFile(dumpFile, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}
	return &CSVPersistence{
		file: file,
	}, nil
}

// Dump dumps the entries passed as a csv to the file
func (p *CSVPersistence) Dump(entries []models.Entry) error {
	defer p.file.Close()
	err := p.file.Truncate(0)
	if err != nil {
		log.Println("error while truncating file", err)
	}
	w := csv.NewWriter(p.file)
	var records [][]string
	for _, entry := range entries {
		records = append(records, []string{fmt.Sprintf("%d", entry.EpochTimestamp), fmt.Sprintf("%d", entry.Hits)})
	}
	err = w.WriteAll(records)
	if err != nil {
		log.Println("error occurred while dumping data to file", err)
	}
	return err
}

// Load loads the entries from persisted file
func (p *CSVPersistence) Load() ([]models.Entry, int64, error) {
	r := csv.NewReader(p.file)
	records, err := r.ReadAll()
	if err != nil {
		return nil, 0, fmt.Errorf("error while loading dump file %v", err)
	}
	var (
		entries   []models.Entry
		totalHits int64
	)
	for _, record := range records {
		epochTime, err := strconv.ParseInt(record[0], 10, 64)
		if err != nil {
			log.Println("error while converting string to epoch timestamp, skipping", err)
			continue
		}
		hits, err := strconv.ParseInt(record[1], 10, 64)
		if err != nil {
			log.Println("error while converting string to hits, skipping", err)
			continue
		}
		totalHits += hits
		entries = append(entries, models.Entry{EpochTimestamp: epochTime, Hits: hits})
	}
	return entries, totalHits, nil
}

package jsonpersistence

import (
	"bytes"
	"encoding/json"
	"log"
	"os"
	"request-window-counter/internal/models"
)

const (
	DumpFileEnv             = "DUMP_FILE"
	DefaultDumpFileLocation = "./dump.json"
)

// JSONPersistence persists the entries to a csv file, mentioned in DumpFileEnv location
type JSONPersistence struct {
	file *os.File
}

func NewPersistence() (*JSONPersistence, error) {
	dumpFile := os.Getenv(DumpFileEnv)
	if dumpFile == "" {
		dumpFile = DefaultDumpFileLocation
	}
	file, err := os.OpenFile(dumpFile, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}
	return &JSONPersistence{
		file: file,
	}, nil
}

// Dump dumps the counters to json file
func (p *JSONPersistence) Dump(counters map[string][]models.Entry) error {
	defer p.file.Close()
	err := p.file.Truncate(0)
	if err != nil {
		log.Println("error while truncating file", err)
	}
	countersJSON, err := json.Marshal(&counters)
	if err != nil {
		return err
	}
	_, err = p.file.Write(countersJSON)
	return err
}

// Load loads the entries from persisted file
func (p *JSONPersistence) Load() (map[string][]models.Entry, error) {
	buf := new(bytes.Buffer)
	n, err := buf.ReadFrom(p.file)
	if err != nil {
		return nil, err
	}
	if n == 0 {
		return map[string][]models.Entry{}, nil
	}
	var counters map[string][]models.Entry
	err = json.Unmarshal(buf.Bytes(), &counters)
	return counters, err
}

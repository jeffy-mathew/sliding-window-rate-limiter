package jsonpersistence

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/jeffy-mathew/sliding-window-rate-limiter/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestNewPersistence(t *testing.T) {
	t.Run("should return persistence with opened file", func(t *testing.T) {
		os.Setenv(DumpFileEnv, "./../../../testdata/dump-0.json")
		jsonPersistence, err := NewPersistence()
		assert.NoError(t, err)
		assert.NotNil(t, jsonPersistence.file)
		jsonPersistence.file.Close()
	})
	t.Run("should return persistence with opened file from default dump file file location", func(t *testing.T) {
		os.Setenv(DumpFileEnv, "")
		jsonPersistence, err := NewPersistence()
		assert.NoError(t, err)
		fileInfo, err := jsonPersistence.file.Stat()
		assert.NoError(t, err)
		assert.Equal(t, DefaultDumpFileLocation, "./"+fileInfo.Name())
		jsonPersistence.file.Close()
		os.Remove(fileInfo.Name())
	})
}

func TestJSONPersistence_Load(t *testing.T) {
	t.Run("should load entries successfully", func(t *testing.T) {
		os.Setenv(DumpFileEnv, "./../../../testdata/dump-0.json")
		globalKey, ipAddr1, ipAddr2, ipAddr3, ipAddr4 := "GLOBAL", "10.0.0.1", "10.0.0.2", "10.0.0.3", "10.0.0.4"
		jsonPersistence, err := NewPersistence()
		assert.NoError(t, err)
		entries, err := jsonPersistence.Load()
		assert.NoError(t, err)
		expectedGlobalEntriesCount, expectedIPAddr1Count, expectedIPAddr2Count, expectedIPAddr3Count, expectedIPAddr4Count := 33, 2, 2, 2, 2
		assert.Equal(t, len(entries[globalKey]), expectedGlobalEntriesCount)
		assert.Equal(t, len(entries[ipAddr1]), expectedIPAddr1Count)
		assert.Equal(t, len(entries[ipAddr2]), expectedIPAddr2Count)
		assert.Equal(t, len(entries[ipAddr3]), expectedIPAddr3Count)
		assert.Equal(t, len(entries[ipAddr4]), expectedIPAddr4Count)
	})
	t.Run("should fail with invalid json input", func(t *testing.T) {
		os.Setenv(DumpFileEnv, "./../../../testdata/dump-1.json")
		jsonPersistence, err := NewPersistence()
		assert.NoError(t, err)
		entries, err := jsonPersistence.Load()
		assert.NotNil(t, err)
		assert.Empty(t, entries)
	})
	t.Run("should load empty json without error", func(t *testing.T) {
		os.Setenv(DumpFileEnv, "./../../../testdata/dump-2.json")
		jsonPersistence, err := NewPersistence()
		assert.NoError(t, err)
		entries, err := jsonPersistence.Load()
		assert.NoError(t, err)
		assert.Empty(t, entries)
	})
	t.Run("should return error if file is closed", func(t *testing.T) {
		os.Setenv(DumpFileEnv, "./../../../testdata/dump-2.json")
		jsonPersistence, err := NewPersistence()
		assert.NoError(t, err)
		jsonPersistence.file.Close()
		entries, err := jsonPersistence.Load()
		assert.Error(t, err)
		assert.Empty(t, entries)
	})
}

func TestJSONPersistence_Dump(t *testing.T) {
	t.Run("should dump entries successfully", func(t *testing.T) {
		dumpFileLocation := "./../../../testdata/dump-3.json"
		os.Setenv(DumpFileEnv, dumpFileLocation)
		jsonPersistence, err := NewPersistence()
		assert.NoError(t, err)
		entries := map[string][]models.Entry{"GLOBAL": {
			{EpochTimestamp: 1623591925, Hits: 1},
			{EpochTimestamp: 1623591927, Hits: 1},
			{EpochTimestamp: 1623591928, Hits: 1},
			{EpochTimestamp: 1623591946, Hits: 1},
			{EpochTimestamp: 1623591947, Hits: 1},
			{EpochTimestamp: 1623591948, Hits: 2},
			{EpochTimestamp: 1623591949, Hits: 1},
			{EpochTimestamp: 1623591950, Hits: 2},
			{EpochTimestamp: 1623591951, Hits: 1},
			{EpochTimestamp: 1623591952, Hits: 2},
			{EpochTimestamp: 1623591953, Hits: 1},
			{EpochTimestamp: 1623591954, Hits: 1},
			{EpochTimestamp: 1623591969, Hits: 1},
		}}
		err = jsonPersistence.Dump(entries)
		assert.NoError(t, err)
		dumpedFile, err := os.Open(dumpFileLocation)
		assert.NoError(t, err)
		defer dumpedFile.Close()
		data, err := ioutil.ReadAll(dumpedFile)
		assert.NoError(t, err)
		expectedFileOut := `{"GLOBAL":[{"epoch_timestamp":1623591925,"hits":1},{"epoch_timestamp":1623591927,"hits":1},{"epoch_timestamp":1623591928,"hits":1},{"epoch_timestamp":1623591946,"hits":1},{"epoch_timestamp":1623591947,"hits":1},{"epoch_timestamp":1623591948,"hits":2},{"epoch_timestamp":1623591949,"hits":1},{"epoch_timestamp":1623591950,"hits":2},{"epoch_timestamp":1623591951,"hits":1},{"epoch_timestamp":1623591952,"hits":2},{"epoch_timestamp":1623591953,"hits":1},{"epoch_timestamp":1623591954,"hits":1},{"epoch_timestamp":1623591969,"hits":1}]}`
		assert.Equal(t, expectedFileOut, string(data))
	})
	t.Run("should return error when truncate file fails", func(t *testing.T) {
		dumpFileLocation := "./../../../testdata/dump-3.json"
		os.Setenv(DumpFileEnv, dumpFileLocation)
		jsonPersistence, err := NewPersistence()
		assert.NoError(t, err)
		entries := map[string][]models.Entry{"GLOBAL": {
			{EpochTimestamp: 1623591925, Hits: 1},
			{EpochTimestamp: 1623591927, Hits: 1},
			{EpochTimestamp: 1623591928, Hits: 1},
			{EpochTimestamp: 1623591946, Hits: 1},
			{EpochTimestamp: 1623591947, Hits: 1},
			{EpochTimestamp: 1623591948, Hits: 2},
			{EpochTimestamp: 1623591949, Hits: 1},
			{EpochTimestamp: 1623591950, Hits: 2},
			{EpochTimestamp: 1623591951, Hits: 1},
			{EpochTimestamp: 1623591952, Hits: 2},
			{EpochTimestamp: 1623591953, Hits: 1},
			{EpochTimestamp: 1623591954, Hits: 1},
			{EpochTimestamp: 1623591969, Hits: 1},
		}}
		jsonPersistence.file.Close()
		err = jsonPersistence.Dump(entries)
		assert.Error(t, err)
	})
}

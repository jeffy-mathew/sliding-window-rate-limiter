package csvpersistence

import (
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"request-window-counter/internal/models"
	"testing"
)

func TestCSVPersistence_Load(t *testing.T) {
	t.Run("should load entries successfully", func(t *testing.T) {
		os.Setenv(DumpFileEnv, "./../../../testdata/dump-0.csv")
		csvPersistence, err := NewPersistence()
		assert.NoError(t, err)
		entries, totalHits, err := csvPersistence.Load()
		assert.NoError(t, err)
		expectedEntries := []models.Entry{
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
		}
		var expectedTotalHits int64 = 16
		assert.Equal(t, expectedEntries, entries)
		assert.Equal(t, expectedTotalHits, totalHits)
	})
	t.Run("should fail with invalid csv input", func(t *testing.T) {
		os.Setenv(DumpFileEnv, "./../../../testdata/dump-1.csv")
		csvPersistence, err := NewPersistence()
		assert.NoError(t, err)
		_, _, err = csvPersistence.Load()
		assert.Error(t, err)
	})
	t.Run("should skip entry when invalid input for epoch time", func(t *testing.T) {
		os.Setenv(DumpFileEnv, "./../../../testdata/dump-2.csv")
		csvPersistence, err := NewPersistence()
		assert.NoError(t, err)
		entries, totalHits, err := csvPersistence.Load()
		assert.NoError(t, err)
		expectedEntries := []models.Entry{
			{EpochTimestamp: 1623591925, Hits: 1},
			{EpochTimestamp: 1623591927, Hits: 1},
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
		}

		var expectedTotalHits int64 = 15
		assert.Equal(t, expectedEntries, entries)
		assert.Equal(t, expectedTotalHits, totalHits)
	})
	t.Run("should skip entry when invalid input for hits", func(t *testing.T) {
		os.Setenv(DumpFileEnv, "./../../../testdata/dump-3.csv")
		csvPersistence, err := NewPersistence()
		assert.NoError(t, err)
		entries, totalHits, err := csvPersistence.Load()
		assert.NoError(t, err)
		expectedEntries := []models.Entry{
			{EpochTimestamp: 1623591925, Hits: 1},
			{EpochTimestamp: 1623591927, Hits: 1},
			{EpochTimestamp: 1623591928, Hits: 1},
			{EpochTimestamp: 1623591946, Hits: 1},
			{EpochTimestamp: 1623591947, Hits: 1},
			{EpochTimestamp: 1623591949, Hits: 1},
			{EpochTimestamp: 1623591950, Hits: 2},
			{EpochTimestamp: 1623591951, Hits: 1},
			{EpochTimestamp: 1623591952, Hits: 2},
			{EpochTimestamp: 1623591953, Hits: 1},
			{EpochTimestamp: 1623591954, Hits: 1},
			{EpochTimestamp: 1623591969, Hits: 1},
		}
		var expectedTotalHits int64 = 14
		assert.Equal(t, expectedEntries, entries)
		assert.Equal(t, expectedTotalHits, totalHits)
	})
}

func TestCSVPersistence_Dump(t *testing.T) {
	t.Run("should dump entries successfully", func(t *testing.T) {
		dumpFileLocation := "./../../../testdata/dump-4.csv"
		os.Setenv(DumpFileEnv, dumpFileLocation)
		csvPersistence, err := NewPersistence()
		assert.NoError(t, err)
		entries := []models.Entry{
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
		}
		err = csvPersistence.Dump(entries)
		assert.NoError(t, err)
		dumpedFile, err := os.Open(dumpFileLocation)
		assert.NoError(t, err)
		defer dumpedFile.Close()
		data, err := ioutil.ReadAll(dumpedFile)
		assert.NoError(t, err)
		expectedFileOut := "1623591925,1\n1623591927,1\n1623591928,1\n1623591946,1\n1623591947,1\n1623591948,2\n1623591949,1\n1623591950,2\n1623591951,1\n1623591952,2\n1623591953,1\n1623591954,1\n1623591969,1\n"
		assert.Equal(t, expectedFileOut, string(data))
	})
}

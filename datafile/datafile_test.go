package datafile

import (
	"key_value_db/record"
	"os"
	"path/filepath"
	"testing"
)

func TornTail(t *testing.T) {
	path := filepath.Join(t.TempDir(), "log")
	lg, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}

	good := record.Encode(record.Record{TimeStamp: 1, Key: []byte("a"), Value: []byte("1")})
	lg.Append(good)
	lg.Append(good[:len(good)-2]) //crash testing simulation mid-write crash
	lg.Close()

	lg, _ = Open(path)
	defer lg.Close()

	count := 0
	validEnd, err := lg.Scan(func(pos int64, r record.Record) error {
		count++
		return nil
	})

	if err != nil {
		t.Fatal(err)
	}

	if count != 1 {
		t.Errorf("records: got %d, want 1", count)
	}

	if validEnd != int64(len(good)) {
		t.Errorf("validEnd: got %d, want %d", validEnd, len(good))
	}

	_ = os.Remove
}

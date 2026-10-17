package store

import (
	"errors"
	"os"
	"sync"

	"key_value_db/datafile"
	"key_value_db/index"
	"key_value_db/record"
)

var (
	ErrNotFound = errors.New("store: key not found")
	ErrEmptyKey = errors.New("store: key must not be empty")
	ErrTooLarge = errors.New("store: value too large")
)

type Store struct {
	mu    sync.RWMutex
	path  string
	log   *datafile.Log
	index *index.Index
}

// Open opens the store at path, rebuilding the index from the log.
func Open(path string) (*Store, error) {
	// A leftover .compact file means an earlier compaction didn't finish.
	// The original file is still the source of truth, so discard the partial copy.
	os.Remove(path + ".compact")

	lg, err := datafile.Open(path)
	if err != nil {
		return nil, err
	}
	idx := index.New()

	validEnd, err := lg.Scan(func(pos int64, r record.Record) error {
		key := string(r.Key)
		if r.Flag == record.FlagTombstone {
			idx.Delete(key)
			return nil
		}
		idx.Set(key, index.Entry{
			Pos:  pos + record.HeaderSize + int64(len(r.Key)),
			Size: uint32(len(r.Value)),
		})
		return nil
	})
	if err != nil {
		lg.Close()
		return nil, err
	}

	if validEnd < lg.Size() {
		if err := lg.Truncate(validEnd); err != nil {
			lg.Close()
			return nil, err
		}
	}

	return &Store{path: path, log: lg, index: idx}, nil
}

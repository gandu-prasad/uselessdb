package store

import (
	"errors"
	"os"
	"path/filepath"
	"time"

	"key_value_db/datafile"
	"key_value_db/index"
	"key_value_db/record"
)

// Compact rewrites the log so it holds only the latest value of each live key.
func (s *Store) Compact() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	tmpPath := s.path + ".compact"
	if err := os.Remove(tmpPath); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	newLog, err := datafile.Open(tmpPath)
	if err != nil {
		return err
	}
	newIdx := index.New()

	// Phase 1: copy each live value into the new file.
	var copyErr error
	s.index.Range(func(key string, e index.Entry) bool {
		val, err := s.log.ReadValue(e.Pos, e.Size)
		if err != nil {
			copyErr = err
			return false
		}
		buf := record.Encode(record.Record{
			TimeStamp: uint32(time.Now().Unix()),
			Flag:      record.FlagValue,
			Key:       []byte(key),
			Value:     val,
		})
		start, err := newLog.Append(buf)
		if err != nil {
			copyErr = err
			return false
		}
		newIdx.Set(key, index.Entry{
			Pos:  start + record.HeaderSize + int64(len(key)),
			Size: e.Size,
		})
		return true
	})
	if copyErr == nil {
		copyErr = newLog.Sync() // the new file must be on disk before it replaces the old one
	}
	if copyErr != nil {
		newLog.Close()
		os.Remove(tmpPath)
		return copyErr // the old file is untouched
	}

	// Phase 2: swap the files. Every write was already synced, so the old
	// file's Close error doesn't affect the data we're keeping.
	_ = s.log.Close()

	if err := os.Rename(tmpPath, s.path); err != nil {
		newLog.Close()
		os.Remove(tmpPath)
		// The old file is still in place. Reopen it so the store keeps working.
		old, reopenErr := datafile.Open(s.path)
		if reopenErr != nil {
			return errors.Join(err, reopenErr)
		}
		s.log = old
		return err
	}

	// The new file is in place. Switch to it. newLog stays open because on
	// Linux and macOS an open file handle follows the file through a rename.
	s.log = newLog
	s.index = newIdx

	// Make the rename itself durable. If the machine loses power before this
	// completes, the old file is still intact, so the store is consistent either way.
	return syncDir(filepath.Dir(s.path))
}

// syncDir fsyncs a directory so that renames inside it survive a crash.
func syncDir(dir string) error {
	d, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer d.Close()
	return d.Sync()
}

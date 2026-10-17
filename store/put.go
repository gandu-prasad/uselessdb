package store

import (
	"key_value_db/index"
	"key_value_db/record"
	"math"
	"time"
)

func (s *Store) appendRecord(key, value []byte, flag record.Flag) (int64, error) {
	buf := record.Encode(record.Record{
		TimeStamp: uint32(time.Now().Unix()),
		Flag:      flag,
		Key:       key,
		Value:     value,
	})

	start, err := s.log.Append(buf)

	if err != nil {
		return 0, err
	}

	if err := s.log.Sync(); err != nil {
		return 0, err
	}
	return start, nil
}

func (s *Store) Put(key, value []byte) error {
	if len(key) == 0 {
		return ErrEmptyKey
	}
	if uint64(len(value)) > math.MaxUint32 {
		return ErrTooLarge
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	start, err := s.appendRecord(key, value, record.FlagValue)
	if err != nil {
		return err
	}

	s.index.Set(string(key), index.Entry{
		Pos:  start + record.HeaderSize + int64(len(key)),
		Size: uint32(len(value)),
	})

	return nil

}

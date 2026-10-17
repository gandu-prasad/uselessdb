package store

import "key_value_db/record"

func (s *Store) Delete(key []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.index.Get(string(key)); !ok {
		return ErrNotFound
	}

	if _, err := s.appendRecord(key, nil, record.FlagTombstone); err != nil {
		return err
	}

	s.index.Delete(string(key))
	return nil
}

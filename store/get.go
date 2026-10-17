package store

func (s *Store) Get(key []byte) ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	e, ok := s.index.Get(string(key))
	if !ok {
		return nil, ErrNotFound
	}
	return s.log.ReadValue(e.Pos, e.Size)
}

package datafile

import "key_value_db/record"

// it  reads n bytes starting at pos the store uses this in Get after looking up the pos in index
func (l *Log) ReadValue(pos int64, n uint32) ([]byte, error) {
	buf := make([]byte, n)

	if _, err := l.f.ReadAt(buf, pos); err != nil {
		return nil, err
	}

	return buf, nil
}

// reads and decodes the full record that starts at pos, returns the record and it's size in bytes
func (l *Log) ReadRecord(pos int64) (record.Record, int, error) {
	header := make([]byte, record.HeaderSize)

	if _, err := l.f.ReadAt(header, pos); err != nil {
		return record.Record{}, 0, err
	}

	total, err := record.TotalSize(header)

	if err != nil {
		return record.Record{}, 0, err
	}

	buf := make([]byte, total)

	if _, err := l.f.ReadAt(buf, pos); err != nil {
		return record.Record{}, 0, err
	}

	return record.Decode(buf)
}

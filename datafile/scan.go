package datafile

import (
	"errors"
	"io"
	"key_value_db/record"
)

func (l *Log) Scan(fn func(pos int64, r record.Record) error) (int64, error) {
	var pos int64

	header := make([]byte, record.HeaderSize)

	for {
		n, err := l.f.ReadAt(header, pos)
		if n < len(header) {
			if err != nil && !errors.Is(err, io.EOF) {
				return pos, err
			}
			return pos, nil
		}
		total, err := record.TotalSize(header)
		if err != nil {
			return pos, nil
		}

		buf := make([]byte, total)

		if n, _ := l.f.ReadAt(buf, pos); n < total {
			return pos, nil //record cut off by a crash
		}

		rec, size, err := record.Decode(buf)

		if err != nil {
			return pos, nil // checksum mismatch: corrupt record
		}

		if err := fn(pos, rec); err != nil {
			return pos, err
		}

		pos += int64(size)
	}
}

func (l *Log) Truncate(size int64) error {
	if err := l.f.Truncate(size); err != nil {
		return err
	}
	l.size = size
	return nil
}

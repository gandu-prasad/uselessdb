package datafile

import "os"

// append only file

type Log struct {
	f    *os.File
	size int64 // current end of the file
}

func Open(path string) (*Log, error) {

	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}

	info, err := f.Stat()

	if err != nil {
		f.Close()
		return nil, err
	}
	return &Log{f: f, size: info.Size()}, nil
}

func (l *Log) Size() int64 {
	return l.size
}

func (l *Log) Close() error {
	return l.f.Close()
}

// writes the buf at end of log and returns the position it was written to
// **does not sync (call sync when need duaribility)
func (l *Log) Append(buf []byte) (int64, error) {
	start := l.size

	if _, err := l.f.WriteAt(buf, start); err != nil {
		return 0, err
	}

	l.size += int64(len(buf))
	return start, nil
}

func (l *Log) Sync() error {
	return l.f.Sync()
}

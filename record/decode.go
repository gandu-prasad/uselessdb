package record

import (
	"encoding/binary"
	"errors"
	"hash/crc32"
)

var (
	ErrShort    = errors.New("record: buffer too short")
	ErrChecksum = errors.New("record: checksum mismatch")
)

//decode parses one record from buf and returns it along with it's total size

func Decode(buf []byte) (Record, int, error) {
	if len(buf) < HeaderSize {
		return Record{}, 0, ErrShort
	}

	storedCRC := binary.LittleEndian.Uint32(buf[crcOffset:])
	keyLen := int(binary.LittleEndian.Uint32(buf[keyLenOffset:]))
	valueLen := int(binary.LittleEndian.Uint32(buf[valueLenOffset:]))

	total := HeaderSize + keyLen + valueLen

	if len(buf) < total {
		return Record{}, 0, ErrShort
	}

	//verify the checksum
	if crc32.ChecksumIEEE(buf[crcSize:total]) != storedCRC {
		return Record{}, 0, ErrChecksum
	}

	r := Record{
		TimeStamp: binary.LittleEndian.Uint32(buf[timeStampOffset:]),
		Flag:      buf[flagOffset],
		Key:       buf[HeaderSize : HeaderSize+keyLen],
		Value:     buf[HeaderSize+keyLen : total],
	}

	return r, total, nil
}

// reads the total length fields from a header and returns a full record size
func TotalSize(header []byte) (int, error) {
	if len(header) < HeaderSize {
		return 0, ErrShort
	}

	keyLen := int(binary.LittleEndian.Uint32(header[keyLenOffset:]))
	valueLen := int(binary.LittleEndian.Uint32(header[valueLenOffset:]))
	return HeaderSize + keyLen + valueLen, nil
}

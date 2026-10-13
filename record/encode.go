package record

import (
	"encoding/binary"
	"hash/crc32"
)

// encoding function :- turns a record into byte ready to append to log
func Encode(r Record) []byte {

	total := HeaderSize + len(r.Key) + len(r.Value)
	buf := make([]byte, total)

	//fields after the crc
	binary.LittleEndian.PutUint32(buf[timeStampOffset:], r.TimeStamp)
	binary.LittleEndian.PutUint32(buf[keyLenOffset:], uint32(len(r.Key)))
	binary.LittleEndian.PutUint32(buf[valueLenOffset:], uint32(len(r.Value)))
	buf[flagOffset] = r.Flag

	copy(buf[HeaderSize:], r.Key)
	copy(buf[HeaderSize+len(r.Key):], r.Value)

	//crc checksum

	crc := crc32.ChecksumIEEE(buf[crcSize:])
	binary.LittleEndian.PutUint32(buf[crcOffset:], crc)

	return buf

}

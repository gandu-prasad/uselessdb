package record

const (
	crcSize       = 4
	timeStampSize = 4
	keyLenSize    = 4
	valueLenSize  = 4
	flagSize      = 1

	//offset of each field inside a record
	crcOffset       = 0
	timeStampOffset = crcOffset + crcSize             // 4
	keyLenOffset    = timeStampOffset + timeStampSize // 8
	valueLenOffset  = keyLenOffset + keyLenSize       // 12
	flagOffset      = valueLenOffset + valueLenSize   // 16

	HeaderSize = flagOffset + flagSize // 17 (0-16)
)

const (
	FlagValue     byte = 0 // key-value pair
	FLagTombstone byte = 1 // delete marker; value will be empty
)

type Record struct {
	TimeStamp uint32
	Flag      byte
	Key       []byte
	Value     []byte
}

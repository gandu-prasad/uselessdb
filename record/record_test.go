package record

import (
	"bytes"
	"testing"
)

func TestRoundTrip(t *testing.T) {
	in := Record{TimeStamp: 1000, Flag: FlagValue, Key: []byte("ab"), Value: []byte("xyz")}

	buf := Encode(in)
	out, n, err := Decode(buf)

	if err != nil {
		t.Fatal(err)
	}
	if n != len(buf) {
		t.Errorf("size: got %d, want %d", n, len(buf))
	}

	if out.TimeStamp != in.TimeStamp || out.Flag != in.Flag || !bytes.Equal(out.Key, in.Key) || !bytes.Equal(out.Value, in.Value) {
		t.Errorf("round trip missmatch: %+v", out)
	}

}

func TestCorruptionDetected(t *testing.T) {
	buf := Encode(Record{TimeStamp: 1})
	buf[len(buf)-1] ^= 0xFF //flip last byte

	if _, _, err := Decode(buf); err != ErrChecksum {
		t.Errorf("expected Checksum, got %v", err)
	}
}

func TestTruncated(t *testing.T) {
	buf := Encode(Record{TimeStamp: 1, Key: []byte("k"), Value: []byte("v")})

	if _, _, err := Decode(buf[:len(buf)-1]); err != ErrShort {
		t.Errorf("expected ErrShort, got %v", err)
	}
}

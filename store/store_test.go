package store

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func openTemp(t *testing.T) (*Store, string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "store.db")
	s, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	return s, path
}

func reopen(t *testing.T, s *Store, path string) *Store {
	t.Helper()
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	s2, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	return s2
}

func TestPutAndGet(t *testing.T) {
	s, _ := openTemp(t)
	defer s.Close()

	if err := s.Put([]byte("alice"), []byte("v1")); err != nil {
		t.Fatal(err)
	}
	got, err := s.Get([]byte("alice"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, []byte("v1")) {
		t.Errorf("got %q, want %q", got, "v1")
	}
}

func TestMissingKey(t *testing.T) {
	s, _ := openTemp(t)
	defer s.Close()

	if _, err := s.Get([]byte("nobody")); !errors.Is(err, ErrNotFound) {
		t.Errorf("Get: got %v, want ErrNotFound", err)
	}
	if err := s.Delete([]byte("nobody")); !errors.Is(err, ErrNotFound) {
		t.Errorf("Delete: got %v, want ErrNotFound", err)
	}
}

func TestOverwrite(t *testing.T) {
	s, _ := openTemp(t)
	defer s.Close()

	s.Put([]byte("k"), []byte("old"))
	s.Put([]byte("k"), []byte("new"))

	got, _ := s.Get([]byte("k"))
	if string(got) != "new" {
		t.Errorf("got %q, want %q", got, "new")
	}
}

func TestEmptyKeyRejected(t *testing.T) {
	s, _ := openTemp(t)
	defer s.Close()

	if err := s.Put(nil, []byte("v")); !errors.Is(err, ErrEmptyKey) {
		t.Errorf("got %v, want ErrEmptyKey", err)
	}
}

func TestStatePersistsAcrossReopen(t *testing.T) {
	s, path := openTemp(t)
	s.Put([]byte("alice"), []byte("1"))
	s.Put([]byte("bob"), []byte("2"))
	s.Put([]byte("alice"), []byte("3"))
	s.Delete([]byte("bob"))

	s = reopen(t, s, path)
	defer s.Close()

	got, err := s.Get([]byte("alice"))
	if err != nil || string(got) != "3" {
		t.Errorf("alice: got %q, %v; want \"3\"", got, err)
	}
	if _, err := s.Get([]byte("bob")); !errors.Is(err, ErrNotFound) {
		t.Errorf("bob should stay deleted after reopen, got %v", err)
	}
}

func TestRecoversFromTornWrite(t *testing.T) {
	s, path := openTemp(t)
	s.Put([]byte("a"), []byte("1"))
	s.Close()

	// Simulate a crash partway through writing a record.
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		t.Fatal(err)
	}
	f.Write([]byte{0x12, 0x34, 0x56})
	f.Close()

	s, err = Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	if got, err := s.Get([]byte("a")); err != nil || string(got) != "1" {
		t.Fatalf("after recovery: got %q, %v; want \"1\"", got, err)
	}

	// New writes must land after the good data, not after the garbage.
	if err := s.Put([]byte("b"), []byte("2")); err != nil {
		t.Fatal(err)
	}
	s = reopen(t, s, path)
	defer s.Close()

	if got, err := s.Get([]byte("b")); err != nil || string(got) != "2" {
		t.Errorf("b after second reopen: got %q, %v; want \"2\"", got, err)
	}
}

func TestCompactKeepsLiveData(t *testing.T) {
	s, path := openTemp(t)

	for i := 0; i < 100; i++ {
		s.Put([]byte("k"), []byte(fmt.Sprintf("v%d", i)))
	}
	s.Put([]byte("keep"), []byte("yes"))
	s.Put([]byte("gone"), []byte("x"))
	s.Delete([]byte("gone"))

	before, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Compact(); err != nil {
		t.Fatal(err)
	}
	after, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if after.Size() >= before.Size() {
		t.Errorf("file did not shrink: before %d, after %d", before.Size(), after.Size())
	}

	if v, _ := s.Get([]byte("k")); string(v) != "v99" {
		t.Errorf("k: got %q, want v99", v)
	}
	if v, _ := s.Get([]byte("keep")); string(v) != "yes" {
		t.Errorf("keep: got %q, want yes", v)
	}
	if _, err := s.Get([]byte("gone")); !errors.Is(err, ErrNotFound) {
		t.Errorf("gone: got %v, want ErrNotFound", err)
	}

	// Writes after compaction must work and survive a restart.
	s.Put([]byte("new"), []byte("1"))
	s = reopen(t, s, path)
	defer s.Close()

	if v, _ := s.Get([]byte("k")); string(v) != "v99" {
		t.Errorf("after reopen k: got %q, want v99", v)
	}
	if v, _ := s.Get([]byte("new")); string(v) != "1" {
		t.Errorf("after reopen new: got %q, want 1", v)
	}
}

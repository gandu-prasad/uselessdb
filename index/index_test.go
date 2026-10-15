package index

import "testing"

func TestSetAndGet(t *testing.T) {
	idx := New()
	idx.Set("alice", Entry{Pos: 17, Size: 5})

	e, ok := idx.Get("alice")

	if !ok {
		t.Fatal("alice not found")
	}

	if e.Pos != 17 || e.Size != 5 {
		t.Errorf("got %+v, want {17 5}", e)
	}
}

func TestGetMissing(t *testing.T) {
	idx := New()
	idx.Set("k", Entry{Pos: 1, Size: 1})
	idx.Set("k", Entry{Pos: 99, Size: 3})

	e, _ := idx.Get("k")

	if e.Pos != 99 || e.Size != 3 {
		t.Errorf("got %+v, want {99 3}", e)
	}
	if idx.Len() != 1 {
		t.Errorf("Len: got %d, want 1", idx.Len())
	}
}

func TestDelete(t *testing.T) {
	idx := New()
	idx.Set("k", Entry{Pos: 1, Size: 1})
	idx.Delete("k")

	if _, ok := idx.Get("k"); ok {
		t.Error("key still present after Delete")
	}
	idx.Delete("never-existed") // must not panic
}

func TestRangeVisitsAll(t *testing.T) {
	idx := New()
	idx.Set("a", Entry{Pos: 1})
	idx.Set("b", Entry{Pos: 2})
	idx.Set("c", Entry{Pos: 3})

	seen := 0
	idx.Range(func(key string, e Entry) bool {
		seen++
		return true
	})
	if seen != 3 {
		t.Errorf("visited %d keys, want 3", seen)
	}
}

package store

import (
	"sync"

	"key_value_db/index"
)

type Store struct {
	mu    sync.RWMutex
	index *index.Index
}

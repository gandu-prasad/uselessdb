package index

// last value of the key lives here
type Entry struct {
	Pos  int64
	Size uint32
}

// index is the in memory hash map from key to location
type Index struct {
	m map[string]Entry
}

func New() *Index {
	return &Index{m: make(map[string]Entry)}
}

func (i *Index) Set(key string, e Entry) {
	i.m[key] = e
}

func (i *Index) Get(key string) (Entry, bool) {
	e, ok := i.m[key]
	return e, ok
}

func (i *Index) Delete(key string) {
	delete(i.m, key)
}

func (i *Index) Len() int {
	return len(i.m)
}

func (i *Index) Range(fn func(key string, e Entry) bool) {
	for k, e := range i.m {
		if !fn(k, e) {
			return
		}
	}
}

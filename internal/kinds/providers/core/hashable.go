package core

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"sort"
)

type hashable struct {
	O string
	T string
	R interface{}
	D hashableDeps
}

type hashableDeps []hashable

func (h hashableDeps) Len() int           { return len(h) }
func (h hashableDeps) Less(i, j int) bool { return h[i].md5Sum() < h[j].md5Sum() }
func (h hashableDeps) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h hashable) hash() string {
	h.sortDeps()
	return h.md5Sum()
}

func (h hashable) sortDeps() {
	sort.Sort(h.D)
	for _, h := range h.D {
		h.sortDeps()
	}
}

func (h hashable) md5Sum() string {
	bytes, err := json.Marshal(&h)
	if err != nil {
		panic(err)
	}
	hash := md5.Sum(bytes)
	return hex.EncodeToString(hash[:])
}

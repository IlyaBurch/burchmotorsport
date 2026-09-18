package main

import (
	"sync"
	"time"
)

// ponytail: process-local cache keyed by string, unbounded. Fine for a few
// hundred openf1 keys; add eviction if memory ever shows up on a graph.
type cache struct {
	mu sync.Mutex
	m  map[string]entry
}

type entry struct {
	val any
	err error
	exp time.Time
}

var store = cache{m: map[string]entry{}}

// get returns the cached value or computes it. ttl is chosen by fn so that
// finished sessions can be cached for good while live ones stay fresh.
func (c *cache) get(key string, fn func() (any, time.Duration, error)) (any, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if e, ok := c.m[key]; ok && time.Now().Before(e.exp) {
		return e.val, e.err
	}
	val, ttl, err := fn()
	if err != nil {
		ttl = 5 * time.Second
	}
	c.m[key] = entry{val, err, time.Now().Add(ttl)}
	return val, err
}

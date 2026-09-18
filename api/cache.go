package main

import (
	"log"
	"sync"
	"time"
)

// ponytail: process-local cache keyed by string, unbounded. Fine for a few
// thousand openf1 keys; add eviction if memory ever shows up on a graph.
type cache struct {
	mu sync.Mutex
	m  map[string]*entry
}

type entry struct {
	mu  sync.Mutex // one fetch per key at a time; other keys proceed
	val any
	err error
	exp time.Time
}

var store = cache{m: map[string]*entry{}}

// get returns the cached value or computes it. fn picks the ttl so finished
// sessions can be cached for good while live ones stay fresh. If fn fails and
// an older value exists, the older value is served (stale beats a 502).
func (c *cache) get(key string, fn func() (any, time.Duration, error)) (any, error) {
	c.mu.Lock()
	e := c.m[key]
	if e == nil {
		e = &entry{}
		c.m[key] = e
	}
	c.mu.Unlock()

	e.mu.Lock()
	defer e.mu.Unlock()
	if time.Now().Before(e.exp) {
		return e.val, e.err
	}
	val, ttl, err := fn()
	if err != nil {
		if e.val != nil && e.err == nil {
			log.Printf("cache %s: serving stale, refresh failed: %v", key, err)
			e.exp = time.Now().Add(5 * time.Second)
			return e.val, nil
		}
		e.err, e.exp = err, time.Now().Add(5*time.Second)
		return nil, err
	}
	e.val, e.err, e.exp = val, nil, time.Now().Add(ttl)
	return val, nil
}

package pokecache

import (
	"log"
	"sync"
	"time"
)

type cacheEntry struct {
	createdAt time.Time
	value     []byte
}

type Cache struct {
	entries map[string]cacheEntry
	mu      sync.Mutex
}

func (c *Cache) Add(key string, value []byte) {
	if key == "" {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	entry, ok := c.entries[key]
	if !ok {
		//nolint:exhaustruct
		entry = cacheEntry{
			value: value,
		}
	}

	entry.createdAt = time.Now()
	c.entries[key] = entry
}

func (c *Cache) Get(key string) ([]byte, bool) {
	if key == "" {
		return nil, false
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if entry, ok := c.entries[key]; ok {
		log.Println("Getting from the cache!", key)

		return entry.value, true
	}

	return nil, false
}

func (c *Cache) reapLoop(dataLifespan time.Duration) {
	ticker := time.NewTicker(dataLifespan)

	for now := range ticker.C {
		c.mu.Lock()
		for k, v := range c.entries {
			if now.After(v.createdAt) {
				delete(c.entries, k)
			}
		}
		c.mu.Unlock()
	}
}

func NewCache(dataLifespan time.Duration) *Cache {
	//nolint:exhaustruct
	cache := &Cache{
		entries: make(map[string]cacheEntry),
	}

	go cache.reapLoop(dataLifespan)

	return cache
}

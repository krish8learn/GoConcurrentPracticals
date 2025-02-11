package cache

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type Cache struct {
	// TODO
	keyValues map[string]entry
	size      int
	ttl       time.Duration

	mu     sync.RWMutex
	cancel context.CancelFunc
}

type entry struct {
	value      any
	timeinsert time.Time
}

func New(size int, ttl time.Duration) (*Cache, error) {

	// validation
	if size <= 0 || ttl <= 0 {
		return nil, fmt.Errorf("no input given")
	}

	// initiation
	ctx, cancel := context.WithCancel(context.Background())
	newCache := &Cache{
		keyValues: make(map[string]entry),
		size:      size,
		ttl:       ttl,
		cancel:    cancel,
	}

	// launch the cleaner keep it ready to check the ttl then delete the old records
	go newCache.Cleaner(ctx)

	return newCache, nil
}

func (c *Cache) Close() {
}

func (c *Cache) Get(key string) (any, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	value, present := c.keyValues[key]
	if !present {
		return nil, false
	}
	return value.value, true
}

func (c *Cache) Set(key string, value any) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.keyValues) == c.size {
		c.DeletePop()
	}

	c.keyValues[key] = entry{
		value:      value,
		timeinsert: time.Now(),
	}
}

func (c *Cache) Keys() []string {

	keys := make([]string, 0, len(c.keyValues))
	for key := range c.keyValues {
		keys = append(keys, key)
	}
	return keys
}

func (c *Cache) DeletePop() {

	c.mu.Lock()
	defer c.mu.Unlock()

	oldKey, oldTime := "", time.Now()

	for key, value := range c.keyValues {
		if value.timeinsert.Before(oldTime) {
			oldKey = key
		}
	}

	delete(c.keyValues, oldKey)
}

func (c *Cache) Cleaner(ctx context.Context) {
	timer := time.NewTimer(c.ttl)
	defer timer.Stop()

	for {
		select {
		case <-timer.C:
			c.Clean()
		case <-ctx.Done():
			return
		}
	}
}

func (c *Cache) Clean() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for key, value := range c.keyValues {
		if now.Sub(value.timeinsert) >= c.ttl {
			delete(c.keyValues, key)
		}
	}
}

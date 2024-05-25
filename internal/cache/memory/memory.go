package memory

import (
	"sync"
	"time"

	"github.com/azrod/dnsr/internal/cache/base"
	"github.com/miekg/dns"
)

// MemoryCache is an in-memory cache
type MemoryCache struct {
	mu    sync.RWMutex
	cache map[string]base.CacheValue
}

var _ base.Cache = &MemoryCache{}

// New creates a new in-memory cache
func New() (*MemoryCache, error) {
	return &MemoryCache{
		cache: make(map[string]base.CacheValue),
	}, nil
}

// Load loads the cache
func (c *MemoryCache) Load(data map[string]base.CacheValue) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cache = data

	return nil
}

// Get returns the value for the given key
func (c *MemoryCache) Get(domain string) ([]dns.RR, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.exists(domain) {
		return nil, base.ErrNotFound
	}

	return c.cache[domain].Value, nil
}

// GetAll returns all the values in the cache
func (c *MemoryCache) GetAll() map[string]base.CacheValue {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.cache
}

// Exists returns true if the key exists
func (c *MemoryCache) Exists(domain string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.exists(domain)
}

// exists returns true if the domain exists
func (c *MemoryCache) exists(domain string) bool {
	_, ok := c.cache[domain]
	return ok
}

// Set sets the value for the given key
func (c *MemoryCache) Set(domain string, value []dns.RR) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	ttl := value[0].Header().Ttl
	// if ttl is below 3600, set it to 3600
	if ttl < 3600 {
		ttl = 3600
	}

	c.cache[domain] = base.CacheValue{
		Value:    value,
		ExpireAt: time.Now().Add(time.Duration(ttl) * time.Second),
	}

	return nil
}

// Delete deletes the value for the given key
func (c *MemoryCache) Delete(domain string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.exists(domain) {
		return base.ErrNotFound
	}

	delete(c.cache, domain)

	return nil
}

// Clear clears the cache
func (c *MemoryCache) Clear() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cache = make(map[string]base.CacheValue)

	return nil
}

// Close closes the cache
func (c *MemoryCache) Close() error {
	return nil
}

// Len returns the number of items in the cache
func (c *MemoryCache) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return len(c.cache)
}

// Keys returns the keys in the cache
func (c *MemoryCache) Keys() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	keys := make([]string, 0, len(c.cache))
	for k := range c.cache {
		keys = append(keys, k)
	}

	return keys
}

// HasExpired returns true if the key has expired
func (c *MemoryCache) HasExpired(domain string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.exists(domain) {
		return true
	}

	return time.Now().After(c.cache[domain].ExpireAt)
}

// GetExpireAt returns the expiration time
func (c *MemoryCache) GetExpireAt(domain string) time.Time {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.exists(domain) {
		return time.Time{}
	}

	return c.cache[domain].ExpireAt
}

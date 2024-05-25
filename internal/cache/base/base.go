package base

import (
	"time"

	"github.com/miekg/dns"
)

type Cache interface {
	// Load loads the cache
	Load(data map[string]CacheValue) error

	// Get returns the value for the given key
	Get(domain string) ([]dns.RR, error)

	// GetAll returns all the values in the cache
	GetAll() map[string]CacheValue

	// Set sets the value for the given key
	Set(domain string, value []dns.RR) error

	// Delete deletes the value for the given key
	Delete(key string) error

	// Clear clears the cache
	Clear() error

	// Close closes the cache
	Close() error

	// Len returns the number of items in the cache
	Len() int

	// Keys returns the keys in the cache
	Keys() []string

	// Exists returns true if the key exists
	Exists(domain string) bool

	// HasExpired returns true if the key has expired
	HasExpired(domain string) bool

	// GetExpireAt returns the expiration time
	GetExpireAt(domain string) time.Time
}

type CacheValue struct {
	Value    []dns.RR
	ExpireAt time.Time
}

// GetExpireAt returns the expiration time
func GetExpireAt(TTL int) time.Time {
	return time.Now().Add(time.Duration(TTL) * time.Second)
}

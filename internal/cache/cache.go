// Package cache provides an in-memory TTL cache, replicating the
// JS Cache class from the reference project (cache.js).
package cache

import (
	"sync"
	"time"
)

type item struct {
	value     any
	expiresAt time.Time
}

// Cache is a thread-safe in-memory key-value store with TTL expiration.
type Cache struct {
	mu         sync.RWMutex
	store      map[string]item
	defaultTTL time.Duration
}

// New creates a new Cache with the given default TTL.
func New(defaultTTL time.Duration) *Cache {
	if defaultTTL <= 0 {
		defaultTTL = 60 * time.Second
	}
	c := &Cache{
		store:      make(map[string]item),
		defaultTTL: defaultTTL,
	}
	go c.cleanupLoop()
	return c
}

// Get retrieves a value by key. Returns nil, false if expired or missing.
func (c *Cache) Get(key string) (any, bool) {
	c.mu.RLock()
	it, ok := c.store[key]
	c.mu.RUnlock()
	if !ok {
		return nil, false
	}
	if time.Now().After(it.expiresAt) {
		c.mu.Lock()
		delete(c.store, key)
		c.mu.Unlock()
		return nil, false
	}
	return it.value, true
}

// Set stores a value with the default TTL.
func (c *Cache) Set(key string, value any) {
	c.SetTTL(key, value, c.defaultTTL)
}

// SetTTL stores a value with a specific TTL.
func (c *Cache) SetTTL(key string, value any, ttl time.Duration) {
	c.mu.Lock()
	c.store[key] = item{value: value, expiresAt: time.Now().Add(ttl)}
	c.mu.Unlock()
}

// Delete removes a key from the cache.
func (c *Cache) Delete(key string) {
	c.mu.Lock()
	delete(c.store, key)
	c.mu.Unlock()
}

// Clear removes all entries.
func (c *Cache) Clear() {
	c.mu.Lock()
	c.store = make(map[string]item)
	c.mu.Unlock()
}

// GetOrSet returns the cached value, or calls fn to populate it.
func (c *Cache) GetOrSet(key string, fn func() (any, error), ttl time.Duration) (any, error) {
	if v, ok := c.Get(key); ok {
		return v, nil
	}
	v, err := fn()
	if err != nil {
		return nil, err
	}
	c.SetTTL(key, v, ttl)
	return v, nil
}

// cleanupLoop periodically removes expired entries.
func (c *Cache) cleanupLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		c.mu.Lock()
		now := time.Now()
		for k, v := range c.store {
			if now.After(v.expiresAt) {
				delete(c.store, k)
			}
		}
		c.mu.Unlock()
	}
}

// Stats returns the number of active and expired entries.
func (c *Cache) Stats() (total, active, expired int) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	now := time.Now()
	total = len(c.store)
	for _, v := range c.store {
		if now.After(v.expiresAt) {
			expired++
		} else {
			active++
		}
	}
	return
}

// Cache key constants — mirroring the JS CacheKeys pattern.
const (
	KeyUsers              = "db:users"
	KeyApplications       = "db:applications"
	KeyDemands            = "db:demands"
	KeyEnterprises        = "db:enterprises"
	KeyJobs               = "db:jobs"
	KeyListings           = "db:listings"
	KeyPosts              = "db:posts"
	KeyComments           = "db:comments"
	KeyCertificates       = "db:certificates"
	KeyCourses            = "db:courses"
	KeyInstructors        = "db:instructors"
	KeyPilots             = "db:pilots"
	KeyProducts           = "db:products"
	KeyPolicies           = "db:policies"
	KeyInspections        = "db:inspections"
	KeyLoans              = "db:loans"
	KeyContracts          = "db:contracts"
	KeyLabourOrders       = "db:labour_orders"
	KeyVenues             = "db:venues"
	KeyReviews            = "db:reviews"
	KeyArticles           = "db:articles"
	KeyMessages           = "db:messages"
	KeyEscrowTransactions = "db:escrow_transactions"
	KeyAdminStats         = "stats:admin"
)

// Global cache instance used by the storage layer.
var Global = New(60 * time.Second)

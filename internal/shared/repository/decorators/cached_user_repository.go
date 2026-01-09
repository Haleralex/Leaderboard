package decorators

import (
	"context"
	"fmt"
	"sync"
	"time"

	authmodels "leaderboard-service/internal/auth/models"
	"leaderboard-service/internal/shared/repository"

	"github.com/google/uuid"
)

// CacheEntry represents a cached value with expiration
type CacheEntry struct {
	Value      interface{}
	Expiration time.Time
}

// SimpleCache is a simple in-memory cache
type SimpleCache struct {
	data map[string]CacheEntry
	mu   sync.RWMutex
}

// NewSimpleCache creates a new cache
func NewSimpleCache() *SimpleCache {
	cache := &SimpleCache{
		data: make(map[string]CacheEntry),
	}

	// Start cleanup goroutine
	go cache.cleanupExpired()

	return cache
}

// Get retrieves a value from cache
func (c *SimpleCache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.data[key]
	if !exists {
		return nil, false
	}

	// Check expiration
	if time.Now().After(entry.Expiration) {
		return nil, false
	}

	return entry.Value, true
}

// Set stores a value in cache with TTL
func (c *SimpleCache) Set(key string, value interface{}, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data[key] = CacheEntry{
		Value:      value,
		Expiration: time.Now().Add(ttl),
	}
}

// Delete removes a value from cache
func (c *SimpleCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.data, key)
}

// Clear removes all entries from cache
func (c *SimpleCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data = make(map[string]CacheEntry)
}

// cleanupExpired removes expired entries periodically
func (c *SimpleCache) cleanupExpired() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.mu.Lock()
		now := time.Now()
		for key, entry := range c.data {
			if now.After(entry.Expiration) {
				delete(c.data, key)
			}
		}
		c.mu.Unlock()
	}
}

// CachedUserRepository decorates UserRepository with caching
type CachedUserRepository struct {
	inner repository.UserRepository
	cache *SimpleCache
	ttl   time.Duration
}

// NewCachedUserRepository creates a cached user repository
func NewCachedUserRepository(inner repository.UserRepository, cache *SimpleCache) repository.UserRepository {
	return &CachedUserRepository{
		inner: inner,
		cache: cache,
		ttl:   5 * time.Minute, // Default TTL
	}
}

// Create creates a user and invalidates cache
func (r *CachedUserRepository) Create(ctx context.Context, user *authmodels.User) error {
	err := r.inner.Create(ctx, user)
	if err != nil {
		return err
	}

	// Cache the created user
	r.cacheUser(user)

	return nil
}

// FindByID retrieves a user by ID with caching
func (r *CachedUserRepository) FindByID(ctx context.Context, id uuid.UUID) (*authmodels.User, error) {
	key := r.userIDKey(id)

	// Check cache first
	if cached, ok := r.cache.Get(key); ok {
		return cached.(*authmodels.User), nil
	}

	// Cache miss - fetch from inner repository
	user, err := r.inner.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Store in cache
	r.cacheUser(user)

	return user, nil
}

// FindByEmail retrieves a user by email with caching
func (r *CachedUserRepository) FindByEmail(ctx context.Context, email string) (*authmodels.User, error) {
	key := r.userEmailKey(email)

	// Check cache first
	if cached, ok := r.cache.Get(key); ok {
		return cached.(*authmodels.User), nil
	}

	// Cache miss - fetch from inner repository
	user, err := r.inner.FindByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	// Store in cache (both by ID and email)
	r.cacheUser(user)

	return user, nil
}

// Update updates a user and invalidates cache
func (r *CachedUserRepository) Update(ctx context.Context, user *authmodels.User) error {
	err := r.inner.Update(ctx, user)
	if err != nil {
		return err
	}

	// Invalidate old cache entries
	r.cache.Delete(r.userIDKey(user.ID))

	// Cache the updated user
	r.cacheUser(user)

	return nil
}

// Delete deletes a user and invalidates cache
func (r *CachedUserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	// Fetch user first to get email for cache invalidation
	user, err := r.inner.FindByID(ctx, id)
	if err == nil {
		r.cache.Delete(r.userEmailKey(user.Email))
	}

	// Delete from repository
	err = r.inner.Delete(ctx, id)
	if err != nil {
		return err
	}

	// Invalidate cache
	r.cache.Delete(r.userIDKey(id))

	return nil
}

// Helper methods

func (r *CachedUserRepository) cacheUser(user *authmodels.User) {
	// Cache by ID
	r.cache.Set(r.userIDKey(user.ID), user, r.ttl)

	// Cache by email
	r.cache.Set(r.userEmailKey(user.Email), user, r.ttl)
}

func (r *CachedUserRepository) userIDKey(id uuid.UUID) string {
	return fmt.Sprintf("user:id:%s", id.String())
}

func (r *CachedUserRepository) userEmailKey(email string) string {
	return fmt.Sprintf("user:email:%s", email)
}

// FindBySpec finds users matching a specification (no caching for complex queries)
func (r *CachedUserRepository) FindBySpec(ctx context.Context, spec repository.Specification[authmodels.User]) ([]*authmodels.User, error) {
	// Specifications are too complex to cache efficiently, delegate to inner repository
	return r.inner.FindBySpec(ctx, spec)
}

// FindOneBySpec finds first user matching a specification (no caching for complex queries)
func (r *CachedUserRepository) FindOneBySpec(ctx context.Context, spec repository.Specification[authmodels.User]) (*authmodels.User, error) {
	// Specifications are too complex to cache efficiently, delegate to inner repository
	return r.inner.FindOneBySpec(ctx, spec)
}

// CountBySpec counts users matching a specification (no caching for counts)
func (r *CachedUserRepository) CountBySpec(ctx context.Context, spec repository.Specification[authmodels.User]) (int64, error) {
	// Counts are typically fast and don't benefit much from caching
	return r.inner.CountBySpec(ctx, spec)
}

package sketch

import (
	"time"

	"github.com/jellydator/ttlcache/v3"
)



type SketchStore struct {
	cache *ttlcache.Cache[string, *Sketch] // Assuming you want to cache base64 strings, adjust type if needed
}

func NewSketchStore() *SketchStore {
	cache := ttlcache.New[string, *Sketch]() // Initialize ttlcache
	go cache.Start()
	return &SketchStore{cache: cache}
}

// NewSketchStoreWithTTL creates a new SketchStore with TTL configuration.
func NewSketchStoreWithTTL(ttl time.Duration, cleanupInterval time.Duration) *SketchStore {
	cache := ttlcache.New[string, *Sketch](
		ttlcache.WithTTL[string, *Sketch](ttl),
		ttlcache.WithDisableTouchOnHit[string, *Sketch](), // Optional: extend TTL on Get, default is false
		ttlcache.WithCapacity[string, *Sketch](1000),      // Optional: set capacity, adjust as needed
	)
	go cache.Start() // Start background cleanup
	return &SketchStore{cache: cache}
}

func (s *SketchStore) GetSketch(key string) (*Sketch, bool, error) {
	item := s.cache.Get(key)
	if item == nil {
		return nil, false, nil // Not found in cache, no error
	}
	return item.Value(), true, nil
}

func (s *SketchStore) SetSketch(key string, sketch *Sketch, ttl time.Duration) error {
	if ttl == 0 {
		s.cache.Set(key, sketch, ttlcache.NoTTL) // No expiration
	} else {
		s.cache.Set(key, sketch, ttl)
	}
	return nil
}

func (s *SketchStore) DeleteSketch(key string) error {
	s.cache.Delete(key)
	return nil
}

func (s *SketchStore) StopCleanup() {
	s.cache.Stop()
}

func (s *SketchStore) DeleteAll() {
	s.cache.DeleteAll()
}

func (s *SketchStore) Count() int {
	return s.cache.Len()
}

package sketch

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// createDummySketch returns a dummy Sketch for testing.
func createDummySketch(id string) *Sketch {
	return &Sketch{
		ID:       id,
		ImageURL: "dummy_url",
		OwnerID:  "dummy_owner",
	}
}

func TestNewSketchStore_Count(t *testing.T) {
	store := NewSketchStore()
	defer store.StopCleanup()

	assert.Equal(t, 0, store.Count(), "Initial count should be zero")
}

func TestSetAndGetSketch(t *testing.T) {

	store := NewSketchStore()
	defer store.StopCleanup()

	dummy := createDummySketch("test1")
	err := store.SetSketch("test1", dummy, 24*time.Hour)
	assert.NoError(t, err, "SetSketch should not error")

	retrieved, found, err := store.GetSketch("test1")
	assert.NoError(t, err, "GetSketch should not error")
	assert.True(t, found, "Sketch should be found")
	assert.Equal(t, dummy, retrieved, "Retrieved sketch should match inserted sketch")
}

func TestDeleteSketch(t *testing.T) {
	store := NewSketchStore()
	defer store.StopCleanup()

	dummy := createDummySketch("test2")
	_ = store.SetSketch("test2", dummy, 24*time.Hour)

	err := store.DeleteSketch("test2")
	assert.NoError(t, err, "DeleteSketch should not error")

	_, found, err := store.GetSketch("test2")
	assert.NoError(t, err, "GetSketch should not error")
	assert.False(t, found, "Sketch should not be found after deletion")
}

func TestDeleteAllAndCount(t *testing.T) {
	store := NewSketchStore()
	defer store.StopCleanup()

	dummy1 := createDummySketch("a")
	dummy2 := createDummySketch("b")
	_ = store.SetSketch("a", dummy1, 24*time.Hour)
	_ = store.SetSketch("b", dummy2, 24*time.Hour)

	assert.Equal(t, 2, store.Count(), "Count should be two after setting two sketches")

	store.DeleteAll()
	assert.Equal(t, 0, store.Count(), "Count should be zero after DeleteAll")
}

func TestTTLExpiration(t *testing.T) {
	ttl := 100 * time.Millisecond
	// Set a shorter cleanup interval.
	store := NewSketchStoreWithTTL(ttl, 50*time.Millisecond)
	defer store.StopCleanup()

	dummy := createDummySketch("ttlTest")
	_ = store.SetSketch("ttlTest", dummy, ttl)

	// Immediately, the item should be present.
	_, found, _ := store.GetSketch("ttlTest")
	t.Log("After insert: found =", found)
	assert.True(t, found, "Sketch should be found initially")

	// Wait for TTL to expire.
	time.Sleep(200 * time.Millisecond)
	_, found, _ = store.GetSketch("ttlTest")
	t.Log("After sleep: found =", found)
	assert.False(t, found, "Sketch should expire after TTL")
}

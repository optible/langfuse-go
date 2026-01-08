package cache

import (
	"testing"
	"time"
)

func TestCache_SetAndGet(t *testing.T) {
	c := New[string](time.Minute)

	// Set a value
	c.Set("key1", "value1")

	// Get the value
	val, found, expired := c.Get("key1")
	if !found {
		t.Error("expected to find key1")
	}
	if expired {
		t.Error("expected key1 to not be expired")
	}
	if val != "value1" {
		t.Errorf("expected value1, got %s", val)
	}

	// Get non-existent key
	_, found, _ = c.Get("nonexistent")
	if found {
		t.Error("expected to not find nonexistent key")
	}
}

func TestCache_Expiration(t *testing.T) {
	c := New[string](50 * time.Millisecond)

	c.Set("key1", "value1")

	// Should not be expired initially
	_, found, expired := c.Get("key1")
	if !found || expired {
		t.Error("expected key1 to exist and not be expired")
	}

	// Wait for expiration
	time.Sleep(60 * time.Millisecond)

	// Should be expired now
	val, found, expired := c.Get("key1")
	if !found {
		t.Error("expected to still find expired key")
	}
	if !expired {
		t.Error("expected key1 to be expired")
	}
	if val != "value1" {
		t.Error("expected to get stale value")
	}
}

func TestCache_SetWithTTL(t *testing.T) {
	c := New[string](time.Hour) // Long default TTL

	// Set with short TTL
	c.SetWithTTL("key1", "value1", 50*time.Millisecond)

	// Should not be expired initially
	_, _, expired := c.Get("key1")
	if expired {
		t.Error("expected key1 to not be expired")
	}

	// Wait for expiration
	time.Sleep(60 * time.Millisecond)

	// Should be expired now despite long default TTL
	_, _, expired = c.Get("key1")
	if !expired {
		t.Error("expected key1 to be expired")
	}
}

func TestCache_Delete(t *testing.T) {
	c := New[string](time.Minute)

	c.Set("key1", "value1")
	c.Delete("key1")

	_, found, _ := c.Get("key1")
	if found {
		t.Error("expected key1 to be deleted")
	}
}

func TestCache_Clear(t *testing.T) {
	c := New[string](time.Minute)

	c.Set("key1", "value1")
	c.Set("key2", "value2")

	if c.Size() != 2 {
		t.Errorf("expected size 2, got %d", c.Size())
	}

	c.Clear()

	if c.Size() != 0 {
		t.Errorf("expected size 0, got %d", c.Size())
	}
}

func TestCache_Cleanup(t *testing.T) {
	c := New[string](50 * time.Millisecond)

	c.Set("key1", "value1")
	c.Set("key2", "value2")

	// Wait for expiration
	time.Sleep(60 * time.Millisecond)

	// Add a fresh entry
	c.Set("key3", "value3")

	// Run cleanup
	c.Cleanup()

	// key1 and key2 should be removed, key3 should remain
	if c.Size() != 1 {
		t.Errorf("expected size 1 after cleanup, got %d", c.Size())
	}

	_, found, _ := c.Get("key3")
	if !found {
		t.Error("expected key3 to still exist")
	}
}

func TestCache_SetTTL(t *testing.T) {
	c := New[string](time.Minute)

	if c.GetTTL() != time.Minute {
		t.Error("expected initial TTL to be 1 minute")
	}

	c.SetTTL(time.Hour)

	if c.GetTTL() != time.Hour {
		t.Error("expected TTL to be updated to 1 hour")
	}
}

func TestCache_ConcurrentAccess(t *testing.T) {
	c := New[int](time.Minute)

	done := make(chan bool)

	// Start multiple goroutines
	for i := 0; i < 10; i++ {
		go func(id int) {
			for j := 0; j < 100; j++ {
				c.Set("key"+string(rune(id)), id*100+j)
				c.Get("key" + string(rune(id)))
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
}

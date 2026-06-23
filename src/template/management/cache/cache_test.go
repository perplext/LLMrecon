package cache

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/perplext/LLMrecon/src/template/format"
)

func tmpl(id string) *format.Template {
	return &format.Template{ID: id, Name: "tmpl-" + id}
}

func TestNewTemplateCache_Defaults(t *testing.T) {
	c := NewTemplateCache(0, 0, "")
	stats := c.GetStats()
	if stats["max_size"].(int) != 100 {
		t.Errorf("default max_size = %v, want 100", stats["max_size"])
	}
	if stats["eviction_policy"].(string) != string(LRU) {
		t.Errorf("default policy = %v, want lru", stats["eviction_policy"])
	}
	if stats["default_ttl_ms"].(int64) != time.Hour.Milliseconds() {
		t.Errorf("default ttl = %v, want 1h", stats["default_ttl_ms"])
	}
}

func TestCache_SetGetReadAfterWrite(t *testing.T) {
	c := NewTemplateCache(time.Hour, 10, LRU)
	c.Set("a", tmpl("a"))

	got, ok := c.Get("a")
	if !ok || got.ID != "a" {
		t.Fatalf("read-after-write failed: ok=%v got=%v", ok, got)
	}
	if _, ok := c.Get("missing"); ok {
		t.Fatal("Get on missing id should be false")
	}
}

func TestCache_Expiry(t *testing.T) {
	c := NewTemplateCache(time.Hour, 10, LRU)
	c.SetWithTTL("a", tmpl("a"), 10*time.Millisecond)

	time.Sleep(25 * time.Millisecond)
	if _, ok := c.Get("a"); ok {
		t.Fatal("expired entry should not be returned")
	}
	// Get should have removed the expired entry.
	if c.Size() != 0 {
		t.Fatalf("expired entry should be evicted on Get, size=%d", c.Size())
	}
}

func TestCache_DeleteClearSize(t *testing.T) {
	c := NewTemplateCache(time.Hour, 10, LRU)
	c.Set("a", tmpl("a"))
	c.Set("b", tmpl("b"))
	if c.Size() != 2 {
		t.Fatalf("size = %d, want 2", c.Size())
	}
	c.Delete("a")
	if _, ok := c.Get("a"); ok {
		t.Fatal("deleted entry still present")
	}
	c.Clear()
	if c.Size() != 0 {
		t.Fatalf("size after clear = %d, want 0", c.Size())
	}
}

func TestCache_Prune(t *testing.T) {
	c := NewTemplateCache(time.Hour, 10, LRU)
	c.Set("fresh", tmpl("fresh"))
	c.SetWithTTL("expired", tmpl("expired"), 1*time.Millisecond)
	time.Sleep(5 * time.Millisecond)

	// Prune with a large maxAge removes only the already-expired entry.
	removed := c.Prune(time.Hour)
	if removed != 1 {
		t.Fatalf("Prune removed %d, want 1", removed)
	}
	if _, ok := c.Get("fresh"); !ok {
		t.Fatal("fresh entry should survive prune")
	}
}

func TestCache_EvictLRU(t *testing.T) {
	c := NewTemplateCache(time.Hour, 2, LRU)
	c.Set("a", tmpl("a"))
	time.Sleep(2 * time.Millisecond)
	c.Set("b", tmpl("b"))
	// Touch "a" so "b" becomes least-recently-used.
	time.Sleep(2 * time.Millisecond)
	c.Get("a")
	c.Set("c", tmpl("c")) // full -> evict LRU (b)

	if _, ok := c.Get("b"); ok {
		t.Fatal("LRU should have evicted b")
	}
	if _, ok := c.Get("a"); !ok {
		t.Fatal("a should remain after LRU eviction")
	}
}

func TestCache_EvictLFU(t *testing.T) {
	c := NewTemplateCache(time.Hour, 2, LFU)
	c.Set("a", tmpl("a"))
	c.Set("b", tmpl("b"))
	// Access "a" so "b" is least-frequently-used.
	c.Get("a")
	c.Get("a")
	c.Set("c", tmpl("c")) // evict LFU (b)

	if _, ok := c.Get("b"); ok {
		t.Fatal("LFU should have evicted b")
	}
}

func TestCache_EvictFIFO(t *testing.T) {
	c := NewTemplateCache(time.Hour, 2, FIFO)
	c.Set("a", tmpl("a"))
	time.Sleep(2 * time.Millisecond)
	c.Set("b", tmpl("b"))
	// Even though "a" is accessed, FIFO evicts by creation order.
	c.Get("a")
	c.Set("c", tmpl("c")) // evict oldest-created (a)

	if _, ok := c.Get("a"); ok {
		t.Fatal("FIFO should have evicted a (oldest created)")
	}
}

func TestCache_SetMaxSizeShrinks(t *testing.T) {
	c := NewTemplateCache(time.Hour, 10, LRU)
	for i := 0; i < 5; i++ {
		c.Set(fmt.Sprintf("k%d", i), tmpl(fmt.Sprintf("k%d", i)))
	}
	c.SetMaxSize(2)
	if c.Size() > 2 {
		t.Fatalf("SetMaxSize should evict down to 2, size=%d", c.Size())
	}
	// Non-positive sizes are ignored.
	c.SetMaxSize(0)
	if c.GetStats()["max_size"].(int) != 2 {
		t.Fatal("SetMaxSize(0) should be ignored")
	}
}

func TestCache_RefreshAndKeys(t *testing.T) {
	c := NewTemplateCache(time.Hour, 10, LRU)
	c.SetWithTTL("a", tmpl("a"), 10*time.Millisecond)

	if !c.Refresh("a") {
		t.Fatal("Refresh on existing entry should return true")
	}
	if c.Refresh("missing") {
		t.Fatal("Refresh on missing entry should return false")
	}
	if !c.RefreshWithTTL("a", time.Hour) {
		t.Fatal("RefreshWithTTL on existing entry should return true")
	}
	// After refresh with long TTL, the entry should outlive the original 10ms.
	time.Sleep(20 * time.Millisecond)
	if _, ok := c.Get("a"); !ok {
		t.Fatal("refreshed entry should not have expired")
	}

	if len(c.GetKeys()) != 1 {
		t.Fatalf("GetKeys = %v", c.GetKeys())
	}
}

// TestCache_ConcurrentAccess exercises the RWMutex paths under -race.
func TestCache_ConcurrentAccess(t *testing.T) {
	c := NewTemplateCache(time.Hour, 1000, LRU)
	var wg sync.WaitGroup
	for g := 0; g < 8; g++ {
		wg.Add(1)
		go func(g int) {
			defer wg.Done()
			for i := 0; i < 200; i++ {
				id := fmt.Sprintf("g%d-%d", g, i)
				c.Set(id, tmpl(id))
				c.Get(id)
				if i%10 == 0 {
					c.Size()
					c.GetStats()
					c.GetKeys()
				}
			}
		}(g)
	}
	wg.Wait()
}

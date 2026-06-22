package keystore

import (
	"testing"
	"time"
)

// newPopulatedKeystore builds an in-memory Keystore with a few entries for
// retrieval tests. Same-package access lets us seed the unexported map directly.
func newPopulatedKeystore() (*Keystore, *KeyRetriever) {
	ks := &Keystore{keys: make(map[string]*KeyEntry)}
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	ks.keys["rsa-1"] = &KeyEntry{
		ID: "rsa-1", Type: KeyTypeRSA, Algorithm: "RSA-2048",
		Metadata: map[string]string{"env": "prod", "team": "blue"},
		CreatedAt: base, UpdatedAt: base,
	}
	ks.keys["aes-1"] = &KeyEntry{
		ID: "aes-1", Type: KeyTypeAES, Algorithm: "AES-256",
		Metadata: map[string]string{"env": "test"},
		CreatedAt: base.AddDate(0, 0, 10), UpdatedAt: base.AddDate(0, 0, 10),
	}
	ks.keys["aes-2"] = &KeyEntry{
		ID: "aes-2", Type: KeyTypeAES, Algorithm: "AES-128",
		Metadata: map[string]string{"env": "prod"},
		CreatedAt: base.AddDate(0, 0, 20), UpdatedAt: base.AddDate(0, 0, 20),
	}
	return ks, NewKeyRetriever(ks)
}

func TestRetriever_GetKey(t *testing.T) {
	_, kr := newPopulatedKeystore()

	if _, err := kr.GetKey(""); err == nil {
		t.Error("GetKey with empty id must error")
	}
	if _, err := kr.GetKey("missing"); err == nil {
		t.Error("GetKey on missing id must error")
	}

	got, err := kr.GetKey("rsa-1")
	if err != nil {
		t.Fatalf("GetKey: %v", err)
	}
	// Returned entry must be a defensive copy of the metadata map.
	got.Metadata["env"] = "tampered"
	again, _ := kr.GetKey("rsa-1")
	if again.Metadata["env"] != "prod" {
		t.Fatal("GetKey must return a deep copy of metadata")
	}
}

func TestRetriever_GetKeys(t *testing.T) {
	_, kr := newPopulatedKeystore()

	got, err := kr.GetKeys(nil)
	if err != nil {
		t.Fatalf("GetKeys(nil): %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("GetKeys(nil) should be empty, got %d", len(got))
	}

	got, err = kr.GetKeys([]string{"rsa-1", "aes-1", "ghost"})
	if err != nil {
		t.Fatalf("GetKeys: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 found keys, got %d", len(got))
	}
	if _, ok := got["ghost"]; ok {
		t.Fatal("non-existent key should be omitted")
	}
}

func TestRetriever_FindKeysByType(t *testing.T) {
	_, kr := newPopulatedKeystore()
	aes, err := kr.FindKeysByType(KeyTypeAES)
	if err != nil {
		t.Fatalf("FindKeysByType: %v", err)
	}
	if len(aes) != 2 {
		t.Fatalf("expected 2 AES keys, got %d", len(aes))
	}
}

func TestRetriever_FindKeysByAlgorithm(t *testing.T) {
	_, kr := newPopulatedKeystore()
	if _, err := kr.FindKeysByAlgorithm(""); err == nil {
		t.Error("empty algorithm must error")
	}
	res, err := kr.FindKeysByAlgorithm("AES-256")
	if err != nil {
		t.Fatalf("FindKeysByAlgorithm: %v", err)
	}
	if len(res) != 1 || res[0].ID != "aes-1" {
		t.Fatalf("unexpected result: %+v", res)
	}
}

func TestRetriever_FindKeysByMetadata(t *testing.T) {
	_, kr := newPopulatedKeystore()
	if _, err := kr.FindKeysByMetadata(nil); err == nil {
		t.Error("empty metadata must error")
	}
	prod, err := kr.FindKeysByMetadata(map[string]string{"env": "prod"})
	if err != nil {
		t.Fatalf("FindKeysByMetadata: %v", err)
	}
	if len(prod) != 2 {
		t.Fatalf("expected 2 prod keys, got %d", len(prod))
	}
}

func TestRetriever_FindKeysCreatedAfter(t *testing.T) {
	_, kr := newPopulatedKeystore()
	cutoff := time.Date(2026, 1, 5, 0, 0, 0, 0, time.UTC)
	res, err := kr.FindKeysCreatedAfter(cutoff)
	if err != nil {
		t.Fatalf("FindKeysCreatedAfter: %v", err)
	}
	if len(res) != 2 {
		t.Fatalf("expected 2 keys created after cutoff, got %d", len(res))
	}
}

func TestRetriever_SearchKeys(t *testing.T) {
	_, kr := newPopulatedKeystore()
	if _, err := kr.SearchKeys(""); err == nil {
		t.Error("empty query must error")
	}
	// Match on algorithm substring.
	res, err := kr.SearchKeys("rsa")
	if err != nil {
		t.Fatalf("SearchKeys: %v", err)
	}
	if len(res) != 1 || res[0].ID != "rsa-1" {
		t.Fatalf("search 'rsa' returned %+v", res)
	}
	// Match on metadata value.
	blue, _ := kr.SearchKeys("blue")
	if len(blue) != 1 || blue[0].ID != "rsa-1" {
		t.Fatalf("search 'blue' returned %+v", blue)
	}
}

func TestRetriever_ListKeysSortAndPaginate(t *testing.T) {
	_, kr := newPopulatedKeystore()

	// Sort by id descending.
	res, err := kr.ListKeys(&KeySearchOptions{
		Sort: &SortOptions{Field: "id", Direction: "desc"},
	})
	if err != nil {
		t.Fatalf("ListKeys: %v", err)
	}
	if res.Total != 3 {
		t.Fatalf("expected total 3, got %d", res.Total)
	}
	if res.Keys[0].ID != "rsa-1" {
		t.Fatalf("desc sort should put rsa-1 first, got %s", res.Keys[0].ID)
	}

	// Paginate: first page of 2.
	page, err := kr.ListKeys(&KeySearchOptions{
		Sort:       &SortOptions{Field: "id", Direction: "asc"},
		Pagination: &PaginationOptions{Limit: 2, Offset: 0},
	})
	if err != nil {
		t.Fatalf("ListKeys page: %v", err)
	}
	if len(page.Keys) != 2 || !page.HasMore {
		t.Fatalf("first page should have 2 keys and HasMore=true, got %d/%v", len(page.Keys), page.HasMore)
	}

	// Offset beyond the end returns an empty page, not an error.
	beyond, err := kr.ListKeys(&KeySearchOptions{
		Pagination: &PaginationOptions{Limit: 2, Offset: 99},
	})
	if err != nil {
		t.Fatalf("ListKeys beyond: %v", err)
	}
	if len(beyond.Keys) != 0 || beyond.HasMore {
		t.Fatalf("offset beyond end should be empty, got %d/%v", len(beyond.Keys), beyond.HasMore)
	}
}

func TestRetriever_CountExistsAndIDs(t *testing.T) {
	_, kr := newPopulatedKeystore()

	if n := kr.GetKeyCount(); n != 3 {
		t.Fatalf("GetKeyCount = %d, want 3", n)
	}
	if !kr.KeyExists("aes-1") {
		t.Error("KeyExists should be true for aes-1")
	}
	if kr.KeyExists("") || kr.KeyExists("ghost") {
		t.Error("KeyExists should be false for empty/missing ids")
	}

	ids := kr.GetKeyIDs()
	if len(ids) != 3 {
		t.Fatalf("GetKeyIDs = %v", ids)
	}
	// GetKeyIDs sorts ascending.
	if ids[0] != "aes-1" || ids[2] != "rsa-1" {
		t.Fatalf("GetKeyIDs not sorted: %v", ids)
	}
}

func TestRetriever_Statistics(t *testing.T) {
	_, kr := newPopulatedKeystore()
	stats := kr.GetKeyStatistics()
	if stats["total_keys"].(int) != 3 {
		t.Fatalf("total_keys = %v", stats["total_keys"])
	}
	types := stats["types"].(map[string]int)
	if types[string(KeyTypeAES)] != 2 {
		t.Fatalf("expected 2 AES in stats, got %d", types[string(KeyTypeAES)])
	}
}

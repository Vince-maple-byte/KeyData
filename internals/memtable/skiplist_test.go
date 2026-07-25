package memtable

import (
	"testing"
	"time"

	"github.com/Vince-maple-byte/KeyData/internals/record"
)

// --- helpers ---

func newList(t *testing.T) *Skiplist {
	t.Helper()
	return CreateSkiplist()
}

// --- CreateSkiplist ---

func TestCreateSkiplist_NotNil(t *testing.T) {
	list := newList(t)
	if list == nil {
		t.Fatal("expected non-nil Skiplist")
	}
}

func TestCreateSkiplist_InitialSizeZero(t *testing.T) {
	list := newList(t)
	if list.size != 0 {
		t.Errorf("expected size 0, got %d", list.size)
	}
}

func TestCreateSkiplist_HeadHasMaxLevels(t *testing.T) {
	list := newList(t)
	if len(list.head.levels) != MAX_LEVEL {
		t.Errorf("expected head to have %d levels, got %d", MAX_LEVEL, len(list.head.levels))
	}
}

// --- Insert ---

func TestInsert_SingleElement(t *testing.T) {
	list := newList(t)
	list.Insert("a", []byte("val-a"))
	if list.size != 1 {
		t.Errorf("expected size 1, got %d", list.size)
	}
}

func TestInsert_MultipleElements(t *testing.T) {
	list := newList(t)
	keys := []string{"banana", "apple", "cherry"}
	for _, k := range keys {
		list.Insert(k, []byte(k+"-value"))
	}
	if list.size != len(keys) {
		t.Errorf("expected size %d, got %d", len(keys), list.size)
	}
}

func TestInsert_UpdateExistingKeyForDifferentTimesItWasInserted(t *testing.T) {
	list := newList(t)
	old, _ := record.CreateRecord("key", "old", "PUT")
	time.Sleep(2 * time.Millisecond)
	new, _ := record.CreateRecord("key", "new", "PUT")
	list.Insert("key", old)
	list.Insert("key", new)

	// Size should not grow on update.
	if list.size != 1 {
		t.Errorf("expected size 1 after update, got %d", list.size)
	}

	val, err := list.Search("key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if record.GetContents(val).Payload != record.GetContents(new).Payload {
		t.Errorf("expected 'new', got %q\nTime for old:%v\nTime for new:%v",
			val, record.GetContents(old).Timestamp, record.GetContents(new).Timestamp)

	}
}

func TestInsert_UpdateExistingKeyWhenInsertedAtTheSameTime(t *testing.T) {
	list := newList(t)
	old, _ := record.CreateRecord("key", "old", "PUT")
	new, _ := record.CreateRecord("key", "new", "PUT")
	list.Insert("key", old)
	list.Insert("key", new)

	// Size should not grow on update.
	if list.size != 1 {
		t.Errorf("expected size 1 after update, got %d", list.size)
	}

	val, err := list.Search("key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if record.GetContents(val).Payload != record.GetContents(new).Payload {
		t.Errorf("expected 'new', got %q\nTime for old:%v\nTime for new:%v",
			val, record.GetContents(old).Timestamp, record.GetContents(new).Timestamp)

	}
}

func TestInsert_NilValue(t *testing.T) {
	list := newList(t)
	list.Insert("nilkey", nil)
	val, err := list.Search("nilkey")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != nil {
		t.Errorf("expected nil value, got %v", val)
	}
}

func TestInsert_EmptyStringKey(t *testing.T) {
	list := newList(t)
	list.Insert("", []byte("empty-key-value"))
	val, err := list.Search("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(val) != "empty-key-value" {
		t.Errorf("expected 'empty-key-value', got %q", val)
	}
}

func TestInsert_PreservesOrder(t *testing.T) {
	list := newList(t)
	// Insert out of alphabetical order.
	list.Insert("c", []byte("c"))
	list.Insert("a", []byte("a"))
	list.Insert("b", []byte("b"))

	// Level-0 linked list must be sorted.
	curr := list.head.levels[0]
	var prev string
	for curr != nil {
		if curr.key < prev {
			t.Errorf("out of order: %q after %q", curr.key, prev)
		}
		prev = curr.key
		curr = curr.levels[0]
	}
}

// --- Search ---

func TestSearch_ExistingKey(t *testing.T) {
	list := newList(t)
	list.Insert("hello", []byte("world"))
	val, err := list.Search("hello")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(val) != "world" {
		t.Errorf("expected 'world', got %q", val)
	}
}

func TestSearch_MissingKey(t *testing.T) {
	list := newList(t)
	_, err := list.Search("ghost")
	if err == nil {
		t.Error("expected error for missing key, got nil")
	}
}

func TestSearch_EmptyList(t *testing.T) {
	list := newList(t)
	_, err := list.Search("anything")
	if err == nil {
		t.Error("expected error on empty list, got nil")
	}
}

func TestSearch_AfterMultipleInserts(t *testing.T) {
	list := newList(t)
	pairs := map[string]string{
		"foo": "1",
		"bar": "2",
		"baz": "3",
	}
	for k, v := range pairs {
		list.Insert(k, []byte(v))
	}
	for k, want := range pairs {
		got, err := list.Search(k)
		if err != nil {
			t.Errorf("Search(%q): unexpected error: %v", k, err)
			continue
		}
		if string(got) != want {
			t.Errorf("Search(%q): expected %q, got %q", k, want, got)
		}
	}
}

// --- Delete ---

func TestDelete_ExistingKey(t *testing.T) {
	list := newList(t)
	list.Insert("del-me", []byte("value"))
	val, err := list.Delete("del-me")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(val) != "value" {
		t.Errorf("expected returned value 'value', got %q", val)
	}
}

func TestDelete_DecreasesSize(t *testing.T) {
	list := newList(t)
	list.Insert("x", []byte("x"))
	list.Insert("y", []byte("y"))
	list.Delete("x")
	if list.size != 1 {
		t.Errorf("expected size 1 after delete, got %d", list.size)
	}
}

func TestDelete_MakesKeyUnsearchable(t *testing.T) {
	list := newList(t)
	list.Insert("gone", []byte("poof"))
	list.Delete("gone")
	_, err := list.Search("gone")
	if err == nil {
		t.Error("expected error searching deleted key, got nil")
	}
}

func TestDelete_MissingKey(t *testing.T) {
	list := newList(t)
	_, err := list.Delete("nope")
	if err == nil {
		t.Error("expected error deleting missing key, got nil")
	}
}

func TestDelete_OnEmptyList(t *testing.T) {
	list := newList(t)
	_, err := list.Delete("anything")
	if err == nil {
		t.Error("expected error deleting from empty list, got nil")
	}
}

func TestDelete_CorrectValueReturned(t *testing.T) {
	list := newList(t)
	list.Insert("k", []byte("expected-val"))
	val, err := list.Delete("k")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(val) != "expected-val" {
		t.Errorf("expected 'expected-val', got %q", val)
	}
}

// --- EntireList ---

func TestEntireList_EmptyList(t *testing.T) {
	list := newList(t)
	res := list.EntireList()
	// Result should not be nil; contents may vary by implementation.
	if res == nil {
		t.Error("expected non-nil slice from EntireList on empty list")
	}
}

func TestEntireList_ReturnsAllValues(t *testing.T) {
	list := newList(t)
	list.Insert("a", []byte("1"))
	list.Insert("b", []byte("2"))
	list.Insert("c", []byte("3"))
	res := list.EntireList()
	if len(res) < 3 {
		t.Errorf("expected at least 3 entries, got %d", len(res))
	}
}

// --- EmptyList ---

func TestEmptyList_SizeAfterEmpty(t *testing.T) {
	list := newList(t)
	list.Insert("a", []byte("1"))
	list.Insert("b", []byte("2"))
	list.EmptyList()
	// NOTE: EmptyList reassigns list internally; the caller's pointer is
	// unchanged. This test documents the current (possibly surprising)
	// behaviour — update if the implementation is fixed.
	size := list.size // no panic expected
	if size != 0 {
		t.Error("Expected an empty list", size)
	}
}

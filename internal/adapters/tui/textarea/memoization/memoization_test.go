package memoization_test

import (
	"encoding/binary"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/textarea/memoization"
)

type actionType int

const (
	set actionType = iota
	get
)

const fuzzCacheCapacity = 10

type cacheAction struct {
	actionType    actionType
	key           memoization.HString
	value         any
	expectedValue any
}

type testCase struct {
	name     string
	capacity int
	actions  []cacheAction
}

func TestCache(t *testing.T) {
	t.Parallel()

	tests := []testCase{
		{
			name:     "TestNewMemoCache",
			capacity: 5,
			actions: []cacheAction{
				{actionType: get, expectedValue: nil},
			},
		},
		{
			name:     "TestSetAndGet",
			capacity: 10,
			actions: []cacheAction{
				{actionType: set, key: "key1", value: "value1"},
				{actionType: get, key: "key1", expectedValue: "value1"},
				{actionType: set, key: "key1", value: "newValue1"},
				{actionType: get, key: "key1", expectedValue: "newValue1"},
				{actionType: get, key: "nonExistentKey", expectedValue: nil},
				{actionType: set, key: "nilKey", value: ""},
				{actionType: get, key: "nilKey", expectedValue: ""},
				{actionType: set, key: "keyA", value: "valueA"},
				{actionType: set, key: "keyB", value: "valueB"},
				{actionType: get, key: "keyA", expectedValue: "valueA"},
				{actionType: get, key: "keyB", expectedValue: "valueB"},
			},
		},
		{
			name:     "TestSetNilValue",
			capacity: 10,
			actions: []cacheAction{
				{actionType: set, key: "nilKey", value: nil},
				{actionType: get, key: "nilKey", expectedValue: nil},
			},
		},
		{
			name:     "TestGetAfterEviction",
			capacity: 2,
			actions: []cacheAction{
				{actionType: set, key: "1", value: 1},
				{actionType: set, key: "2", value: 2},
				{actionType: set, key: "3", value: 3},
				{actionType: get, key: "1", expectedValue: nil},
				{actionType: get, key: "2", expectedValue: 2},
			},
		},
		{
			name:     "TestGetAfterLRU",
			capacity: 2,
			actions: []cacheAction{
				{actionType: set, key: "1", value: 1},
				{actionType: set, key: "2", value: 2},
				{actionType: get, key: "1", expectedValue: 1},
				{actionType: set, key: "3", value: 3},
				{actionType: get, key: "1", expectedValue: 1},
				{actionType: get, key: "3", expectedValue: 3},
				{actionType: get, key: "2", expectedValue: nil},
			},
		},
		{
			name:     "TestLRU_Capacity3",
			capacity: 3,
			actions: []cacheAction{
				{actionType: set, key: "1", value: 1},
				{actionType: set, key: "2", value: 2},
				{actionType: set, key: "3", value: 3},
				{actionType: get, key: "1", expectedValue: 1}, // Accessing key "1"
				{
					actionType: set,
					key:        "4",
					value:      4,
				}, // Should evict key "2" since "1" was recently accessed
				{actionType: get, key: "2", expectedValue: nil},
				{actionType: get, key: "1", expectedValue: 1},
				{actionType: get, key: "3", expectedValue: 3},
				{actionType: get, key: "4", expectedValue: 4},
			},
		},
		// Test LRU behavior with varying accesses
		{
			name:     "TestLRU_VaryingAccesses",
			capacity: 3,
			actions: []cacheAction{
				{actionType: set, key: "1", value: 1},
				{actionType: set, key: "2", value: 2},
				{actionType: set, key: "3", value: 3},
				{actionType: get, key: "1", expectedValue: 1}, // Accessing key "1"
				{actionType: get, key: "2", expectedValue: 2}, // Accessing key "2"
				{actionType: set, key: "4", value: 4},         // Should evict key "3"
				{actionType: get, key: "3", expectedValue: nil},
				{actionType: get, key: "1", expectedValue: 1},
				{actionType: get, key: "2", expectedValue: 2},
				{actionType: get, key: "4", expectedValue: 4},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cache := memoization.NewMemoCache[memoization.HString, any](tt.capacity)
			for _, action := range tt.actions {
				applyCacheAction(t, cache, action)
			}
		})
	}
}

func applyCacheAction(t *testing.T, cache *memoization.MemoCache[memoization.HString, any], action cacheAction) {
	t.Helper()

	switch action.actionType {
	case set:
		cache.Set(action.key, action.value)
	case get:
		got, _ := cache.Get(action.key)
		assert.Equal(t, action.expectedValue, got, "Get(%v)", action.key)
	}
}

func FuzzCache(f *testing.F) {
	// Define some seed values for initial scenarios
	for _, seed := range [][]byte{
		[]byte("7\x010\x0000000020"),
		{0, 0, 0, 0}, // Set key 0 to 0
		{1, 0, 0, 1}, // Set key 0 to 1
		{2, 0},       // Get key 0
	} {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, in []byte) {
		if len(in) < 1 {
			t.Skip() // Skip the test if the input is less than 1 byte
		}

		run := fuzzRun{t: t, in: in, next: 0, cache: nil, model: nil}
		run.reset(fuzzCacheCapacity)

		for run.next < len(in) {
			run.step()
		}
	})
}

type fuzzRun struct {
	t     *testing.T
	in    []byte
	next  int
	cache *memoization.MemoCache[memoization.HInt, int]
	model *lruModel
}

func (r *fuzzRun) step() {
	opCode := r.take(1)[0] % 4 // Determine the operation: Set, Get, or Reset (added case for Reset)

	switch opCode {
	case 0, 1:
		r.set()
	case 2:
		r.get()
	case 3:
		r.resize()
	}
}

func (r *fuzzRun) take(length int) []byte {
	if r.next+length > len(r.in) {
		r.t.Skip() // Not enough input to continue, so skip
	}

	taken := r.in[r.next : r.next+length]
	r.next += length

	return taken
}

func (r *fuzzRun) set() {
	args := r.take(3)
	key := memoization.HInt(binary.BigEndian.Uint16(args[:2]))
	value := int(args[2])

	r.cache.Set(key, value)
	r.model.set(key, value, r.cache.Capacity())
}

func (r *fuzzRun) get() {
	key := memoization.HInt(r.take(1)[0])
	expectedValue := r.model.get(key)

	if got, _ := r.cache.Get(key); got != expectedValue {
		r.t.Fatalf("Get(%v) = %v, want %v", key, got, expectedValue) // The values do not match
	}
}

func (r *fuzzRun) resize() {
	newCacheSize := int(r.take(1)[0]) // Read the new cache size from the input
	if newCacheSize == 0 {
		r.t.Skip() // If the size is zero, we skip this test
	}

	r.reset(newCacheSize)
}

func (r *fuzzRun) reset(capacity int) {
	r.cache = memoization.NewMemoCache[memoization.HInt, int](capacity)
	r.model = &lruModel{expectedValues: make(map[memoization.HInt]int), accessOrder: make([]memoization.HInt, 0)}
}

type lruModel struct {
	expectedValues map[memoization.HInt]int
	accessOrder    []memoization.HInt
}

func (m *lruModel) set(key memoization.HInt, value, capacity int) {
	m.forget(key)
	m.expectedValues[key] = value
	m.accessOrder = append(m.accessOrder, key)

	// If we exceeded the cache size, we need to evict the least recently used item
	if len(m.accessOrder) > capacity {
		evictedKey := m.accessOrder[0]
		m.accessOrder = m.accessOrder[1:]

		delete(m.expectedValues, evictedKey)
	}
}

func (m *lruModel) get(key memoization.HInt) int {
	expectedValue, ok := m.expectedValues[key]
	if ok {
		m.forget(key)
		m.accessOrder = append(m.accessOrder, key)
	}

	return expectedValue
}

func (m *lruModel) forget(key memoization.HInt) {
	for index, accessedKey := range m.accessOrder {
		if accessedKey == key {
			m.accessOrder = slices.Delete(m.accessOrder, index, index+1)

			return
		}
	}
}

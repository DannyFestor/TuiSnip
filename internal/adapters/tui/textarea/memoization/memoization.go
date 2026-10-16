package memoization

import (
	"container/list"
	"crypto/sha256"
	"fmt"
	"sync"
)

type Hasher interface {
	Hash() string
}

type entry[T any] struct {
	key   string
	value T
}

type MemoCache[H Hasher, T any] struct {
	capacity     int
	mutex        sync.Mutex
	cache        map[string]*list.Element
	evictionList *list.List
}

func NewMemoCache[H Hasher, T any](capacity int) *MemoCache[H, T] {
	return &MemoCache[H, T]{
		capacity:     capacity,
		mutex:        sync.Mutex{},
		cache:        make(map[string]*list.Element),
		evictionList: list.New(),
	}
}

func (m *MemoCache[H, T]) Capacity() int {
	return m.capacity
}

func (m *MemoCache[H, T]) Size() int {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	return m.evictionList.Len()
}

func (m *MemoCache[H, T]) Get(h H) (T, bool) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if stored, found := m.lookup(h.Hash()); found {
		return stored.value, true
	}

	var result T

	return result, false
}

func (m *MemoCache[H, T]) Set(h H, value T) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	hashedKey := h.Hash()
	if stored, found := m.lookup(hashedKey); found {
		stored.value = value

		return
	}

	if m.evictionList.Len() >= m.capacity {
		m.evictLeastRecentlyUsed()
	}

	m.add(hashedKey, value)
}

func (m *MemoCache[H, T]) lookup(hashedKey string) (*entry[T], bool) {
	element, found := m.cache[hashedKey]
	if !found {
		return nil, false
	}

	m.evictionList.MoveToFront(element)
	stored, ok := element.Value.(*entry[T])

	return stored, ok
}

func (m *MemoCache[H, T]) evictLeastRecentlyUsed() {
	toEvict := m.evictionList.Back()
	if toEvict == nil {
		return
	}

	evictedEntry, ok := m.evictionList.Remove(toEvict).(*entry[T])
	if !ok {
		return
	}

	delete(m.cache, evictedEntry.key)
}

func (m *MemoCache[H, T]) add(hashedKey string, value T) {
	newEntry := &entry[T]{
		key:   hashedKey,
		value: value,
	}
	element := m.evictionList.PushFront(newEntry)
	m.cache[hashedKey] = element
}

type HString string

func (h HString) Hash() string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(h)))
}

type HInt int

func (h HInt) Hash() string {
	return fmt.Sprintf("%x", sha256.Sum256(fmt.Appendf(nil, "%d", h)))
}

package main

import (
	"fmt"
	"sync"
)

type Cache[K comparable, V any] struct {
	items map[K]V
	mu    sync.Mutex // !Mutex for controlling concurrent access to the cache
}

func New[K comparable, V any]() *Cache[K, V] {
	return &Cache[K, V]{
		items: make(map[K]V),
	}
}

func (c *Cache[K, V]) Set(key K, value V) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items[key] = value
}

func (c *Cache[K, V]) Get(key K) (V, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	value, found := c.items[key]
	return value, found
}

func (c *Cache[K, V]) Remove(key K) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.items, key)
}

func (c *Cache[K, V]) Pop(key K) (V, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	value, found := c.items[key]

	if found {
		delete(c.items, key)
	}

	return value, found
}

func main() {
	cache := New[string, int]()

	cache.Set("one", 1)
	cache.Set("two", 2)
	cache.Set("three", 3)

	value, found := cache.Get("two")
	if found {
		fmt.Printf("Value for key 'two': %v\n", value)
	} else {
		fmt.Println("Key 'two' not found in the cache")
	}

	poppedValue, found := cache.Pop("three")
	if found {
		fmt.Printf("Popped value for key 'three': %v\n", poppedValue)
	} else {
		fmt.Println("Key 'three' not found in the cache")
	}

	cache.Remove("one")

	removedValue, found := cache.Get("one")
	if found {
		fmt.Printf("Value for key 'one': %v\n", removedValue)
	} else {
		fmt.Println("Key 'one' not found in the cache (after removal)")
	}
}

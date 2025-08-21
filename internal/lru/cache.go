// Copyright 2022 Kyle Lemons
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package lru

import (
	"sync"
)

type syncValue[V any] struct {
	mu sync.Mutex
	v  V
	ok bool
}

type Cache[K comparable, V any] struct {
	n int

	mu    sync.Mutex
	order []K
	byKey map[K]*syncValue[V]
}

func NewCache[K comparable, V any](maxEntries int) *Cache[K, V] {
	return &Cache[K, V]{
		n:     maxEntries,
		byKey: make(map[K]*syncValue[V]),
	}
}

func (c *Cache[K, V]) Query(k K, populate func() (V, error)) (value V, err error) {
	sv := c.get(k)
	sv.mu.Lock()
	defer sv.mu.Unlock()
	if sv.ok {
		return sv.v, nil
	}

	v, err := populate()
	if err != nil {
		var zero V
		return zero, err
	}
	sv.v = v
	sv.ok = true
	return v, nil
}

func (c *Cache[K, V]) get(k K) *syncValue[V] {
	c.mu.Lock()
	defer c.mu.Unlock()
	hit, ok := c.byKey[k]
	if ok {
		return hit
	}

	miss := new(syncValue[V])
	c.order = append(c.order, k)
	for len(c.order) > c.n {
		k := c.order[0]
		c.order = c.order[1:]
		delete(c.byKey, k)
	}
	return miss
}

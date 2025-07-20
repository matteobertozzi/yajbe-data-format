// Licensed to the Apache Software Foundation (ASF) under one or more
// contributor license agreements.  See the NOTICE file distributed with
// this work for additional information regarding copyright ownership.
// The ASF licenses this file to You under the Apache License, Version 2.0
// (the "License"); you may not use this file except in compliance with
// the License.  You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package yajbe

// EnumMapping interface for string to index mapping
type EnumMapping interface {
	// Get returns the string at the given index
	Get(index int) (string, bool)
	// Add adds a string and returns its index, or -1 if not indexed
	Add(key string) int
}

// EnumMappingConfig represents configuration for enum mapping
type EnumMappingConfig interface {
	CreateMapping() EnumMapping
}

// LRUEnumMappingConfig configuration for LRU-based enum mapping
type LRUEnumMappingConfig struct {
	LRUSize int
	MinFreq int
}

func (c *LRUEnumMappingConfig) CreateMapping() EnumMapping {
	return NewLRUEnumMapping(c.LRUSize, c.MinFreq)
}

// LRUEnumMapping implements EnumMapping using an LRU cache
type LRUEnumMapping struct {
	lruSize    int
	minFreq    int
	strings    []string
	frequencies map[string]int
	stringToIndex map[string]int
	nextIndex  int
}

// NewLRUEnumMapping creates a new LRU-based enum mapping
func NewLRUEnumMapping(lruSize, minFreq int) *LRUEnumMapping {
	return &LRUEnumMapping{
		lruSize:       lruSize,
		minFreq:       minFreq,
		strings:       make([]string, 0, lruSize),
		frequencies:   make(map[string]int),
		stringToIndex: make(map[string]int),
		nextIndex:     0,
	}
}

// Get returns the string at the given index
func (m *LRUEnumMapping) Get(index int) (string, bool) {
	if index < 0 || index >= len(m.strings) {
		return "", false
	}
	return m.strings[index], true
}

// Add adds a string and returns its index, or -1 if not indexed
func (m *LRUEnumMapping) Add(key string) int {
	// Check if already indexed
	if index, exists := m.stringToIndex[key]; exists {
		return index
	}
	
	// Increment frequency
	m.frequencies[key]++
	
	// Check if frequency meets minimum threshold
	if m.frequencies[key] < m.minFreq {
		return -1
	}
	
	// Check if we have space in the LRU
	if len(m.strings) >= m.lruSize {
		// TODO: Implement proper LRU eviction
		return -1
	}
	
	// Add to mapping
	index := m.nextIndex
	m.strings = append(m.strings, key)
	m.stringToIndex[key] = index
	m.nextIndex++
	
	return index
}
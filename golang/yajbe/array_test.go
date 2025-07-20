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

import (
	"math/rand"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestArraySimple(t *testing.T) {
	// Test decode-only for special array encodings
	assertDecode(t, "20", []interface{}{})
	assertDecode(t, "2f01", []interface{}{})
	assertDecode(t, "2f6001", []interface{}{float64(0)}) // 0 decodes as float64 for JSON compatibility
	assertDecode(t, "2f606001", []interface{}{float64(0), float64(0)})
	assertDecode(t, "2f4001", []interface{}{float64(1)}) // 1 decodes as float64 for JSON compatibility
	assertDecode(t, "2f414101", []interface{}{float64(2), float64(2)})

	// Test encode/decode for regular arrays
	assertEncodeDecode(t, []interface{}{int64(1)}, "2140")
	assertEncodeDecode(t, []interface{}{int64(2), int64(2)}, "224141")
	
	// Test arrays with repeated elements
	zeroArray10 := make([]interface{}, 10)
	for i := range zeroArray10 {
		zeroArray10[i] = int64(0)
	}
	assertEncodeDecode(t, zeroArray10, "2a60606060606060606060")
	
	zeroArray11 := make([]interface{}, 11)
	for i := range zeroArray11 {
		zeroArray11[i] = int64(0)
	}
	assertEncodeDecode(t, zeroArray11, "2b016060606060606060606060")
	
	zeroArray255 := make([]interface{}, 0xff)
	for i := range zeroArray255 {
		zeroArray255[i] = int64(0)
	}
	assertEncodeDecode(t, zeroArray255, "2bf5"+strings.Repeat("60", 0xff))
	
	zeroArray265 := make([]interface{}, 265)
	for i := range zeroArray265 {
		zeroArray265[i] = int64(0)
	}
	assertEncodeDecode(t, zeroArray265, "2bff"+strings.Repeat("60", 265))
	
	zeroArrayLarge := make([]interface{}, 0xffff)
	for i := range zeroArrayLarge {
		zeroArrayLarge[i] = int64(0)
	}
	assertEncodeDecode(t, zeroArrayLarge, "2cf5ff"+strings.Repeat("60", 0xffff))
	
	zeroArrayXLarge := make([]interface{}, 0xffffff)
	for i := range zeroArrayXLarge {
		zeroArrayXLarge[i] = int64(0)
	}
	assertEncodeDecode(t, zeroArrayXLarge, "2df5ffff"+strings.Repeat("60", 0xffffff))

	// Test string arrays
	assertEncodeDecode(t, []interface{}{"a"}, "21c161")
	
	// Test null arrays
	nullArray255 := make([]interface{}, 0xff)
	for i := range nullArray255 {
		nullArray255[i] = nil
	}
	assertEncodeDecode(t, nullArray255, "2bf5"+strings.Repeat("00", 0xff))
	
	nullArray265 := make([]interface{}, 265)
	for i := range nullArray265 {
		nullArray265[i] = nil
	}
	assertEncodeDecode(t, nullArray265, "2bff"+strings.Repeat("00", 265))
}

func TestArraySmallLength(t *testing.T) {
	// Test arrays 0-9 elements (1 byte overhead)
	for i := 0; i < 10; i++ {
		input := make([]interface{}, i)
		for j := 0; j < i; j++ {
			input[j] = int64(i & 7)
		}
		
		enc, err := Marshal(input)
		require.NoError(t, err)
		assert.Equal(t, 1+i, len(enc))
		
		var decoded interface{}
		err = Unmarshal(enc, &decoded)
		require.NoError(t, err)
		
		// Convert expected values to float64 for JSON compatibility
		expected := make([]interface{}, i)
		for j := 0; j < i; j++ {
			expected[j] = float64(i & 7)
		}
		assert.Equal(t, expected, decoded)
	}

	// Test arrays 11-265 elements (2 byte overhead)
	for i := 11; i <= 265; i++ {
		input := make([]interface{}, i)
		for j := 0; j < i; j++ {
			input[j] = int64(i & 7)
		}
		
		enc, err := Marshal(input)
		require.NoError(t, err)
		assert.Equal(t, 2+i, len(enc))
		
		var decoded interface{}
		err = Unmarshal(enc, &decoded)
		require.NoError(t, err)
		
		// Convert expected values to float64 for JSON compatibility
		expected := make([]interface{}, i)
		for j := 0; j < i; j++ {
			expected[j] = float64(i & 7)
		}
		assert.Equal(t, expected, decoded)
	}

	// Test arrays 266-0xfff elements (3 byte overhead)
	for i := 266; i <= 0xfff; i++ {
		input := make([]interface{}, i)
		for j := 0; j < i; j++ {
			input[j] = int64(i & 7)
		}
		
		enc, err := Marshal(input)
		require.NoError(t, err)
		assert.Equal(t, 3+i, len(enc))
		
		var decoded interface{}
		err = Unmarshal(enc, &decoded)
		require.NoError(t, err)
		
		// Convert expected values to float64 for JSON compatibility
		expected := make([]interface{}, i)
		for j := 0; j < i; j++ {
			expected[j] = float64(i & 7)
		}
		assert.Equal(t, expected, decoded)
	}
}

func TestArrayRandEncodeDecode(t *testing.T) {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	
	for i := 0; i < 100; i++ {
		length := rng.Intn(1 << 10)
		input := make([]interface{}, length)
		
		for j := 0; j < length; j++ {
			input[j] = map[string]interface{}{
				"boolValue":  rng.Float64() > 0.5,
				"intValue":   int64(rng.Float64() * 2147483648),
				"floatValue": rng.Float64(),
			}
		}
		
		enc, err := Marshal(input)
		require.NoError(t, err)
		
		var decoded interface{}
		err = Unmarshal(enc, &decoded)
		require.NoError(t, err)
		
		// For complex objects, the structure should match but int values decode as float64
		if decodedSlice, ok := decoded.([]interface{}); ok {
			assert.Equal(t, len(input), len(decodedSlice))
			for j, inputVal := range input {
				if inputMap, ok := inputVal.(map[string]interface{}); ok {
					if decodedMap, ok := decodedSlice[j].(map[string]interface{}); ok {
						assert.Equal(t, inputMap["boolValue"], decodedMap["boolValue"])
						assert.Equal(t, float64(inputMap["intValue"].(int64)), decodedMap["intValue"]) // int64 -> float64
						assert.Equal(t, inputMap["floatValue"], decodedMap["floatValue"])
					}
				}
			}
		} else {
			t.Errorf("Expected []interface{}, got %T", decoded)
		}
	}
}

func TestArrayRandSet(t *testing.T) {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	input := make(map[string]bool) // Use map as set
	
	for i := 0; i < 16; i++ {
		key := "k" + strconv.FormatInt(int64(rng.Intn(2147483648)), 10)
		input[key] = true
		
		// Convert set to slice for encoding
		inputSlice := make([]interface{}, 0, len(input))
		for k := range input {
			inputSlice = append(inputSlice, k)
		}
		
		enc, err := Marshal(inputSlice)
		require.NoError(t, err)
		
		var decoded interface{}
		err = Unmarshal(enc, &decoded)
		require.NoError(t, err)
		
		// Verify all keys are present in decoded slice
		if decodedSlice, ok := decoded.([]interface{}); ok {
			assert.Equal(t, len(inputSlice), len(decodedSlice))
			decodedSet := make(map[string]bool)
			for _, item := range decodedSlice {
				if str, ok := item.(string); ok {
					decodedSet[str] = true
				}
			}
			assert.Equal(t, input, decodedSet)
		} else {
			t.Errorf("Expected []interface{}, got %T", decoded)
		}
	}
}
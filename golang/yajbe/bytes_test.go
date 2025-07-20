/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the 'License'); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an 'AS IS' BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package yajbe

import (
	"fmt"
	"strings"
	"testing"
)

func TestYajbeBytes(t *testing.T) {
	bt := NewBaseYajbeTest(t)

	t.Run("Simple", func(t *testing.T) {
		bt.assertEncodeDecode([]byte{}, "80")
		bt.assertEncodeDecode(make([]byte, 1), "8100")
		bt.assertEncodeDecode(make([]byte, 3), "83000000")
		bt.assertEncodeDecode(make([]byte, 59), "bb"+strings.Repeat("00", 59))
		bt.assertEncodeDecode(make([]byte, 60), "bc01"+strings.Repeat("00", 60))
		bt.assertEncodeDecode(make([]byte, 127), "bc44"+strings.Repeat("00", 127))
		bt.assertEncodeDecode(make([]byte, 0xff), "bcc4"+strings.Repeat("00", 255))
		bt.assertEncodeDecode(make([]byte, 256), "bcc5"+strings.Repeat("00", 256))
		bt.assertEncodeDecode(make([]byte, 314), "bcff"+strings.Repeat("00", 314))
		bt.assertEncodeDecode(make([]byte, 315), "bd0001"+strings.Repeat("00", 315))
		bt.assertEncodeDecode(make([]byte, 0xffff), "bdc4ff"+strings.Repeat("00", 0xffff))
	})

	t.Run("SmallByteLength", func(t *testing.T) {
		// Test byte arrays with length 0-59 (inline length encoding)
		for i := 0; i < 60; i++ {
			input := make([]byte, i)
			data, err := Marshal(input)
			if err != nil {
				t.Fatalf("Failed to marshal byte array of length %d: %v", i, err)
			}

			expectedSize := 1 + i // 1 byte header + byte array
			if len(data) != expectedSize {
				t.Errorf("Byte array length %d: expected %d bytes, got %d", i, expectedSize, len(data))
			}

			var result []byte
			err = Unmarshal(data, &result)
			if err != nil {
				t.Fatalf("Failed to unmarshal byte array of length %d: %v", i, err)
			}

			if len(result) != len(input) {
				t.Errorf("Byte array length %d round-trip failed: got length %d", i, len(result))
			}

			for j, expected := range input {
				if result[j] != expected {
					t.Errorf("Byte array element mismatch at index %d", j)
					break
				}
			}
		}

		// Test byte arrays with length 60-314 (1-byte length extension)
		testLengths := []int{60, 100, 200, 314}
		for _, i := range testLengths {
			input := make([]byte, i)
			data, err := Marshal(input)
			if err != nil {
				t.Fatalf("Failed to marshal byte array of length %d: %v", i, err)
			}

			expectedSize := 2 + i // 2 byte header + byte array
			if len(data) != expectedSize {
				t.Errorf("Byte array length %d: expected %d bytes, got %d", i, expectedSize, len(data))
			}

			var result []byte
			err = Unmarshal(data, &result)
			if err != nil {
				t.Fatalf("Failed to unmarshal byte array of length %d: %v", i, err)
			}

			if len(result) != len(input) {
				t.Errorf("Byte array length %d round-trip failed: got length %d", i, len(result))
			}
		}

		// Test byte arrays with length 315+ (2-byte length extension)
		testLengths = []int{315, 500, 1000, 8191}
		for _, i := range testLengths {
			input := make([]byte, i)
			data, err := Marshal(input)
			if err != nil {
				t.Fatalf("Failed to marshal byte array of length %d: %v", i, err)
			}

			expectedSize := 3 + i // 3 byte header + byte array
			if len(data) != expectedSize {
				t.Errorf("Byte array length %d: expected %d bytes, got %d", i, expectedSize, len(data))
			}

			var result []byte
			err = Unmarshal(data, &result)
			if err != nil {
				t.Fatalf("Failed to unmarshal byte array of length %d: %v", i, err)
			}

			if len(result) != len(input) {
				t.Errorf("Byte array length %d round-trip failed: got length %d", i, len(result))
			}
		}
	})

	t.Run("ByteValues", func(t *testing.T) {
		// Test all possible byte values
		allBytes := make([]byte, 256)
		for i := 0; i < 256; i++ {
			allBytes[i] = byte(i)
		}

		data, err := Marshal(allBytes)
		if err != nil {
			t.Fatalf("Failed to marshal all byte values: %v", err)
		}

		var result []byte
		err = Unmarshal(data, &result)
		if err != nil {
			t.Fatalf("Failed to unmarshal all byte values: %v", err)
		}

		if len(result) != len(allBytes) {
			t.Fatalf("All bytes length mismatch: expected %d, got %d", len(allBytes), len(result))
		}

		for i, expected := range allBytes {
			if result[i] != expected {
				t.Errorf("All bytes element %d mismatch: expected %d, got %d", i, expected, result[i])
			}
		}
	})

	t.Run("SpecialPatterns", func(t *testing.T) {
		// Test special byte patterns
		patterns := [][]byte{
			{},                             // Empty
			{0},                            // Single zero
			{255},                          // Single max
			{0, 255},                       // Zero and max
			{0x00, 0x01, 0x02, 0x03},       // Sequential
			{0xFF, 0xFE, 0xFD, 0xFC},       // Reverse sequential
			{0xAA, 0x55, 0xAA, 0x55},       // Alternating
			[]byte("Hello, World!"),        // ASCII text
			[]byte{0xC0, 0xFF, 0xEE},       // High values
			[]byte{0x80, 0x81, 0x82, 0x83}, // High bit set
		}

		for i, pattern := range patterns {
			data, err := Marshal(pattern)
			if err != nil {
				t.Fatalf("Failed to marshal pattern %d: %v", i, err)
			}

			var result []byte
			err = Unmarshal(data, &result)
			if err != nil {
				t.Fatalf("Failed to unmarshal pattern %d: %v", i, err)
			}

			if len(result) != len(pattern) {
				t.Errorf("Pattern %d length mismatch: expected %d, got %d", i, len(pattern), len(result))
				continue
			}

			for j, expected := range pattern {
				if result[j] != expected {
					t.Errorf("Pattern %d element %d mismatch: expected %d, got %d", i, j, expected, result[j])
				}
			}
		}
	})

	t.Run("Random", func(t *testing.T) {
		// Test random byte arrays
		for i := 0; i < 100; i++ {
			length := bt.random.Intn(1 << 20) // Up to 1MB
			input := make([]byte, length)
			bt.random.Read(input) // Fill with random bytes

			data, err := Marshal(input)
			if err != nil {
				t.Fatalf("Failed to marshal random byte array of length %d: %v", length, err)
			}

			var result []byte
			err = Unmarshal(data, &result)
			if err != nil {
				t.Fatalf("Failed to unmarshal random byte array of length %d: %v", length, err)
			}

			if len(result) != len(input) {
				t.Errorf("Random byte array length mismatch: expected %d, got %d", len(input), len(result))
				continue
			}

			// Compare bytes
			for j, expected := range input {
				if result[j] != expected {
					t.Errorf("Random byte array element %d mismatch: expected %d, got %d", j, expected, result[j])
					break // Don't spam on failure
				}
			}
		}
	})

	t.Run("LargeArrays", func(t *testing.T) {
		// Test large byte arrays for performance
		sizes := []int{1024, 10240, 102400} // 1KB, 10KB, 100KB

		for _, size := range sizes {
			t.Run(fmt.Sprintf("Size%d", size), func(t *testing.T) {
				largeArray := make([]byte, size)
				// Fill with a pattern for easier verification
				for i := range largeArray {
					largeArray[i] = byte(i % 256)
				}

				data, err := Marshal(largeArray)
				if err != nil {
					t.Fatalf("Failed to marshal large byte array of size %d: %v", size, err)
				}

				var result []byte
				err = Unmarshal(data, &result)
				if err != nil {
					t.Fatalf("Failed to unmarshal large byte array of size %d: %v", size, err)
				}

				if len(result) != len(largeArray) {
					t.Fatalf("Large byte array length mismatch: expected %d, got %d", len(largeArray), len(result))
				}

				// Verify some elements (don't check all for performance)
				checkIndices := []int{0, size / 4, size / 2, 3 * size / 4, size - 1}
				for _, i := range checkIndices {
					if result[i] != largeArray[i] {
						t.Errorf("Large byte array element %d mismatch: expected %d, got %d", i, largeArray[i], result[i])
					}
				}
			})
		}
	})

	t.Run("EmptyAndNil", func(t *testing.T) {
		// Test empty byte arrays
		emptyArray := []byte{}
		data, err := Marshal(emptyArray)
		if err != nil {
			t.Fatalf("Failed to marshal empty byte array: %v", err)
		}
		var resultEmpty []byte
		err = Unmarshal(data, &resultEmpty)
		if err != nil {
			t.Fatalf("Failed to unmarshal empty byte array: %v", err)
		}
		if len(resultEmpty) != 0 {
			t.Errorf("Empty byte array should have length 0, got %d", len(resultEmpty))
		}

		// Test nil byte slice
		var nilSlice []byte
		data, err = Marshal(nilSlice)
		if err != nil {
			t.Fatalf("Failed to marshal nil byte slice: %v", err)
		}
		var resultNil []byte
		err = Unmarshal(data, &resultNil)
		if err != nil {
			t.Fatalf("Failed to unmarshal nil byte slice: %v", err)
		}
		// Nil slices should unmarshal as empty slices
		if len(resultNil) != 0 {
			t.Errorf("Nil byte slice should unmarshal to empty slice, got length %d", len(resultNil))
		}
	})

	t.Run("StringConversion", func(t *testing.T) {
		// Test that byte arrays from strings work correctly
		testStrings := []string{
			"",
			"hello",
			"Hello, World!",
			"UTF-8: 🚀🎉💫",
			string([]byte{0, 1, 2, 3, 255}), // Binary data
		}

		for _, str := range testStrings {
			originalBytes := []byte(str)

			data, err := Marshal(originalBytes)
			if err != nil {
				t.Fatalf("Failed to marshal string bytes %q: %v", str, err)
			}

			var resultBytes []byte
			err = Unmarshal(data, &resultBytes)
			if err != nil {
				t.Fatalf("Failed to unmarshal string bytes %q: %v", str, err)
			}

			if string(resultBytes) != str {
				t.Errorf("String byte conversion failed: %q != %q", str, string(resultBytes))
			}
		}
	})

	t.Run("EncodingEfficiency", func(t *testing.T) {
		// Test that byte array encoding is efficient
		tests := []struct {
			bytes       []byte
			maxOverhead int // Maximum expected overhead in bytes
		}{
			{[]byte{}, 1},              // Empty should be 1 byte
			{[]byte{0}, 2},             // Single byte should be 2 bytes
			{make([]byte, 59), 60},     // 59 bytes should be 60 bytes
			{make([]byte, 60), 62},     // 60 bytes should be 62 bytes
			{make([]byte, 255), 257},   // 255 bytes should be 257 bytes
			{make([]byte, 1000), 1003}, // 1000 bytes should be ~1003 bytes
		}

		for _, test := range tests {
			data, err := Marshal(test.bytes)
			if err != nil {
				t.Fatalf("Failed to marshal byte array: %v", err)
			}

			if len(data) > test.maxOverhead {
				t.Errorf("Byte array encoding inefficient: %d bytes encoded to %d bytes (max expected %d)",
					len(test.bytes), len(data), test.maxOverhead)
			}
		}
	})

	t.Run("Comparison", func(t *testing.T) {
		// Compare YAJBE byte array encoding with JSON
		testByteArrays := [][]byte{
			{},
			{0, 1, 2, 3},
			make([]byte, 100),
			[]byte("Hello, World!"),
		}

		for i, bytes := range testByteArrays {
			yajbeSize, jsonSize, cborSize := bt.compareWithOthers(bytes)
			t.Logf("ByteArray[%d] len=%d: YAJBE=%d, JSON=%d (%.2f%%), CBOR=%d (%.2f%%)",
				i, len(bytes), yajbeSize, jsonSize, float64(jsonSize)/float64(yajbeSize)*100, cborSize, float64(cborSize)/float64(yajbeSize)*100)
		}
	})
}

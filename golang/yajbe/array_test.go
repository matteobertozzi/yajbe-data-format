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
	"testing"
)

func TestYajbeArrays(t *testing.T) {
	bt := NewBaseYajbeTest(t)

	t.Run("Simple", func(t *testing.T) {
		// Empty arrays
		bt.assertDecode("20", []any{})
		bt.assertDecode("2f01", []any{})

		// Integer arrays
		bt.assertDecode("2f01", []any{})
		bt.assertDecode("2f6001", []any{0})
		bt.assertDecode("2f606001", []any{0, 0})
		bt.assertDecode("2f4001", []any{1})
		bt.assertDecode("2f414101", []any{2, 2})

		// Test various array sizes
		bt.assertArrayEncodeDecode([]int{}, "20")
		bt.assertArrayEncodeDecode([]int{1}, "2140")
		bt.assertArrayEncodeDecode([]int{2, 2}, "224141")
		bt.assertArrayEncodeDecode(make([]int, 10), "2a"+
			"60606060606060606060") // 10 zeros
		bt.assertArrayEncodeDecode(make([]int, 11), "2b01"+
			"6060606060606060606060") // 11 zeros

		// String arrays
		bt.assertArrayEncodeDecode([]string{}, "20")
		bt.assertArrayEncodeDecode([]string{""}, "21c0")
		bt.assertArrayEncodeDecode([]string{"a"}, "21c161")
	})

	t.Run("SmallArrayLength", func(t *testing.T) {
		// Test arrays with length 0-10 (inline length encoding)
		for i := 0; i < 11; i++ {
			input := make([]int, i)
			data, err := Marshal(input)
			if err != nil {
				t.Fatalf("Failed to marshal array of length %d: %v", i, err)
			}

			expectedSize := 1 + i // 1 byte header + array elements
			if len(data) != expectedSize {
				t.Errorf("Array length %d: expected %d bytes, got %d", i, expectedSize, len(data))
			}

			var result []int
			err = Unmarshal(data, &result)
			if err != nil {
				t.Fatalf("Failed to unmarshal array of length %d: %v", i, err)
			}

			if len(result) != len(input) {
				t.Errorf("Array length %d round-trip failed: got length %d", i, len(result))
			}
		}

		// Test arrays with length 11-265 (1-byte length extension)
		testLengths := []int{11, 50, 100, 200, 265}
		for _, i := range testLengths {
			input := make([]int, i)
			data, err := Marshal(input)
			if err != nil {
				t.Fatalf("Failed to marshal array of length %d: %v", i, err)
			}

			expectedSize := 2 + i // 2 byte header + array elements
			if len(data) != expectedSize {
				t.Errorf("Array length %d: expected %d bytes, got %d", i, expectedSize, len(data))
			}

			var result []int
			err = Unmarshal(data, &result)
			if err != nil {
				t.Fatalf("Failed to unmarshal array of length %d: %v", i, err)
			}

			if len(result) != len(input) {
				t.Errorf("Array length %d round-trip failed: got length %d", i, len(result))
			}
		}

		// Test arrays with length 266+ (2-byte length extension)
		testLengths = []int{266, 500, 1000, 8191}
		for _, i := range testLengths {
			input := make([]int, i)
			data, err := Marshal(input)
			if err != nil {
				t.Fatalf("Failed to marshal array of length %d: %v", i, err)
			}

			expectedSize := 3 + i // 3 byte header + array elements
			if len(data) != expectedSize {
				t.Errorf("Array length %d: expected %d bytes, got %d", i, expectedSize, len(data))
			}

			var result []int
			err = Unmarshal(data, &result)
			if err != nil {
				t.Fatalf("Failed to unmarshal array of length %d: %v", i, err)
			}

			if len(result) != len(input) {
				t.Errorf("Array length %d round-trip failed: got length %d", i, len(result))
			}
		}
	})

	t.Run("TypedArrays", func(t *testing.T) {
		// Test arrays of different types
		intArray := []int{1, 2, 3, 4, 5}
		data, err := Marshal(intArray)
		if err != nil {
			t.Fatalf("Failed to marshal int array: %v", err)
		}
		var resultInt []int
		err = Unmarshal(data, &resultInt)
		if err != nil {
			t.Fatalf("Failed to unmarshal int array: %v", err)
		}
		if len(resultInt) != len(intArray) {
			t.Errorf("Int array length mismatch: %d != %d", len(resultInt), len(intArray))
		}

		stringArray := []string{"hello", "world", "test"}
		data, err = Marshal(stringArray)
		if err != nil {
			t.Fatalf("Failed to marshal string array: %v", err)
		}
		var resultString []string
		err = Unmarshal(data, &resultString)
		if err != nil {
			t.Fatalf("Failed to unmarshal string array: %v", err)
		}
		if len(resultString) != len(stringArray) {
			t.Errorf("String array length mismatch: %d != %d", len(resultString), len(stringArray))
		}

		floatArray := []float64{1.1, 2.2, 3.3}
		data, err = Marshal(floatArray)
		if err != nil {
			t.Fatalf("Failed to marshal float array: %v", err)
		}
		var resultFloat []float64
		err = Unmarshal(data, &resultFloat)
		if err != nil {
			t.Fatalf("Failed to unmarshal float array: %v", err)
		}
		if len(resultFloat) != len(floatArray) {
			t.Errorf("Float array length mismatch: %d != %d", len(resultFloat), len(floatArray))
		}
	})

	t.Run("NestedArrays", func(t *testing.T) {
		// Test arrays of arrays
		nestedIntArray := [][]int{{1, 2}, {3, 4, 5}, {}}
		data, err := Marshal(nestedIntArray)
		if err != nil {
			t.Fatalf("Failed to marshal nested int array: %v", err)
		}
		var resultNested [][]int
		err = Unmarshal(data, &resultNested)
		if err != nil {
			t.Fatalf("Failed to unmarshal nested int array: %v", err)
		}
		if len(resultNested) != len(nestedIntArray) {
			t.Errorf("Nested array length mismatch: %d != %d", len(resultNested), len(nestedIntArray))
		}

		// Test mixed type arrays (any)
		mixedArray := []any{1, "hello", 3.14, true, nil}
		data, err = Marshal(mixedArray)
		if err != nil {
			t.Fatalf("Failed to marshal mixed array: %v", err)
		}
		var resultMixed []any
		err = Unmarshal(data, &resultMixed)
		if err != nil {
			t.Fatalf("Failed to unmarshal mixed array: %v", err)
		}
		if len(resultMixed) != len(mixedArray) {
			t.Errorf("Mixed array length mismatch: %d != %d", len(resultMixed), len(mixedArray))
		}
	})

	t.Run("LargeArrays", func(t *testing.T) {
		// Test performance with large arrays
		sizes := []int{1000, 10000}

		for _, size := range sizes {
			t.Run(fmt.Sprintf("Size%d", size), func(t *testing.T) {
				largeArray := bt.randIntBlock(size)

				data, err := Marshal(largeArray)
				if err != nil {
					t.Fatalf("Failed to marshal large array of size %d: %v", size, err)
				}

				var result []int
				err = Unmarshal(data, &result)
				if err != nil {
					t.Fatalf("Failed to unmarshal large array of size %d: %v", size, err)
				}

				if len(result) != len(largeArray) {
					t.Fatalf("Large array length mismatch: expected %d, got %d", len(largeArray), len(result))
				}

				// Verify some elements (don't check all for performance)
				checkIndices := []int{0, size / 4, size / 2, 3 * size / 4, size - 1}
				for _, i := range checkIndices {
					if result[i] != largeArray[i] {
						t.Errorf("Large array element %d mismatch: expected %d, got %d", i, largeArray[i], result[i])
					}
				}
			})
		}
	})

	t.Run("EmptyAndNil", func(t *testing.T) {
		// Test empty arrays
		emptyIntArray := []int{}
		data, err := Marshal(emptyIntArray)
		if err != nil {
			t.Fatalf("Failed to marshal empty int array: %v", err)
		}
		var resultEmpty []int
		err = Unmarshal(data, &resultEmpty)
		if err != nil {
			t.Fatalf("Failed to unmarshal empty int array: %v", err)
		}
		if len(resultEmpty) != 0 {
			t.Errorf("Empty array should have length 0, got %d", len(resultEmpty))
		}

		// Test nil slice
		var nilSlice []int
		data, err = Marshal(nilSlice)
		if err != nil {
			t.Fatalf("Failed to marshal nil slice: %v", err)
		}
		var resultNil []int
		err = Unmarshal(data, &resultNil)
		if err != nil {
			t.Fatalf("Failed to unmarshal nil slice: %v", err)
		}
		// Nil slices should unmarshal as empty slices
		if len(resultNil) != 0 {
			t.Errorf("Nil slice should unmarshal to empty slice, got length %d", len(resultNil))
		}
	})

	t.Run("Arrays2D", func(t *testing.T) {
		// Test 2D arrays
		array2D := [][]int{
			{1, 2, 3},
			{4, 5},
			{6, 7, 8, 9},
			{},
		}

		data, err := Marshal(array2D)
		if err != nil {
			t.Fatalf("Failed to marshal 2D array: %v", err)
		}

		var result2D [][]int
		err = Unmarshal(data, &result2D)
		if err != nil {
			t.Fatalf("Failed to unmarshal 2D array: %v", err)
		}

		if len(result2D) != len(array2D) {
			t.Fatalf("2D array length mismatch: expected %d, got %d", len(array2D), len(result2D))
		}

		for i, row := range array2D {
			if len(result2D[i]) != len(row) {
				t.Errorf("2D array row %d length mismatch: expected %d, got %d", i, len(row), len(result2D[i]))
				continue
			}
			for j, val := range row {
				if result2D[i][j] != val {
					t.Errorf("2D array[%d][%d] mismatch: expected %d, got %d", i, j, val, result2D[i][j])
				}
			}
		}
	})

	t.Run("Random", func(t *testing.T) {
		// Test random arrays
		for i := 0; i < 100; i++ {
			length := bt.random.Intn(1000)
			originalArray := bt.randIntBlock(length)

			data, err := Marshal(originalArray)
			if err != nil {
				t.Fatalf("Failed to marshal random array of length %d: %v", length, err)
			}

			var result []int
			err = Unmarshal(data, &result)
			if err != nil {
				t.Fatalf("Failed to unmarshal random array of length %d: %v", length, err)
			}

			if len(result) != len(originalArray) {
				t.Errorf("Random array length mismatch: expected %d, got %d", len(originalArray), len(result))
				continue
			}

			for j, expected := range originalArray {
				if result[j] != expected {
					t.Errorf("Random array element %d mismatch: expected %d, got %d", j, expected, result[j])
					break // Don't spam on failure
				}
			}
		}
	})

	t.Run("EncodingEfficiency", func(t *testing.T) {
		// Test that array encoding is efficient
		tests := []struct {
			array       []int
			maxOverhead int // Maximum expected overhead in bytes
		}{
			{[]int{}, 1},            // Empty array should be 1 byte
			{[]int{0}, 2},           // Single zero should be 2 bytes
			{make([]int, 10), 11},   // 10 zeros should be 11 bytes
			{make([]int, 100), 102}, // 100 zeros should be 102 bytes
		}

		for _, test := range tests {
			data, err := Marshal(test.array)
			if err != nil {
				t.Fatalf("Failed to marshal array: %v", err)
			}

			if len(data) > test.maxOverhead {
				t.Errorf("Array encoding inefficient: %d elements encoded to %d bytes (max expected %d)",
					len(test.array), len(data), test.maxOverhead)
			}
		}
	})

	t.Run("Comparison", func(t *testing.T) {
		// Compare YAJBE array encoding with JSON
		testArrays := [][]int{
			{},
			{1, 2, 3},
			make([]int, 100),
			bt.randIntBlock(1000),
		}

		for i, arr := range testArrays {
			yajbeSize, jsonSize, cborSize := bt.compareWithOthers(arr)
			t.Logf("Array[%d] len=%d: YAJBE=%d, JSON=%d (%.2f%%), CBOR=%d (%.2f%%)",
				i, len(arr), yajbeSize, jsonSize, float64(jsonSize)/float64(yajbeSize)*100, cborSize, float64(cborSize)/float64(yajbeSize)*100)
		}
	})
}

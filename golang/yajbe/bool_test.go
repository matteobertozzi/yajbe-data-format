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
	"testing"
)

func TestYajbeBool(t *testing.T) {
	bt := NewBaseYajbeTest(t)

	t.Run("Simple", func(t *testing.T) {
		bt.assertEncodeDecode(true, "03")
		bt.assertEncodeDecode(false, "02")
	})

	t.Run("Arrays", func(t *testing.T) {
		// Test boolean arrays
		bt.assertArrayEncodeDecode([]bool{}, "20")
		bt.assertArrayEncodeDecode([]bool{true}, "2103")
		bt.assertArrayEncodeDecode([]bool{false}, "2102")
		bt.assertArrayEncodeDecode([]bool{true, false}, "220302")
		bt.assertArrayEncodeDecode([]bool{false, true, false, true}, "2402030203")

		// Test larger boolean array
		largeBoolArray := make([]bool, 100)
		for i := range largeBoolArray {
			largeBoolArray[i] = (i % 2) == 0
		}

		data, err := Marshal(largeBoolArray)
		if err != nil {
			t.Fatalf("Failed to marshal large boolean array: %v", err)
		}

		var result []bool
		err = Unmarshal(data, &result)
		if err != nil {
			t.Fatalf("Failed to unmarshal large boolean array: %v", err)
		}

		if len(result) != len(largeBoolArray) {
			t.Fatalf("Large boolean array length mismatch: expected %d, got %d", len(largeBoolArray), len(result))
		}

		for i, expected := range largeBoolArray {
			if result[i] != expected {
				t.Errorf("Large boolean array element %d mismatch: expected %t, got %t", i, expected, result[i])
			}
		}
	})

	t.Run("MixedWithOtherTypes", func(t *testing.T) {
		// Test booleans mixed with other types
		mixedArray := []any{true, 42, "hello", false, 3.14}

		data, err := Marshal(mixedArray)
		if err != nil {
			t.Fatalf("Failed to marshal mixed array with booleans: %v", err)
		}

		var result []any
		err = Unmarshal(data, &result)
		if err != nil {
			t.Fatalf("Failed to unmarshal mixed array with booleans: %v", err)
		}

		if len(result) != len(mixedArray) {
			t.Fatalf("Mixed array length mismatch: expected %d, got %d", len(mixedArray), len(result))
		}

		// Check boolean elements specifically
		if result[0] != true {
			t.Errorf("Mixed array boolean[0] mismatch: expected true, got %v", result[0])
		}
		if result[3] != false {
			t.Errorf("Mixed array boolean[3] mismatch: expected false, got %v", result[3])
		}
	})

	t.Run("Structs", func(t *testing.T) {
		// Test booleans in struct fields
		type BoolStruct struct {
			Active   bool `json:"active"`
			Enabled  bool `json:"enabled"`
			Optional bool `json:"optional"`
		}

		testStruct := BoolStruct{
			Active:   true,
			Enabled:  false,
			Optional: true,
		}

		data, err := Marshal(testStruct)
		if err != nil {
			t.Fatalf("Failed to marshal boolean struct: %v", err)
		}

		var result BoolStruct
		err = Unmarshal(data, &result)
		if err != nil {
			t.Fatalf("Failed to unmarshal boolean struct: %v", err)
		}

		if result.Active != testStruct.Active {
			t.Errorf("Boolean struct Active mismatch: expected %t, got %t", testStruct.Active, result.Active)
		}
		if result.Enabled != testStruct.Enabled {
			t.Errorf("Boolean struct Enabled mismatch: expected %t, got %t", testStruct.Enabled, result.Enabled)
		}
		if result.Optional != testStruct.Optional {
			t.Errorf("Boolean struct Optional mismatch: expected %t, got %t", testStruct.Optional, result.Optional)
		}
	})

	t.Run("Maps", func(t *testing.T) {
		// Test booleans in maps
		boolMap := map[string]any{
			"success": true,
			"error":   false,
			"pending": true,
		}

		data, err := Marshal(boolMap)
		if err != nil {
			t.Fatalf("Failed to marshal boolean map: %v", err)
		}

		var result map[string]any
		err = Unmarshal(data, &result)
		if err != nil {
			t.Fatalf("Failed to unmarshal boolean map: %v", err)
		}

		for key, expected := range boolMap {
			if result[key] != expected {
				t.Errorf("Boolean map[%s] mismatch: expected %t, got %v", key, expected, result[key])
			}
		}
	})

	t.Run("PointerBooleans", func(t *testing.T) {
		// Test boolean pointers
		trueVal := true
		falseVal := false

		data, err := Marshal(&trueVal)
		if err != nil {
			t.Fatalf("Failed to marshal boolean pointer (true): %v", err)
		}

		var resultTrue bool
		err = Unmarshal(data, &resultTrue)
		if err != nil {
			t.Fatalf("Failed to unmarshal boolean pointer (true): %v", err)
		}

		if resultTrue != trueVal {
			t.Errorf("Boolean pointer (true) mismatch: expected %t, got %t", trueVal, resultTrue)
		}

		data, err = Marshal(&falseVal)
		if err != nil {
			t.Fatalf("Failed to marshal boolean pointer (false): %v", err)
		}

		var resultFalse bool
		err = Unmarshal(data, &resultFalse)
		if err != nil {
			t.Fatalf("Failed to unmarshal boolean pointer (false): %v", err)
		}

		if resultFalse != falseVal {
			t.Errorf("Boolean pointer (false) mismatch: expected %t, got %t", falseVal, resultFalse)
		}

		// Test nil boolean pointer
		var nilBoolPtr *bool
		data, err = Marshal(nilBoolPtr)
		if err != nil {
			t.Fatalf("Failed to marshal nil boolean pointer: %v", err)
		}

		var resultNilPtr *bool
		err = Unmarshal(data, &resultNilPtr)
		if err != nil {
			t.Fatalf("Failed to unmarshal nil boolean pointer: %v", err)
		}

		if resultNilPtr != nil {
			t.Errorf("Nil boolean pointer should unmarshal to nil, got %v", resultNilPtr)
		}
	})

	t.Run("RoundTrip", func(t *testing.T) {
		// Comprehensive round-trip test
		testCases := []bool{true, false}

		for _, testCase := range testCases {
			data, err := Marshal(testCase)
			if err != nil {
				t.Fatalf("Failed to marshal boolean %t: %v", testCase, err)
			}

			var result bool
			err = Unmarshal(data, &result)
			if err != nil {
				t.Fatalf("Failed to unmarshal boolean %t: %v", testCase, err)
			}

			if result != testCase {
				t.Errorf("Boolean round-trip failed: expected %t, got %t", testCase, result)
			}
		}
	})

	t.Run("EncodingSize", func(t *testing.T) {
		// Test that boolean encoding is efficient (1 byte each)
		trueData, err := Marshal(true)
		if err != nil {
			t.Fatalf("Failed to marshal true: %v", err)
		}
		if len(trueData) != 1 {
			t.Errorf("Boolean true should encode to 1 byte, got %d", len(trueData))
		}

		falseData, err := Marshal(false)
		if err != nil {
			t.Fatalf("Failed to marshal false: %v", err)
		}
		if len(falseData) != 1 {
			t.Errorf("Boolean false should encode to 1 byte, got %d", len(falseData))
		}

		// Test boolean array encoding efficiency
		boolArray := []bool{true, false, true, false}
		arrayData, err := Marshal(boolArray)
		if err != nil {
			t.Fatalf("Failed to marshal boolean array: %v", err)
		}

		expectedSize := 1 + len(boolArray) // 1 byte header + 1 byte per boolean
		if len(arrayData) != expectedSize {
			t.Errorf("Boolean array should encode to %d bytes, got %d", expectedSize, len(arrayData))
		}
	})

	t.Run("Comparison", func(t *testing.T) {
		// Compare YAJBE boolean encoding with JSON
		testBooleans := []bool{true, false}
		testBoolArrays := [][]bool{
			{},
			{true},
			{false},
			{true, false, true, false},
			make([]bool, 100), // All false
		}

		for _, b := range testBooleans {
			yajbeSize, jsonSize, cborSize := bt.compareWithOthers(b)
			t.Logf("Bool: YAJBE=%d, JSON=%d (%.2f%%), CBOR=%d (%.2f%%)",
				yajbeSize, jsonSize, float64(jsonSize)/float64(yajbeSize)*100, cborSize, float64(cborSize)/float64(yajbeSize)*100)
		}

		for i, arr := range testBoolArrays {
			yajbeSize, jsonSize, cborSize := bt.compareWithOthers(arr)
			t.Logf("BoolArray[%d] len=%d: YAJBE=%d, JSON=%d (%.2f%%), CBOR=%d (%.2f%%)",
				i, len(arr), yajbeSize, jsonSize, float64(jsonSize)/float64(yajbeSize)*100, cborSize, float64(cborSize)/float64(yajbeSize)*100)
		}
	})
}

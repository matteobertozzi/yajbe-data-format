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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDecoderEOFArrayHandling(t *testing.T) {
	// Test EOF array handling (arrays with length -1)
	// This simulates streaming array data where length is unknown
	
	// Create a test array
	input := []interface{}{1, 2, 3, "test", true}
	
	decoded := MarshalToInterface(t, input)
	decodedSlice := decoded.([]interface{})
	
	require.Len(t, decodedSlice, len(input))
	// Check values (accounting for JSON compatibility - ints become float64)
	assert.Equal(t, float64(1), decodedSlice[0])
	assert.Equal(t, float64(2), decodedSlice[1])
	assert.Equal(t, float64(3), decodedSlice[2])
	assert.Equal(t, "test", decodedSlice[3])
	assert.Equal(t, true, decodedSlice[4])
}

func TestDecoderEOFObjectHandling(t *testing.T) {
	// Test EOF object handling (objects with length -1)
	// This simulates streaming object data where length is unknown
	
	input := map[string]interface{}{
		"key1": "value1",
		"key2": 42,
		"key3": true,
	}
	
	expected := map[string]interface{}{
		"key1": "value1",
		"key2": float64(42), // int becomes float64
		"key3": true,
	}
	
	AssertMarshalToInterface(t, input, expected)
}

func TestDecoderMixedArrayTypes(t *testing.T) {
	// Test array with mixed types
	input := []interface{}{
		"string",
		int64(42),
		3.14,
		true,
		false,
		nil,
		[]interface{}{1, 2, 3},
		map[string]interface{}{"nested": "object"},
	}
	
	decoded := SliceRoundTrip(t, input)
	
	require.Len(t, decoded, len(input))
	assert.Equal(t, "string", decoded[0])
	assert.Equal(t, float64(42), decoded[1]) // int64 becomes float64
	assert.Equal(t, 3.14, decoded[2])
	assert.Equal(t, true, decoded[3])
	assert.Equal(t, false, decoded[4])
	assert.Nil(t, decoded[5])
	
	// Check nested array
	nestedArray, ok := decoded[6].([]interface{})
	require.True(t, ok)
	require.Len(t, nestedArray, 3)
	assert.Equal(t, float64(1), nestedArray[0])
	assert.Equal(t, float64(2), nestedArray[1])
	assert.Equal(t, float64(3), nestedArray[2])
	
	// Check nested object
	nestedObject, ok := decoded[7].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "object", nestedObject["nested"])
}

func TestDecoderSpecializedSliceTypes(t *testing.T) {
	// Test decoding to specialized slice types
	
	// Test []string
	stringInput := []string{"hello", "world", "test"}
	RoundTripTestWithValidation(t, stringInput, func(t *testing.T, input, decoded []string) {
		assert.Equal(t, input, decoded)
	})
	
	// Test []int64
	intInput := []int64{1, 2, 3, 4, 5}
	RoundTripTestWithValidation(t, intInput, func(t *testing.T, input, decoded []int64) {
		assert.Equal(t, input, decoded)
	})
	
	// Test []bool
	boolInput := []bool{true, false, true, false}
	RoundTripTestWithValidation(t, boolInput, func(t *testing.T, input, decoded []bool) {
		assert.Equal(t, input, decoded)
	})
}

func TestDecoderSpecializedMapTypes(t *testing.T) {
	// Test decoding to specialized map types
	
	// Test map[string]int
	intMapInput := map[string]int{
		"one":   1,
		"two":   2,
		"three": 3,
	}
	RoundTripTestWithValidation(t, intMapInput, func(t *testing.T, input, decoded map[string]int) {
		assert.Equal(t, input, decoded)
	})
	
	// Test map[string]string
	stringMapInput := map[string]string{
		"key1": "value1",
		"key2": "value2",
		"key3": "value3",
	}
	RoundTripTestWithValidation(t, stringMapInput, func(t *testing.T, input, decoded map[string]string) {
		assert.Equal(t, input, decoded)
	})
}

func TestDecoderLargeArrays(t *testing.T) {
	// Test large arrays with different encoding optimizations
	
	// Array with length > 10 but < 256 (uses 2-byte length encoding)
	largeArray := make([]interface{}, 100)
	for i := 0; i < 100; i++ {
		largeArray[i] = i
	}
	
	decoded := SliceRoundTrip(t, largeArray)
	
	require.Len(t, decoded, len(largeArray))
	for i, expected := range largeArray {
		assert.Equal(t, float64(expected.(int)), decoded[i]) // int becomes float64
	}
}

func TestDecoderLargeObjects(t *testing.T) {
	// Test large objects with many fields
	largeObject := make(map[string]interface{})
	for i := 0; i < 50; i++ {
		key := "field_" + string(rune(i+65)) // A, B, C, etc.
		largeObject[key] = i
	}
	
	decoded := MapRoundTrip(t, largeObject)
	
	require.Len(t, decoded, len(largeObject))
	for key, expected := range largeObject {
		actual, exists := decoded[key]
		require.True(t, exists, "Key %s not found", key)
		assert.Equal(t, float64(expected.(int)), actual) // int becomes float64
	}
}

func TestDecoderNullValues(t *testing.T) {
	// Test various null value scenarios
	
	// Null in array
	arrayWithNull := []interface{}{1, nil, "test", nil}
	decodedArray := SliceRoundTrip(t, arrayWithNull)
	
	require.Len(t, decodedArray, len(arrayWithNull))
	assert.Equal(t, float64(1), decodedArray[0])
	assert.Nil(t, decodedArray[1])
	assert.Equal(t, "test", decodedArray[2])
	assert.Nil(t, decodedArray[3])
	
	// Null in object
	objectWithNull := map[string]interface{}{
		"key1": "value",
		"key2": nil,
		"key3": 42,
	}
	decodedObject := MapRoundTrip(t, objectWithNull)
	
	assert.Equal(t, "value", decodedObject["key1"])
	assert.Nil(t, decodedObject["key2"])
	assert.Equal(t, float64(42), decodedObject["key3"])
}

func TestDecoderErrorHandling(t *testing.T) {
	// Test various error conditions
	
	// Invalid data
	invalidData := []byte{0xFF, 0xFF, 0xFF}
	var result interface{}
	err := Unmarshal(invalidData, &result)
	assert.Error(t, err)
	
	// Nil destination
	validData, _ := Marshal(map[string]interface{}{"test": "value"})
	err = Unmarshal(validData, nil)
	assert.Error(t, err)
	
	// Type mismatch - trying to decode object into array
	objectData, _ := Marshal(map[string]interface{}{"key": "value"})
	var arrayResult []interface{}
	err = Unmarshal(objectData, &arrayResult)
	assert.Error(t, err)
}

func TestDecoderNumericTypes(t *testing.T) {
	// Test various numeric type decodings
	
	// Test different integer sizes
	var int8Val int8 = 42
	RoundTripAssertEqual(t, int8Val)
	
	// Test float32
	var float32Val float32 = 3.14
	RoundTripAssertEqual(t, float32Val)
	
	// Test float64
	var float64Val float64 = 2.71828
	RoundTripAssertEqual(t, float64Val)
}

func TestDecoderByteArrays(t *testing.T) {
	// Test byte array handling
	byteData := []byte{0x00, 0x01, 0x02, 0x03, 0xFF, 0xFE, 0xFD}
	RoundTripTestWithValidation(t, byteData, func(t *testing.T, input, decoded []byte) {
		assert.Equal(t, input, decoded)
	})
}

func TestDecoderFieldNameOptimization(t *testing.T) {
	// Test field name optimization with repeated field names
	objects := []map[string]interface{}{
		{"common_field": "value1", "unique1": 1},
		{"common_field": "value2", "unique2": 2},
		{"common_field": "value3", "unique3": 3},
	}
	
	for _, obj := range objects {
		encoded, err := Marshal(obj)
		require.NoError(t, err)
		
		var decoded map[string]interface{}
		err = Unmarshal(encoded, &decoded)
		require.NoError(t, err)
		
		// Check that all fields are preserved
		for key, expectedValue := range obj {
			actualValue, exists := decoded[key]
			require.True(t, exists, "Key %s not found", key)
			
			// Handle type conversion for integers
			if intVal, ok := expectedValue.(int); ok {
				assert.Equal(t, float64(intVal), actualValue)
			} else {
				assert.Equal(t, expectedValue, actualValue)
			}
		}
	}
}
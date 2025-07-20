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

// RoundTripTest performs a complete marshal/unmarshal test for any type
func RoundTripTest[T any](t *testing.T, input T) T {
	t.Helper()
	
	encoded, err := Marshal(input)
	require.NoError(t, err)
	
	var decoded T
	err = Unmarshal(encoded, &decoded)
	require.NoError(t, err)
	
	return decoded
}

// RoundTripTestWithValidation performs marshal/unmarshal and runs custom validation
func RoundTripTestWithValidation[T any](t *testing.T, input T, validator func(t *testing.T, input, decoded T)) {
	t.Helper()
	
	decoded := RoundTripTest(t, input)
	validator(t, input, decoded)
}

// RoundTripAssertEqual performs marshal/unmarshal and asserts equality
func RoundTripAssertEqual[T comparable](t *testing.T, input T) {
	t.Helper()
	
	decoded := RoundTripTest(t, input)
	assert.Equal(t, input, decoded)
}

// MarshalToInterface marshals input and unmarshals to interface{} for JSON compatibility testing
func MarshalToInterface(t *testing.T, input interface{}) interface{} {
	t.Helper()
	
	encoded, err := Marshal(input)
	require.NoError(t, err)
	
	var decoded interface{}
	err = Unmarshal(encoded, &decoded)
	require.NoError(t, err)
	
	return decoded
}

// AssertMarshalToInterface marshals input, unmarshals to interface{}, and validates with expected
func AssertMarshalToInterface(t *testing.T, input interface{}, expected interface{}) {
	t.Helper()
	
	decoded := MarshalToInterface(t, input)
	assert.Equal(t, expected, decoded)
}

// TestStructField is a generic helper for testing struct field equality
func AssertStructField[T any](t *testing.T, fieldName string, expected, actual T) {
	t.Helper()
	assert.Equal(t, expected, actual, "Field %s mismatch", fieldName)
}

// AssertSliceEqual compares slices with proper type conversion for JSON compatibility
func AssertSliceEqual[T any](t *testing.T, expected []T, actual []interface{}, converter func(T) interface{}) {
	t.Helper()
	
	require.Len(t, actual, len(expected), "Slice length mismatch")
	
	for i, expectedItem := range expected {
		expectedConverted := converter(expectedItem)
		assert.Equal(t, expectedConverted, actual[i], "Slice item %d mismatch", i)
	}
}

// AssertMapStructFields validates map[string]interface{} representation of structs
func AssertMapStructFields(t *testing.T, decoded map[string]interface{}, assertions map[string]interface{}) {
	t.Helper()
	
	for fieldName, expectedValue := range assertions {
		actualValue, exists := decoded[fieldName]
		require.True(t, exists, "Field %s not found in decoded map", fieldName)
		assert.Equal(t, expectedValue, actualValue, "Field %s value mismatch", fieldName)
	}
}

// IntToFloat64 converter for JSON compatibility (int becomes float64)
func IntToFloat64(i int) interface{} {
	return float64(i)
}

// Int64ToFloat64 converter for JSON compatibility (int64 becomes float64)
func Int64ToFloat64(i int64) interface{} {
	return float64(i)
}

// IdentityConverter for types that don't need conversion
func IdentityConverter[T any](t T) interface{} {
	return t
}

// TestErrorCondition tests that an operation produces an expected error
func TestErrorCondition(t *testing.T, operation func() error, expectError bool) {
	t.Helper()
	
	err := operation()
	if expectError {
		assert.Error(t, err)
	} else {
		assert.NoError(t, err)
	}
}

// TestMarshalError tests that marshaling produces an error
func TestMarshalError(t *testing.T, input interface{}, expectError bool) {
	t.Helper()
	
	TestErrorCondition(t, func() error {
		_, err := Marshal(input)
		return err
	}, expectError)
}

// TestUnmarshalError tests that unmarshaling produces an error
func TestUnmarshalError(t *testing.T, data []byte, target interface{}, expectError bool) {
	t.Helper()
	
	TestErrorCondition(t, func() error {
		return Unmarshal(data, target)
	}, expectError)
}

// BatchRoundTripTest tests multiple inputs of the same type
func BatchRoundTripTest[T comparable](t *testing.T, testCases []struct {
	name  string
	input T
}) {
	t.Helper()
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			RoundTripAssertEqual(t, tc.input)
		})
	}
}

// BatchRoundTripTestWithValidation tests multiple inputs with custom validation
func BatchRoundTripTestWithValidation[T any](t *testing.T, testCases []struct {
	name  string
	input T
}, validator func(t *testing.T, input, decoded T)) {
	t.Helper()
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			RoundTripTestWithValidation(t, tc.input, validator)
		})
	}
}

// Common struct types used across tests (moved from individual files)
type SimpleStruct struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

type NestedStruct struct {
	ID     int64        `json:"id"`
	Person SimpleStruct `json:"person"`
	Active bool         `json:"active"`
}

type StructWithPointers struct {
	Name   *string       `json:"name"`
	Age    *int          `json:"age"`
	Nested *SimpleStruct `json:"nested"`
}

type StructWithSlices struct {
	Names   []string       `json:"names"`
	Numbers []int          `json:"numbers"`
	People  []SimpleStruct `json:"people"`
}

type StructWithMaps struct {
	StringMap map[string]string      `json:"stringMap"`
	IntMap    map[string]int         `json:"intMap"`
	NestedMap map[string]interface{} `json:"nestedMap"`
}

type ComplexStruct struct {
	Basic    SimpleStruct           `json:"basic"`
	Slice    []string               `json:"slice"`
	Map      map[string]int         `json:"map"`
	Pointer  *SimpleStruct          `json:"pointer"`
	Optional *string               `json:"optional,omitempty"`
}

// Additional test helpers to reduce boilerplate patterns

// QuickMarshal performs marshaling with automatic error handling (for testing only)
func QuickMarshal(t *testing.T, input interface{}) []byte {
	t.Helper()
	encoded, err := Marshal(input)
	require.NoError(t, err)
	return encoded
}

// QuickUnmarshal performs unmarshaling with automatic error handling (for testing only)
func QuickUnmarshal[T any](t *testing.T, data []byte) T {
	t.Helper()
	var decoded T
	err := Unmarshal(data, &decoded)
	require.NoError(t, err)
	return decoded
}

// AssertRoundTripWithConversion for types that need conversion during round-trip
func AssertRoundTripWithConversion[T, U any](t *testing.T, input T, converter func(T) U) {
	t.Helper()
	encoded := QuickMarshal(t, input)
	var decoded T
	err := Unmarshal(encoded, &decoded)
	require.NoError(t, err)
	expected := converter(input)
	convertedDecoded := converter(decoded)
	assert.Equal(t, expected, convertedDecoded)
}

// MultiTypeRoundTrip tests that the same data can be decoded into different compatible types
func MultiTypeRoundTrip[T, U any](t *testing.T, input T) (T, U) {
	t.Helper()
	encoded := QuickMarshal(t, input)
	decodedT := QuickUnmarshal[T](t, encoded)
	decodedU := QuickUnmarshal[U](t, encoded)
	return decodedT, decodedU
}

// AssertNoRoundTripLoss verifies that marshal->unmarshal preserves all data
func AssertNoRoundTripLoss[T comparable](t *testing.T, input T) {
	t.Helper()
	encoded := QuickMarshal(t, input)
	decoded := QuickUnmarshal[T](t, encoded)
	assert.Equal(t, input, decoded, "Round-trip should preserve all data")
}

// CreateTestCases creates a slice of test cases from a map for convenience
func CreateTestCases[T any](cases map[string]T) []struct {
	name  string
	input T
} {
	result := make([]struct {
		name  string
		input T
	}, 0, len(cases))
	
	for name, input := range cases {
		result = append(result, struct {
			name  string
			input T
		}{name: name, input: input})
	}
	
	return result
}

// AssertMapContains checks that a decoded map contains expected key-value pairs
func AssertMapContains(t *testing.T, decoded map[string]interface{}, expectedPairs map[string]interface{}) {
	t.Helper()
	for key, expectedValue := range expectedPairs {
		actualValue, exists := decoded[key]
		require.True(t, exists, "Key %s should exist in decoded map", key)
		
		// Handle integer to float64 conversion for JSON compatibility
		if intVal, ok := expectedValue.(int); ok {
			expectedValue = float64(intVal)
		}
		if int64Val, ok := expectedValue.(int64); ok {
			expectedValue = float64(int64Val)
		}
		
		assert.Equal(t, expectedValue, actualValue, "Value for key %s should match", key)
	}
}

// AssertSliceContains checks that a decoded slice contains expected elements
func AssertSliceContains(t *testing.T, decoded []interface{}, expectedElements []interface{}) {
	t.Helper()
	require.Len(t, decoded, len(expectedElements), "Slice should have expected length")
	
	for i, expectedElement := range expectedElements {
		// Handle integer to float64 conversion for JSON compatibility
		if intVal, ok := expectedElement.(int); ok {
			expectedElement = float64(intVal)
		}
		if int64Val, ok := expectedElement.(int64); ok {
			expectedElement = float64(int64Val)
		}
		
		assert.Equal(t, expectedElement, decoded[i], "Element at index %d should match", i)
	}
}

// SliceRoundTrip specifically for []interface{} with automatic type conversion
func SliceRoundTrip(t *testing.T, input []interface{}) []interface{} {
	t.Helper()
	encoded := QuickMarshal(t, input)
	return QuickUnmarshal[[]interface{}](t, encoded)
}

// MapRoundTrip specifically for map[string]interface{} with automatic type conversion
func MapRoundTrip(t *testing.T, input map[string]interface{}) map[string]interface{} {
	t.Helper()
	encoded := QuickMarshal(t, input)
	return QuickUnmarshal[map[string]interface{}](t, encoded)
}
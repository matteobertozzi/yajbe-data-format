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
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDecoderEmptyContainers(t *testing.T) {
	// Test empty arrays
	emptyArray := []interface{}{}
	decoded := RoundTripTest(t, emptyArray)
	assert.Empty(t, decoded)
	
	// Test empty objects
	emptyObject := map[string]interface{}{}
	decodedObject := RoundTripTest(t, emptyObject)
	assert.Empty(t, decodedObject)
}

func TestDecoderNestedEmptyContainers(t *testing.T) {
	// Test nested empty containers
	nested := map[string]interface{}{
		"empty_array":  []interface{}{},
		"empty_object": map[string]interface{}{},
		"nested": map[string]interface{}{
			"also_empty_array":  []interface{}{},
			"also_empty_object": map[string]interface{}{},
		},
	}
	
	decoded := MapRoundTrip(t, nested)
	
	// Check empty array
	emptyArr, exists := decoded["empty_array"].([]interface{})
	require.True(t, exists)
	assert.Empty(t, emptyArr)
	
	// Check empty object
	emptyObj, exists := decoded["empty_object"].(map[string]interface{})
	require.True(t, exists)
	assert.Empty(t, emptyObj)
	
	// Check nested empty containers
	nestedObj, exists := decoded["nested"].(map[string]interface{})
	require.True(t, exists)
	
	nestedEmptyArr, exists := nestedObj["also_empty_array"].([]interface{})
	require.True(t, exists)
	assert.Empty(t, nestedEmptyArr)
	
	nestedEmptyObj, exists := nestedObj["also_empty_object"].(map[string]interface{})
	require.True(t, exists)
	assert.Empty(t, nestedEmptyObj)
}

func TestDecoderSingleElementContainers(t *testing.T) {
	// Test arrays with single element
	singleArray := []interface{}{42}
	decodedArray := SliceRoundTrip(t, singleArray)
	require.Len(t, decodedArray, 1)
	assert.Equal(t, float64(42), decodedArray[0])
	
	// Test objects with single field
	singleObject := map[string]interface{}{"key": "value"}
	decodedObject := MapRoundTrip(t, singleObject)
	require.Len(t, decodedObject, 1)
	assert.Equal(t, "value", decodedObject["key"])
}

func TestDecoderVeryLargeNumbers(t *testing.T) {
	// Test very large integers that might hit encoding limits
	largeInt := int64(9223372036854775807) // Max int64
	encoded, err := Marshal(largeInt)
	require.NoError(t, err)
	
	// When decoding to interface{}, it becomes float64 for JSON compatibility
	var decodedInterface interface{}
	err = Unmarshal(encoded, &decodedInterface)
	require.NoError(t, err)
	assert.Equal(t, float64(largeInt), decodedInterface)
	
	// When decoding to int64, it should preserve the value
	var decodedInt64 int64
	err = Unmarshal(encoded, &decodedInt64)
	require.NoError(t, err)
	assert.Equal(t, largeInt, decodedInt64)
}

func TestDecoderSpecialFloats(t *testing.T) {
	// Test special float values
	testCases := []struct {
		name  string
		input float64
	}{
		{"zero", 0.0},
		{"negative_zero", -0.0},
		{"small_positive", 1e-10},
		{"small_negative", -1e-10},
		{"large_positive", 1e10},
		{"large_negative", -1e10},
	}
	
	BatchRoundTripTest(t, testCases)
}

func TestDecoderLongStrings(t *testing.T) {
	// Test very long strings
	longString := strings.Repeat("a", 10000)
	RoundTripAssertEqual(t, longString)
}

func TestDecoderUnicodeStrings(t *testing.T) {
	// Test Unicode strings
	testCases := []struct {
		name  string
		input string
	}{
		{"chinese", "Hello, 世界"},
		{"emojis", "🚀🌟⭐"},
		{"spanish", "Ñoño niño"},
		{"arabic", "المرحبا"},
		{"japanese", "こんにちは"},
		{"pride_flags", "🏳️‍🌈🏳️‍⚧️"},
	}
	
	BatchRoundTripTest(t, testCases)
}

func TestDecoderMixedNesting(t *testing.T) {
	// Test mixed deep nesting of arrays and objects
	deeply_nested := map[string]interface{}{
		"level1": []interface{}{
			map[string]interface{}{
				"level2": []interface{}{
					map[string]interface{}{
						"level3": []interface{}{
							"deep_value",
							map[string]interface{}{
								"level4": "very_deep_value",
							},
						},
					},
				},
			},
		},
	}
	
	encoded, err := Marshal(deeply_nested)
	require.NoError(t, err)
	
	var decoded map[string]interface{}
	err = Unmarshal(encoded, &decoded)
	require.NoError(t, err)
	
	// Navigate through the nested structure to verify correctness
	level1, exists := decoded["level1"].([]interface{})
	require.True(t, exists)
	require.Len(t, level1, 1)
	
	level2_container, ok := level1[0].(map[string]interface{})
	require.True(t, ok)
	level2, exists := level2_container["level2"].([]interface{})
	require.True(t, exists)
	require.Len(t, level2, 1)
	
	level3_container, ok := level2[0].(map[string]interface{})
	require.True(t, ok)
	level3, exists := level3_container["level3"].([]interface{})
	require.True(t, exists)
	require.Len(t, level3, 2)
	
	assert.Equal(t, "deep_value", level3[0])
	
	level4_container, ok := level3[1].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "very_deep_value", level4_container["level4"])
}

func TestDecoderCircularStructReferences(t *testing.T) {
	// Test struct types that could have circular references (but don't in our data)
	type Node struct {
		Value string `json:"value"`
		Child *Node  `json:"child,omitempty"`
	}
	
	// Create a non-circular structure
	node := Node{
		Value: "root",
		Child: &Node{
			Value: "child",
			// No circular reference back to parent
		},
	}
	
	encoded, err := Marshal(node)
	require.NoError(t, err)
	
	var decoded Node
	err = Unmarshal(encoded, &decoded)
	require.NoError(t, err)
	
	assert.Equal(t, "root", decoded.Value)
	require.NotNil(t, decoded.Child)
	assert.Equal(t, "child", decoded.Child.Value)
	// Note: YAJBE creates zero-value instances for null pointer fields
	assert.NotNil(t, decoded.Child.Child)
	assert.Equal(t, "", decoded.Child.Child.Value)
}

func TestDecoderBooleanEdgeCases(t *testing.T) {
	// Test boolean values in different contexts
	boolTests := []struct {
		name  string
		input interface{}
	}{
		{"true_value", true},
		{"false_value", false},
		{"bool_in_array", []interface{}{true, false, true}},
		{"bool_in_object", map[string]interface{}{"active": true, "disabled": false}},
		{"mixed_bool_array", []interface{}{true, int64(1), false, "test", nil}},
	}
	
	for _, test := range boolTests {
		t.Run(test.name, func(t *testing.T) {
			encoded, err := Marshal(test.input)
			require.NoError(t, err)
			
			var decoded interface{}
			err = Unmarshal(encoded, &decoded)
			require.NoError(t, err)
			
			// Convert expected values for JSON compatibility where needed
			expected := convertToExpectedDecoded(test.input)
			assert.Equal(t, expected, decoded)
		})
	}
}

func TestDecoderFieldNameCaseSensitivity(t *testing.T) {
	// Test that field names are case-sensitive
	type CaseStruct struct {
		Field string `json:"field"`
		FIELD string `json:"FIELD"`
		Field_Name string `json:"field_name"`
		FieldName string `json:"fieldName"`
	}
	
	input := map[string]interface{}{
		"field": "lowercase",
		"FIELD": "uppercase", 
		"field_name": "underscore",
		"fieldName": "camelCase",
	}
	
	encoded, err := Marshal(input)
	require.NoError(t, err)
	
	var decoded CaseStruct
	err = Unmarshal(encoded, &decoded)
	require.NoError(t, err)
	
	assert.Equal(t, "lowercase", decoded.Field)
	assert.Equal(t, "uppercase", decoded.FIELD)
	assert.Equal(t, "underscore", decoded.Field_Name)
	assert.Equal(t, "camelCase", decoded.FieldName)
}

func TestDecoderSpecialCharacterFields(t *testing.T) {
	// Test field names with special characters
	input := map[string]interface{}{
		"field-with-dashes":   "dash_value",
		"field.with.dots":     "dot_value",
		"field with spaces":   "space_value",
		"field@with#symbols":  "symbol_value",
		"数字字段":                "unicode_field",
		"🚀rocket":            "emoji_field",
	}
	
	encoded, err := Marshal(input)
	require.NoError(t, err)
	
	var decoded map[string]interface{}
	err = Unmarshal(encoded, &decoded)
	require.NoError(t, err)
	
	for key, expectedValue := range input {
		actualValue, exists := decoded[key]
		require.True(t, exists, "Key %s not found", key)
		assert.Equal(t, expectedValue, actualValue)
	}
}
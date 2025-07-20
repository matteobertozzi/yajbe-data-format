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

func TestStructDecoderBasic(t *testing.T) {
	// Test basic struct decoding
	input := SimpleStruct{Name: "John", Age: 30}
	RoundTripAssertEqual(t, input)
}

func TestStructDecoderNested(t *testing.T) {
	// Test nested struct decoding
	input := NestedStruct{
		ID:     123,
		Person: SimpleStruct{Name: "Alice", Age: 25},
		Active: true,
	}
	RoundTripAssertEqual(t, input)
}

func TestStructDecoderWithPointers(t *testing.T) {
	// Test struct with pointer fields
	name := "Bob"
	age := 35
	nested := SimpleStruct{Name: "Carol", Age: 28}
	
	input := StructWithPointers{
		Name:   &name,
		Age:    &age,
		Nested: &nested,
	}
	
	RoundTripTestWithValidation(t, input, func(t *testing.T, input, decoded StructWithPointers) {
		require.NotNil(t, decoded.Name)
		require.NotNil(t, decoded.Age)
		require.NotNil(t, decoded.Nested)
		
		assert.Equal(t, *input.Name, *decoded.Name)
		assert.Equal(t, *input.Age, *decoded.Age)
		assert.Equal(t, input.Nested.Name, decoded.Nested.Name)
		assert.Equal(t, input.Nested.Age, decoded.Nested.Age)
	})
}

func TestStructDecoderWithNilPointers(t *testing.T) {
	// Test struct with nil pointer fields
	// Note: YAJBE marshals nil pointer fields as null values
	// When unmarshaling, null values to struct pointer fields create zero-value instances
	input := StructWithPointers{
		Name:   nil,
		Age:    nil,
		Nested: nil,
	}
	
	RoundTripTestWithValidation(t, input, func(t *testing.T, input, decoded StructWithPointers) {
		// YAJBE creates zero-value instances for null values when decoding to pointer fields
		require.NotNil(t, decoded.Name)
		assert.Equal(t, "", *decoded.Name)
		
		require.NotNil(t, decoded.Age)
		assert.Equal(t, 0, *decoded.Age)
		
		require.NotNil(t, decoded.Nested)
		assert.Equal(t, "", decoded.Nested.Name)
		assert.Equal(t, 0, decoded.Nested.Age)
	})
}

func TestStructDecoderWithSlices(t *testing.T) {
	// Test struct with slice fields
	input := StructWithSlices{
		Names:   []string{"Alice", "Bob", "Charlie"},
		Numbers: []int{1, 2, 3, 4, 5},
		People: []SimpleStruct{
			{Name: "Dave", Age: 30},
			{Name: "Eve", Age: 25},
		},
	}
	
	RoundTripTestWithValidation(t, input, func(t *testing.T, input, decoded StructWithSlices) {
		assert.Equal(t, input.Names, decoded.Names)
		assert.Equal(t, input.Numbers, decoded.Numbers)
		require.Len(t, decoded.People, len(input.People))
		for i, person := range input.People {
			assert.Equal(t, person.Name, decoded.People[i].Name)
			assert.Equal(t, person.Age, decoded.People[i].Age)
		}
	})
}

func TestStructDecoderWithMaps(t *testing.T) {
	// Test struct with map fields
	input := StructWithMaps{
		StringMap: map[string]string{
			"key1": "value1",
			"key2": "value2",
		},
		IntMap: map[string]int{
			"a": 1,
			"b": 2,
		},
		NestedMap: map[string]interface{}{
			"person1": SimpleStruct{Name: "Alice", Age: 30},
			"person2": SimpleStruct{Name: "Bob", Age: 25},
		},
	}
	
	encoded, err := Marshal(input)
	require.NoError(t, err)
	
	var decoded StructWithMaps
	err = Unmarshal(encoded, &decoded)
	require.NoError(t, err)
	
	assert.Equal(t, input.StringMap, decoded.StringMap)
	assert.Equal(t, input.IntMap, decoded.IntMap)
	
	require.Len(t, decoded.NestedMap, len(input.NestedMap))
	
	// Check person1
	person1, exists := decoded.NestedMap["person1"].(map[string]interface{})
	require.True(t, exists)
	assert.Equal(t, "Alice", person1["name"])
	assert.Equal(t, float64(30), person1["age"]) // int becomes float64
	
	// Check person2
	person2, exists := decoded.NestedMap["person2"].(map[string]interface{})
	require.True(t, exists)
	assert.Equal(t, "Bob", person2["name"])
	assert.Equal(t, float64(25), person2["age"]) // int becomes float64
}

func TestStructDecoderComplex(t *testing.T) {
	// Test complex struct with multiple field types
	optional := "optional value"
	input := ComplexStruct{
		Basic:   SimpleStruct{Name: "John", Age: 30},
		Slice:   []string{"a", "b", "c"},
		Map:     map[string]int{"x": 1, "y": 2},
		Pointer: &SimpleStruct{Name: "Jane", Age: 25},
		Optional: &optional,
	}
	
	encoded, err := Marshal(input)
	require.NoError(t, err)
	
	var decoded ComplexStruct
	err = Unmarshal(encoded, &decoded)
	require.NoError(t, err)
	
	assert.Equal(t, input.Basic.Name, decoded.Basic.Name)
	assert.Equal(t, input.Basic.Age, decoded.Basic.Age)
	assert.Equal(t, input.Slice, decoded.Slice)
	assert.Equal(t, input.Map, decoded.Map)
	
	require.NotNil(t, decoded.Pointer)
	assert.Equal(t, input.Pointer.Name, decoded.Pointer.Name)
	assert.Equal(t, input.Pointer.Age, decoded.Pointer.Age)
	
	require.NotNil(t, decoded.Optional)
	assert.Equal(t, optional, *decoded.Optional)
}

func TestStructDecoderUnknownFields(t *testing.T) {
	// Test that unknown fields are ignored during decoding
	// We'll encode a map with extra fields and decode to a struct with fewer fields
	input := map[string]interface{}{
		"name":    "John",
		"age":     30,
		"unknown": "should be ignored",
		"extra":   123,
	}
	
	encoded, err := Marshal(input)
	require.NoError(t, err)
	
	var decoded SimpleStruct
	err = Unmarshal(encoded, &decoded)
	require.NoError(t, err)
	
	assert.Equal(t, "John", decoded.Name)
	assert.Equal(t, 30, decoded.Age)
}

func TestStructDecoderJSONTags(t *testing.T) {
	// Test that JSON tags are respected
	type TaggedStruct struct {
		FieldA string `json:"field_a"`
		FieldB int    `json:"field_b"`
		FieldC bool   `json:"field_c"`
	}
	
	// Encode using a map with the JSON tag names
	input := map[string]interface{}{
		"field_a": "test",
		"field_b": 42,
		"field_c": true,
	}
	
	encoded, err := Marshal(input)
	require.NoError(t, err)
	
	var decoded TaggedStruct
	err = Unmarshal(encoded, &decoded)
	require.NoError(t, err)
	
	assert.Equal(t, "test", decoded.FieldA)
	assert.Equal(t, 42, decoded.FieldB)
	assert.Equal(t, true, decoded.FieldC)
}

func TestStructDecoderEmptyStruct(t *testing.T) {
	// Test empty struct
	type EmptyStruct struct{}
	
	input := EmptyStruct{}
	RoundTripAssertEqual(t, input)
}

func TestStructDecoderArrayOfStructs(t *testing.T) {
	// Test array of structs
	input := []SimpleStruct{
		{Name: "Alice", Age: 25},
		{Name: "Bob", Age: 30},
		{Name: "Charlie", Age: 35},
	}
	
	RoundTripTestWithValidation(t, input, func(t *testing.T, input, decoded []SimpleStruct) {
		require.Len(t, decoded, len(input))
		for i, expected := range input {
			assert.Equal(t, expected.Name, decoded[i].Name)
			assert.Equal(t, expected.Age, decoded[i].Age)
		}
	})
}

func TestStructDecoderMapOfStructs(t *testing.T) {
	// Test map of structs (using map[string]interface{} since custom struct maps aren't supported)
	input := map[string]interface{}{
		"person1": SimpleStruct{Name: "Alice", Age: 25},
		"person2": SimpleStruct{Name: "Bob", Age: 30},
	}
	
	encoded, err := Marshal(input)
	require.NoError(t, err)
	
	var decoded map[string]interface{}
	err = Unmarshal(encoded, &decoded)
	require.NoError(t, err)
	
	require.Len(t, decoded, len(input))
	
	// Check person1
	person1, exists := decoded["person1"].(map[string]interface{})
	require.True(t, exists)
	assert.Equal(t, "Alice", person1["name"])
	assert.Equal(t, float64(25), person1["age"]) // int becomes float64
	
	// Check person2
	person2, exists := decoded["person2"].(map[string]interface{})
	require.True(t, exists)
	assert.Equal(t, "Bob", person2["name"])
	assert.Equal(t, float64(30), person2["age"]) // int becomes float64
}

func TestStructDecoderDeeplyNested(t *testing.T) {
	// Test deeply nested structures
	type Level3 struct {
		Value string `json:"value"`
	}
	
	type Level2 struct {
		Level3 Level3 `json:"level3"`
		Data   []int  `json:"data"`
	}
	
	type Level1 struct {
		Level2 Level2            `json:"level2"`
		Map    map[string]string `json:"map"`
	}
	
	input := Level1{
		Level2: Level2{
			Level3: Level3{Value: "deep value"},
			Data:   []int{1, 2, 3},
		},
		Map: map[string]string{"key": "value"},
	}
	
	RoundTripTestWithValidation(t, input, func(t *testing.T, input, decoded Level1) {
		assert.Equal(t, input.Level2.Level3.Value, decoded.Level2.Level3.Value)
		assert.Equal(t, input.Level2.Data, decoded.Level2.Data)
		assert.Equal(t, input.Map, decoded.Map)
	})
}
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
	"bytes"
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStreamingBasic(t *testing.T) {
	// Test basic streaming functionality with io.Writer/Reader
	var buffer bytes.Buffer

	// Test writing to stream
	err := MarshalToWriter(&buffer, map[string]interface{}{
		"name":   "John",
		"age":    30,
		"active": true,
	})
	require.NoError(t, err)

	// Debug: Print what was written to the buffer
	data := buffer.Bytes()
	fmt.Printf("Encoded data: %s (len=%d)\n", hex.EncodeToString(data), len(data))

	// Test basic Marshal/Unmarshal with the same data for comparison
	testObj := map[string]interface{}{
		"name":   "John",
		"age":    30,
		"active": true,
	}
	basicData, err := Marshal(testObj)
	require.NoError(t, err)
	fmt.Printf("Basic Marshal data: %s (len=%d)\n", hex.EncodeToString(basicData), len(basicData))

	var basicResult map[string]interface{}
	err = Unmarshal(basicData, &basicResult)
	if err != nil {
		fmt.Printf("Basic Unmarshal error: %v\n", err)
	} else {
		fmt.Printf("Basic Unmarshal success: %+v\n", basicResult)
	}

	// Test reading from stream
	var result map[string]interface{}
	err = UnmarshalFromReader(&buffer, &result)
	if err != nil {
		fmt.Printf("UnmarshalFromReader error: %v\n", err)

		// Try to decode with basic Unmarshal using the same data
		var testResult map[string]interface{}
		testErr := Unmarshal(data, &testResult)
		if testErr != nil {
			fmt.Printf("Basic Unmarshal of stream data error: %v\n", testErr)
		} else {
			fmt.Printf("Basic Unmarshal of stream data success: %+v\n", testResult)
		}
	}
	require.NoError(t, err)

	// Verify contents (accounting for JSON compatibility)
	expected := map[string]interface{}{
		"name":   "John",
		"age":    float64(30), // int becomes float64
		"active": true,
	}
	assert.Equal(t, expected, result)
}

func TestStreamingMultipleObjects(t *testing.T) {
	// Test streaming multiple objects to same writer
	var buffer bytes.Buffer

	objects := []interface{}{
		map[string]interface{}{"id": 1, "name": "Alice"},
		map[string]interface{}{"id": 2, "name": "Bob"},
		map[string]interface{}{"id": 3, "name": "Charlie"},
	}

	// Write all objects to stream
	for _, obj := range objects {
		err := MarshalToWriter(&buffer, obj)
		require.NoError(t, err)
	}

	// Read all objects from stream
	var results []interface{}
	for i := 0; i < len(objects); i++ {
		var result map[string]interface{}
		err := UnmarshalFromReader(&buffer, &result)
		require.NoError(t, err)
		results = append(results, result)
	}

	// Verify all objects
	for i, result := range results {
		resultMap := result.(map[string]interface{})
		assert.Equal(t, float64(i+1), resultMap["id"]) // int becomes float64

		names := []string{"Alice", "Bob", "Charlie"}
		assert.Equal(t, names[i], resultMap["name"])
	}
}

func TestStreamingFieldNameOptimization(t *testing.T) {
	// Test that repeated field names in streaming are optimized
	var buffer bytes.Buffer

	// Create objects with same field names to test compression
	objects := []interface{}{
		map[string]interface{}{"common_field": "value1", "another_field": 1},
		map[string]interface{}{"common_field": "value2", "another_field": 2},
		map[string]interface{}{"common_field": "value3", "another_field": 3},
	}

	// Write all objects
	for _, obj := range objects {
		err := MarshalToWriter(&buffer, obj)
		require.NoError(t, err)
	}

	// The buffer should be smaller than if no field name compression was used
	// This is a basic test - we don't verify the exact compression amount
	assert.Greater(t, buffer.Len(), 0, "Buffer should contain encoded data")

	// Verify we can decode everything correctly
	originalData := buffer.Bytes()
	buffer = *bytes.NewBuffer(originalData)

	for i := 0; i < len(objects); i++ {
		var result map[string]interface{}
		err := UnmarshalFromReader(&buffer, &result)
		require.NoError(t, err)

		assert.Equal(t, fmt.Sprintf("value%d", i+1), result["common_field"])
		assert.Equal(t, float64(i+1), result["another_field"]) // int becomes float64
	}
}

func TestStreamingWithBufferWriter(t *testing.T) {
	// Test streaming using BufferWriter directly
	writer := NewWriterFromBuffer(make([]byte, 0, 256))
	encoder := Encoder{writer: writer}

	// Encode multiple values
	testValues := []interface{}{
		"hello",
		42,
		true,
		[]interface{}{1, 2, 3},
		map[string]interface{}{"key": "value"},
	}

	for _, value := range testValues {
		err := encoder.Encode(value)
		require.NoError(t, err)
	}

	// Create reader from buffer
	encoded := writer.Bytes()
	reader := NewReaderFromBytes(encoded)
	decoder := Decoder{reader: reader}

	// Decode all values
	for i, expectedValue := range testValues {
		var decoded interface{}
		err := decoder.Decode(&decoded)
		require.NoError(t, err, "Failed to decode value at index %d", i)

		// Handle type conversions for JSON compatibility
		switch expected := expectedValue.(type) {
		case int:
			assert.Equal(t, float64(expected), decoded)
		case []interface{}:
			// Array elements become float64
			decodedSlice := decoded.([]interface{})
			assert.Len(t, decodedSlice, len(expected))
			for j, elem := range expected {
				assert.Equal(t, float64(elem.(int)), decodedSlice[j])
			}
		case map[string]interface{}:
			// String values stay the same
			assert.Equal(t, expected, decoded)
		default:
			assert.Equal(t, expected, decoded)
		}
	}
}

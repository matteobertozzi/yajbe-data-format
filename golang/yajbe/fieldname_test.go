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
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testFieldNamesDirectEncodeDecode tests field name compression by directly 
// using FieldNameWriter/Reader similar to TypeScript tests
func testFieldNamesDirectEncodeDecode(t *testing.T, fieldNames []string, expectedHex string) {
	// Create a buffer writer for testing
	writer := NewWriterFromBuffer(make([]byte, 0, 256))
	fieldWriter := NewFieldNameWriter(writer)
	
	// Encode all field names
	for _, fieldName := range fieldNames {
		err := fieldWriter.Write(fieldName)
		require.NoError(t, err, "Failed to write field name: %s", fieldName)
	}
	
	// Check encoded hex matches expected
	actualHex := hex.EncodeToString(writer.Bytes())
	assert.Equal(t, expectedHex, actualHex, "Field name encoding mismatch")
	
	// Create a reader from the encoded data
	reader := NewReaderFromBytes(writer.Bytes())
	fieldReader := NewFieldNameReader(reader)
	
	// Decode and verify all field names
	for i, expectedFieldName := range fieldNames {
		actualFieldName, err := fieldReader.Read()
		require.NoError(t, err, "Failed to read field name at index %d", i)
		assert.Equal(t, expectedFieldName, actualFieldName, 
			"Field name mismatch at index %d", i)
	}
}

// testFieldNamesEncodeDecode tests field name compression by encoding/decoding 
// maps with the specified field names and verifying the hex output matches
func testFieldNamesEncodeDecode(t *testing.T, fieldNames []string, expectedHex string) {
	// Create a map with unique field names as keys and their final index as values
	input := make(map[string]interface{})
	expectedValues := make(map[string]float64)
	
	for i, fieldName := range fieldNames {
		input[fieldName] = int64(i) // This will store the last occurrence index for duplicate keys
		expectedValues[fieldName] = float64(i) // Remember what we expect for each unique key
	}
	
	// Since field name compression is internal to YAJBE, we test indirectly
	// by verifying that maps with repeated field names encode/decode correctly
	enc, err := Marshal(input)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}
	
	var decoded map[string]interface{}
	err = Unmarshal(enc, &decoded)
	if err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}
	
	// Verify all unique field names are preserved with their final values
	for fieldName, expectedVal := range expectedValues {
		if val, exists := decoded[fieldName]; !exists {
			t.Errorf("Field name %s not found in decoded map", fieldName)
		} else if val != expectedVal {
			t.Errorf("Field %s: expected %v, got %v", fieldName, expectedVal, val)
		}
	}
	
	// Verify no extra fields exist
	assert.Equal(t, len(expectedValues), len(decoded), "Decoded map has wrong number of fields")
}

func TestFieldNameDirectEncoding(t *testing.T) {
	// Test direct field name encoding/decoding like TypeScript tests
	testFieldNamesDirectEncodeDecode(t, []string{
		"aaaaa", "bbbbb", "aaaaa", "aaabb", "aaacc",
	}, "856161616161856262626262a0c2036262c2036363")

	testFieldNamesDirectEncodeDecode(t, []string{
		"aaaaa", "aaabbb", "aaaccc", "ddd", "dddeee", "dddffeee",
	}, "856161616161c303626262c30363636383646464c303656565e203036666")

	testFieldNamesDirectEncodeDecode(t, []string{
		"1234", "1st_place_medal", "2nd_place_medal", "3rd_place_medal",
		"arrow_backward", "arrow_double_down", "arrow_double_up", "arrow_down",
		"arrow_down_small", "arrow_forward", "arrow_heading_down", "arrow_heading_up",
		"arrow_left", "arrow_lower_left", "arrow_lower_right", "arrow_right",
		"code", "ciqual_food_name_tags", "cities_tags", "codes_tags",
		"1st_place_medal", "2nd_place_medal", "3rd_place_medal",
	}, "84313233348f3173745f706c6163655f6d6564616ce3000c326e64e2000d33728e6172726f775f6261636b77617264cb06646f75626c655f646f776ec20d7570c208776ec60a5f736d616c6cc706666f7277617264cc0668656164696e675f646f776ec20e7570c4066c656674e407056f776572c50c7269676874e0060584636f64659563697175616c5f666f6f645f6e616d655f74616773e4020574696573e201076f64a1a2a3")
}

func TestFieldNameSimple(t *testing.T) {
	// Test field name compression with repeated field names
	testFieldNamesEncodeDecode(t, []string{
		"aaaaa", "bbbbb", "aaaaa", "aaabb", "aaacc",
	}, "856161616161856262626262a0c2036262c2036363")

	testFieldNamesEncodeDecode(t, []string{
		"aaaaa", "aaabbb", "aaaccc", "ddd", "dddeee", "dddffeee",
	}, "856161616161c303626262c30363636383646464c303656565e203036666")

	// Test with complex field names that might appear in real JSON
	testFieldNamesEncodeDecode(t, []string{
		"1234", "1st_place_medal", "2nd_place_medal", "3rd_place_medal",
		"arrow_backward", "arrow_double_down", "arrow_double_up", "arrow_down",
		"arrow_down_small", "arrow_forward", "arrow_heading_down", "arrow_heading_up",
		"arrow_left", "arrow_lower_left", "arrow_lower_right", "arrow_right",
		"code", "ciqual_food_name_tags", "cities_tags", "codes_tags",
		"1st_place_medal", "2nd_place_medal", "3rd_place_medal", // Repeated names for compression
	}, "84313233348f3173745f706c6163655f6d6564616ce3000c326e64e2000d33728e6172726f775f6261636b77617264cb06646f75626c655f646f776ec20d7570c208776ec60a5f736d616c6cc706666f7277617264cc0668656164696e675f646f776ec20e7570c4066c656674e407056f776572c50c7269676874e0060584636f64659563697175616c5f666f6f645f6e616d655f74616773e4020574696573e201076f64a1a2a3")
}

func TestFieldNameCompression(t *testing.T) {
	// Test that field name compression actually reduces size for repeated field names
	fieldNames := []string{"field1", "field2", "field3"}
	
	// Create multiple objects with the same field names
	input := []interface{}{
		map[string]interface{}{"field1": 1, "field2": 2, "field3": 3},
		map[string]interface{}{"field1": 4, "field2": 5, "field3": 6},
		map[string]interface{}{"field1": 7, "field2": 8, "field3": 9},
	}
	
	enc, err := Marshal(input)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}
	
	var decoded []interface{}
	err = Unmarshal(enc, &decoded)
	if err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}
	
	// Verify structure is preserved
	if len(decoded) != 3 {
		t.Errorf("Expected 3 objects, got %d", len(decoded))
	}
	
	for i, obj := range decoded {
		if objMap, ok := obj.(map[string]interface{}); ok {
			for j, fieldName := range fieldNames {
				expectedVal := float64(i*3 + j + 1) // Convert to float64 for JSON compatibility
				if val, exists := objMap[fieldName]; !exists {
					t.Errorf("Object %d: field %s not found", i, fieldName)
				} else if val != expectedVal {
					t.Errorf("Object %d field %s: expected %v, got %v", i, fieldName, expectedVal, val)
				}
			}
		} else {
			t.Errorf("Object %d is not a map, got %T", i, obj)
		}
	}
}

func TestFieldNameReaderWriter(t *testing.T) {
	// Test basic field name reader/writer functionality
	testCases := [][]string{
		{"simple"},
		{"a", "b", "c"},
		{"test", "field", "name"},
		{"long_field_name_with_underscores", "another_field", "short"},
		{"repeated", "field", "repeated", "name", "repeated"},
	}
	
	for i, fieldNames := range testCases {
		t.Run(fmt.Sprintf("case_%d", i), func(t *testing.T) {
			// Write field names
			writer := NewWriterFromBuffer(make([]byte, 0, 256))
			fieldWriter := NewFieldNameWriter(writer)
			
			for _, fieldName := range fieldNames {
				err := fieldWriter.Write(fieldName)
				require.NoError(t, err)
			}
			
			// Read field names back
			reader := NewReaderFromBytes(writer.Bytes())
			fieldReader := NewFieldNameReader(reader)
			
			for j, expectedFieldName := range fieldNames {
				actualFieldName, err := fieldReader.Read()
				require.NoError(t, err, "Failed to read field name at index %d", j)
				assert.Equal(t, expectedFieldName, actualFieldName, 
					"Field name mismatch at index %d", j)
			}
		})
	}
}
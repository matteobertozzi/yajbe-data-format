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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// assertEncode tests that input encodes to the expected hex string
func assertEncode(t *testing.T, input interface{}, expectedHex string) {
	enc, err := Marshal(input)
	require.NoError(t, err)
	actual := hex.EncodeToString(enc)
	assert.Equal(t, expectedHex, actual, "Encoding mismatch for input %v", input)
}

// assertDecode tests that the hex string decodes to the expected input
func assertDecode(t *testing.T, expectedHex string, expected interface{}) {
	data, err := hex.DecodeString(expectedHex)
	require.NoError(t, err)
	
	var decoded interface{}
	err = Unmarshal(data, &decoded)
	require.NoError(t, err)
	
	
	// For JSON compatibility, convert expected to what it should become after decoding
	expectedDecoded := convertToExpectedDecoded(expected)
	assert.Equal(t, expectedDecoded, decoded, "Decoding mismatch for hex %s", expectedHex)
}

// assertEncodeDecode tests both encoding and decoding
func assertEncodeDecode(t *testing.T, input interface{}, expectedHex string) {
	// Test encoding
	enc, err := Marshal(input)
	require.NoError(t, err)
	actual := hex.EncodeToString(enc)
	assert.Equal(t, expectedHex, actual, "Encoding mismatch for input %v", input)
	
	// Test decoding - handle JSON compatibility where integers decode as float64
	var decoded interface{}
	err = Unmarshal(enc, &decoded)
	require.NoError(t, err)
	
	// For JSON compatibility, handle type conversions when unmarshaling to interface{}
	expectedDecoded := convertToExpectedDecoded(input)
	
	assert.Equal(t, expectedDecoded, decoded, "Round-trip mismatch for input %v", input)
}

// convertToExpectedDecoded converts input values to what we expect them to become
// after decoding to interface{} for JSON compatibility
func convertToExpectedDecoded(input interface{}) interface{} {
	switch v := input.(type) {
	case int64:
		return float64(v)
	case float32:
		return float64(v)
	case []bool:
		// []bool becomes []interface{} with bool elements
		interfaceSlice := make([]interface{}, len(v))
		for i, val := range v {
			interfaceSlice[i] = val
		}
		return interfaceSlice
	case []interface{}:
		// Convert each element in the slice
		interfaceSlice := make([]interface{}, len(v))
		for i, val := range v {
			interfaceSlice[i] = convertToExpectedDecoded(val)
		}
		return interfaceSlice
	case map[string]interface{}:
		// Convert each value in the map
		interfaceMap := make(map[string]interface{})
		for k, val := range v {
			interfaceMap[k] = convertToExpectedDecoded(val)
		}
		return interfaceMap
	default:
		return input
	}
}
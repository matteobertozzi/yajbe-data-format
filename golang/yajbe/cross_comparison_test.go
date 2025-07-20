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

// TestCrossCompatibilityBool tests that bool encoding matches Dart reference
func TestCrossCompatibilityBool(t *testing.T) {
	// Test cases from Dart implementation
	testCases := []struct {
		input    bool
		expected string
	}{
		{false, "02"},
		{true, "03"},
	}

	for _, tc := range testCases {
		enc, err := Marshal(tc.input)
		require.NoError(t, err)
		actual := hex.EncodeToString(enc)
		assert.Equal(t, tc.expected, actual, "Bool encoding mismatch for %v", tc.input)
	}
}

// TestCrossCompatibilityInt tests that int encoding matches Dart reference
func TestCrossCompatibilityInt(t *testing.T) {
	// Test cases from Dart implementation
	testCases := []struct {
		input    int64
		expected string
	}{
		// Positive ints
		{1, "40"},
		{7, "46"},
		{24, "57"},
		{25, "5800"},
		{127, "5866"},
		{128, "5867"},
		{0xff, "58e6"},
		{0xffff, "59e6ff"},
		{0xffffff, "5ae6ffff"},
		{0xffffffff, "5be6ffffff"},
		
		// Negative ints
		{0, "60"},
		{-1, "61"},
		{-7, "67"},
		{-23, "77"},
		{-24, "7800"},
		{-25, "7801"},
		{-0xff, "78e7"},
		{-0xffff, "79e7ff"},
		{-0xffffff, "7ae7ffff"},
		{-0xffffffff, "7be7ffffff"},
	}

	for _, tc := range testCases {
		enc, err := Marshal(tc.input)
		require.NoError(t, err)
		actual := hex.EncodeToString(enc)
		assert.Equal(t, tc.expected, actual, "Int encoding mismatch for %v", tc.input)
	}
}

// TestCrossCompatibilityString tests that string encoding matches Dart reference
func TestCrossCompatibilityString(t *testing.T) {
	// Test cases from Dart implementation
	testCases := []struct {
		input    string
		expected string
	}{
		{"", "c0"},
		{"a", "c161"},
		{"abc", "c3616263"},
	}

	for _, tc := range testCases {
		enc, err := Marshal(tc.input)
		require.NoError(t, err)
		actual := hex.EncodeToString(enc)
		assert.Equal(t, tc.expected, actual, "String encoding mismatch for %v", tc.input)
	}
}

// TestCrossCompatibilityFloat tests that float encoding matches Dart reference
func TestCrossCompatibilityFloat(t *testing.T) {
	// Test cases from Dart implementation (decode-only, since exact encoding may vary)
	testCases := []struct {
		hexInput string
		expected float64
	}{
		{"0500000000", 0.0},
		{"050000803f", 1.0},
		{"05cdcc8c3f", 1.1},
		{"050a1101c2", -32.26664},
		{"060000000000000080", -0.0},
		{"0600000000000010c0", -4.0},
		{"060000000000fcef40", 65504.0},
		{"0600000000006af840", 100000.0},
	}

	for _, tc := range testCases {
		data, err := hex.DecodeString(tc.hexInput)
		require.NoError(t, err)
		
		var decoded float64
		err = Unmarshal(data, &decoded)
		require.NoError(t, err)
		
		// Use tolerance for float comparison
		delta := 0.000001
		assert.InDelta(t, tc.expected, decoded, delta, "Float decoding mismatch for hex %s", tc.hexInput)
	}
}
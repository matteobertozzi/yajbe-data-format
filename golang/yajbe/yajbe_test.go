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
	"encoding/hex"
	"encoding/json"
	"math/rand"
	"strings"
	"testing"

	"github.com/fxamacker/cbor/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// BaseYajbeTest provides common utilities for YAJBE testing
type BaseYajbeTest struct {
	t      *testing.T
	random *rand.Rand
}

// NewBaseYajbeTest creates a new base test instance
func NewBaseYajbeTest(t *testing.T) *BaseYajbeTest {
	return &BaseYajbeTest{
		t:      t,
		random: rand.New(rand.NewSource(12345)), // Fixed seed for reproducible tests
	}
}

// assertHexEquals compares expected hex string with actual bytes
func (bt *BaseYajbeTest) assertHexEquals(expected string, actual []byte) {
	expected = strings.ReplaceAll(expected, " ", "")
	expected = strings.ToLower(expected)
	actualHex := hex.EncodeToString(actual)
	assert.Equal(bt.t, expected, actualHex, "Hex encoding mismatch")
}

// assertEncodeDecode tests encoding and decoding with expected hex output
func (bt *BaseYajbeTest) assertEncodeDecode(input any, expectedHex string) {
	// Test encoding
	data, err := Marshal(input)
	require.NoError(bt.t, err, "Marshal failed")
	bt.assertHexEquals(expectedHex, data)

	// Test round-trip for typed unmarshaling
	bt.assertRoundTrip(input, data)
}

// assertRoundTrip tests that encoding and decoding preserves the value
func (bt *BaseYajbeTest) assertRoundTrip(original any, encoded []byte) {
	switch v := original.(type) {
	case int:
		var result int
		err := Unmarshal(encoded, &result)
		require.NoError(bt.t, err)
		assert.Equal(bt.t, v, result)
	case int64:
		var result int64
		err := Unmarshal(encoded, &result)
		require.NoError(bt.t, err)
		assert.Equal(bt.t, v, result)
	case float32:
		var result float32
		err := Unmarshal(encoded, &result)
		require.NoError(bt.t, err)
		assert.InDelta(bt.t, v, result, 0.00001)
	case float64:
		var result float64
		err := Unmarshal(encoded, &result)
		require.NoError(bt.t, err)
		assert.InDelta(bt.t, v, result, 0.00001)
	case string:
		var result string
		err := Unmarshal(encoded, &result)
		require.NoError(bt.t, err)
		assert.Equal(bt.t, v, result)
	case []byte:
		var result []byte
		err := Unmarshal(encoded, &result)
		require.NoError(bt.t, err)
		assert.Equal(bt.t, v, result)
	case bool:
		var result bool
		err := Unmarshal(encoded, &result)
		require.NoError(bt.t, err)
		assert.Equal(bt.t, v, result)
	case nil:
		var result any
		err := Unmarshal(encoded, &result)
		require.NoError(bt.t, err)
		assert.Nil(bt.t, result)
	case []int:
		var result []int
		err := Unmarshal(encoded, &result)
		require.NoError(bt.t, err)
		assert.Equal(bt.t, v, result)
	case []string:
		var result []string
		err := Unmarshal(encoded, &result)
		require.NoError(bt.t, err)
		assert.Equal(bt.t, v, result)
	case []float64:
		var result []float64
		err := Unmarshal(encoded, &result)
		require.NoError(bt.t, err)
		assert.Equal(bt.t, v, result)
	case []bool:
		var result []bool
		err := Unmarshal(encoded, &result)
		require.NoError(bt.t, err)
		assert.Equal(bt.t, v, result)
	case []int64:
		var result []int64
		err := Unmarshal(encoded, &result)
		require.NoError(bt.t, err)
		assert.Equal(bt.t, v, result)
	case []float32:
		var result []float32
		err := Unmarshal(encoded, &result)
		require.NoError(bt.t, err)
		assert.Equal(bt.t, v, result)
	default:
		// Generic interface unmarshaling
		var result any
		err := Unmarshal(encoded, &result)
		require.NoError(bt.t, err)
		assert.Equal(bt.t, original, result)
	}
}

// assertDecode tests only decoding from hex input
func (bt *BaseYajbeTest) assertDecode(hexStr string, expected any) {
	data, err := hex.DecodeString(strings.ReplaceAll(hexStr, " ", ""))
	require.NoError(bt.t, err, "Hex decode failed")

	var result any
	err = Unmarshal(data, &result)
	require.NoError(bt.t, err, "Unmarshal failed")
	assert.Equal(bt.t, expected, result)
}

// assertArrayEncodeDecode tests array encoding and decoding
func (bt *BaseYajbeTest) assertArrayEncodeDecode(input any, expectedHex string) {
	data, err := Marshal(input)
	require.NoError(bt.t, err, "Array marshal failed")
	bt.assertHexEquals(expectedHex, data)

	// Test round-trip
	bt.assertRoundTrip(input, data)
}

// Random data generators

// randIntBlock generates random integers with balanced byte widths
func (bt *BaseYajbeTest) randIntBlock(length int) []int {
	result := make([]int, length)
	for i := range result {
		switch bt.random.Intn(4) {
		case 0: // Small values (-23 to 24)
			result[i] = bt.random.Intn(48) - 23
		case 1: // 1-byte values
			result[i] = bt.random.Intn(256) + 25
		case 2: // 2-byte values
			result[i] = bt.random.Intn(65536) + 256 + 25
		case 3: // Large values
			result[i] = bt.random.Int()
		}
	}
	return result
}

// randLongBlock generates random int64s with balanced byte widths
func (bt *BaseYajbeTest) randLongBlock(length int) []int64 {
	result := make([]int64, length)
	for i := range result {
		switch bt.random.Intn(5) {
		case 0: // Small values
			result[i] = int64(bt.random.Intn(48) - 23)
		case 1: // 1-byte values
			result[i] = int64(bt.random.Intn(256) + 25)
		case 2: // 2-byte values
			result[i] = int64(bt.random.Intn(65536) + 256 + 25)
		case 3: // 4-byte values
			result[i] = int64(bt.random.Int31())
		case 4: // 8-byte values
			result[i] = bt.random.Int63()
		}
	}
	return result
}

// randText generates random text of specified length
func (bt *BaseYajbeTest) randText(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789 "
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[bt.random.Intn(len(charset))]
	}
	return string(result)
}

// generateFieldName generates random field names for testing
func (bt *BaseYajbeTest) generateFieldName(minLen, maxLen int) string {
	length := minLen + bt.random.Intn(maxLen-minLen+1)
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ_"
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[bt.random.Intn(len(charset))]
	}
	return string(result)
}

// Test utilities for dataset comparison
func (bt *BaseYajbeTest) compareWithOthers(data any) (yajbeSize, jsonSize, cborSize int) {
	yajbeData, err := Marshal(data)
	require.NoError(bt.t, err)

	jsonData, err := json.Marshal(data)
	require.NoError(bt.t, err)

	cborData, err := cbor.Marshal(data)
	require.NoError(bt.t, err)

	yajbeSize = len(yajbeData)
	jsonSize = len(jsonData)
	cborSize = len(cborData)
	return
}

// Basic smoke tests
func TestBasicTypes(t *testing.T) {
	bt := NewBaseYajbeTest(t)

	t.Run("Null", func(t *testing.T) {
		bt.assertEncodeDecode(nil, "00")
	})

	t.Run("Boolean", func(t *testing.T) {
		bt.assertEncodeDecode(true, "03")
		bt.assertEncodeDecode(false, "02")
	})

	t.Run("SimpleInts", func(t *testing.T) {
		bt.assertEncodeDecode(0, "60")
		bt.assertEncodeDecode(1, "40")
		bt.assertEncodeDecode(-1, "61")
	})

	t.Run("SimpleStrings", func(t *testing.T) {
		bt.assertEncodeDecode("", "c0")
		bt.assertEncodeDecode("a", "c161")
		bt.assertEncodeDecode("abc", "c3616263")
	})
}

// Test that our optimized implementation maintains correctness
func TestOptimizedCorrectness(t *testing.T) {
	bt := NewBaseYajbeTest(t)

	// Test various data types with random data
	t.Run("RandomInts", func(t *testing.T) {
		for i := 0; i < 100; i++ {
			value := bt.random.Int()
			data, err := Marshal(value)
			require.NoError(t, err)

			var result int
			err = Unmarshal(data, &result)
			require.NoError(t, err)
			assert.Equal(t, value, result)
		}
	})

	t.Run("RandomStrings", func(t *testing.T) {
		for i := 0; i < 100; i++ {
			value := bt.randText(bt.random.Intn(1000))
			data, err := Marshal(value)
			require.NoError(t, err)

			var result string
			err = Unmarshal(data, &result)
			require.NoError(t, err)
			assert.Equal(t, value, result)
		}
	})

	t.Run("ComplexStructures", func(t *testing.T) {
		type TestStruct struct {
			ID    int      `json:"id"`
			Name  string   `json:"name"`
			Value float64  `json:"value"`
			Tags  []string `json:"tags"`
		}

		for i := 0; i < 10; i++ {
			original := TestStruct{
				ID:    bt.random.Int(),
				Name:  bt.randText(20),
				Value: bt.random.Float64() * 1000,
				Tags:  []string{bt.randText(10), bt.randText(15)},
			}

			data, err := Marshal(original)
			require.NoError(t, err)

			var result TestStruct
			err = Unmarshal(data, &result)
			require.NoError(t, err)

			assert.Equal(t, original.ID, result.ID)
			assert.Equal(t, original.Name, result.Name)
			assert.InDelta(t, original.Value, result.Value, 0.00001)
			assert.Equal(t, original.Tags, result.Tags)
		}
	})
}

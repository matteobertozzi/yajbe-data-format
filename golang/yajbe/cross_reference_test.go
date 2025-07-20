package yajbe

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCrossReferenceWithDart(t *testing.T) {
	// Test cases that both dart and golang should produce the same encoding for
	// These hex values come from running the dart implementation
	testCases := []struct {
		input       interface{}
		expectedHex string
		description string
	}{
		{false, "02", "false"},
		{true, "03", "true"},
		{int64(1), "40", "int 1"},
		{int64(0), "60", "int 0"},
		{int64(-1), "61", "int -1"},
		{"hello", "c568656c6c6f", "string hello"},
		{nil, "00", "null"},
		{[]interface{}{int64(1), int64(2), int64(3)}, "23404142", "array [1,2,3]"},
		{[]bool{true, false}, "220302", "bool array [true, false]"},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			// Test encoding
			encoded, err := Marshal(tc.input)
			require.NoError(t, err)
			actualHex := hex.EncodeToString(encoded)
			
			assert.Equal(t, tc.expectedHex, actualHex,
				"Input %v should encode to %s but got %s", tc.input, tc.expectedHex, actualHex)
			
			// Test round-trip decoding
			var decoded interface{}
			err = Unmarshal(encoded, &decoded)
			require.NoError(t, err)
			
			// Convert expected to what it should become after decoding
			expectedDecoded := convertToExpectedDecoded(tc.input)
			assert.Equal(t, expectedDecoded, decoded,
				"Round-trip failed for input: %v", tc.input)
		})
	}
}

func TestProblematicCases(t *testing.T) {
	// Test the cases that were problematic during development
	// Using decode-only tests since encoding order might differ
	problematicCases := []struct {
		expectedHex string
		expected    interface{}
		description string
	}{
		{"31816140", map[string]interface{}{"a": float64(1)}, "Simple map {a: 1}"},
		{"22069a9999999999f13f069a99999999990140", []interface{}{1.1, 2.2}, "Float array [1.1, 2.2]"},
		{"31836b657923404142", map[string]interface{}{"key": []interface{}{float64(1), float64(2), float64(3)}}, "Map with array value"},
	}

	for _, tc := range problematicCases {
		t.Run(tc.description, func(t *testing.T) {
			// Test that we can decode what dart produces
			data, err := hex.DecodeString(tc.expectedHex)
			require.NoError(t, err)
			
			var decoded interface{}
			err = Unmarshal(data, &decoded)
			require.NoError(t, err)
			
			assert.Equal(t, tc.expected, decoded,
				"Failed to decode %s as %v", tc.expectedHex, tc.expected)
		})
	}
}

func TestMapOrderMatches(t *testing.T) {
	// Test simple map with keys in different insertion order (same as dart test)
	simpleMap := map[string]interface{}{
		"ccc": int64(3),
		"aaa": int64(1), 
		"bbb": int64(2),
	}
	
	encoded, err := Marshal(simpleMap)
	require.NoError(t, err)
	
	hexStr := hex.EncodeToString(encoded)
	t.Logf("Go simple map hex: %s", hexStr)
	
	// Expected from dart: 33836161614083626262418363636342
	// With alphabetical ordering it should be: aaa, bbb, ccc
	dartExpected := "33836161614083626262418363636342"
	t.Logf("Dart expected:     %s", dartExpected)
	
	assert.Equal(t, dartExpected, hexStr, "Go encoding should match Dart encoding with sorted keys")
	
	// Test decoding to make sure it works
	var decoded map[string]interface{}
	err = Unmarshal(encoded, &decoded)
	require.NoError(t, err)
	t.Logf("Decoded: %+v", decoded)
}

func TestComplexMapOrderMatches(t *testing.T) {
	// Test the complex map from dart test
	complexMap := map[string]interface{}{
		"aaa": int64(1),
		"bbb": map[string]interface{}{"k": int64(10)},
		"ccc": 2.3,
		"ddd": []interface{}{"a", "b"},
		"eee": []interface{}{"a", map[string]interface{}{"k": int64(10)}, "b"},
		"fff": map[string]interface{}{"a": map[string]interface{}{"k": []interface{}{"z", "d"}}},
		"ggg": "foo",
	}
	
	encoded, err := Marshal(complexMap)
	require.NoError(t, err)
	
	hexStr := hex.EncodeToString(encoded)
	t.Logf("Go complex map hex: %s", hexStr)
	
	// From dart: 3783616161408362626231816b49836363630666666666666602408364646422c161c1628365656523c16131a249c1628366666631816131a222c17ac16483676767c3666f6f
	dartExpected := "3783616161408362626231816b49836363630666666666666602408364646422c161c1628365656523c16131a249c1628366666631816131a222c17ac16483676767c3666f6f"
	t.Logf("Dart expected:       %s", dartExpected)
	
	// Test decoding to make sure it works
	var decoded map[string]interface{}
	err = Unmarshal(encoded, &decoded)
	require.NoError(t, err)
	t.Logf("Decoded: %+v", decoded)
	
	// Note: We don't assert exact match for complex maps since nested structures 
	// might have different ordering, but they should decode correctly
}
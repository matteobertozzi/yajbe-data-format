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
	"fmt"
	"strings"
	"testing"
	"unicode/utf8"
)

func TestYajbeStrings(t *testing.T) {
	bt := NewBaseYajbeTest(t)

	t.Run("Simple", func(t *testing.T) {
		bt.assertEncodeDecode("", "c0")
		bt.assertEncodeDecode("a", "c161")
		bt.assertEncodeDecode("abc", "c3616263")
		bt.assertEncodeDecode(strings.Repeat("x", 59), "fb"+strings.Repeat("78", 59))
		bt.assertEncodeDecode(strings.Repeat("y", 60), "fc01"+strings.Repeat("79", 60))
		bt.assertEncodeDecode(strings.Repeat("y", 127), "fc44"+strings.Repeat("79", 127))
		bt.assertEncodeDecode(strings.Repeat("y", 255), "fcc4"+strings.Repeat("79", 255))
		bt.assertEncodeDecode(strings.Repeat("z", 0x100), "fcc5"+strings.Repeat("7a", 256))
		bt.assertEncodeDecode(strings.Repeat("z", 314), "fcff"+strings.Repeat("7a", 314))
		bt.assertEncodeDecode(strings.Repeat("z", 315), "fd0001"+strings.Repeat("7a", 315))
		bt.assertEncodeDecode(strings.Repeat("z", 0xffff), "fdc4ff"+strings.Repeat("7a", 0xffff))
	})

	t.Run("SmallStringLength", func(t *testing.T) {
		// Test strings with length 0-59 (inline length encoding)
		for i := 0; i < 60; i++ {
			input := strings.Repeat("x", i)
			data, err := Marshal(input)
			if err != nil {
				t.Fatalf("Failed to marshal string of length %d: %v", i, err)
			}

			expectedSize := 1 + i // 1 byte header + string bytes
			if len(data) != expectedSize {
				t.Errorf("String length %d: expected %d bytes, got %d", i, expectedSize, len(data))
			}

			var result string
			err = Unmarshal(data, &result)
			if err != nil {
				t.Fatalf("Failed to unmarshal string of length %d: %v", i, err)
			}

			if result != input {
				t.Errorf("String length %d round-trip failed", i)
			}
		}

		// Test strings with length 60-314 (1-byte length extension)
		for i := 60; i <= 314; i++ {
			input := strings.Repeat("x", i)
			data, err := Marshal(input)
			if err != nil {
				t.Fatalf("Failed to marshal string of length %d: %v", i, err)
			}

			expectedSize := 2 + i // 2 byte header + string bytes
			if len(data) != expectedSize {
				t.Errorf("String length %d: expected %d bytes, got %d", i, expectedSize, len(data))
			}

			var result string
			err = Unmarshal(data, &result)
			if err != nil {
				t.Fatalf("Failed to unmarshal string of length %d: %v", i, err)
			}

			if result != input {
				t.Errorf("String length %d round-trip failed", i)
			}
		}

		// Test strings with length 315-8191 (2-byte length extension)
		testLengths := []int{315, 500, 1000, 2000, 4000, 8191}
		for _, i := range testLengths {
			input := strings.Repeat("x", i)
			data, err := Marshal(input)
			if err != nil {
				t.Fatalf("Failed to marshal string of length %d: %v", i, err)
			}

			expectedSize := 3 + i // 3 byte header + string bytes
			if len(data) != expectedSize {
				t.Errorf("String length %d: expected %d bytes, got %d", i, expectedSize, len(data))
			}

			var result string
			err = Unmarshal(data, &result)
			if err != nil {
				t.Fatalf("Failed to unmarshal string of length %d: %v", i, err)
			}

			if result != input {
				t.Errorf("String length %d round-trip failed", i)
			}
		}
	})

	t.Run("UTF8", func(t *testing.T) {
		// Test various UTF-8 strings
		utf8Tests := []string{
			"Hello, 世界",
			"🚀🎉💫",
			"Ñiño güey",
			"Тест",
			"مرحبا",
			"שלום",
			"こんにちは",
			"🇺🇸🇪🇸🇫🇷🇩🇪🇮🇹🇯🇵🇰🇷🇨🇳",
			"aᄫ",    // Korean Jamo
			"𝓗𝓮𝓵𝓵𝓸", // Mathematical script letters
		}

		for _, testStr := range utf8Tests {
			if !utf8.ValidString(testStr) {
				t.Errorf("Test string is not valid UTF-8: %s", testStr)
				continue
			}

			data, err := Marshal(testStr)
			if err != nil {
				t.Fatalf("Failed to marshal UTF-8 string %s: %v", testStr, err)
			}

			var result string
			err = Unmarshal(data, &result)
			if err != nil {
				t.Fatalf("Failed to unmarshal UTF-8 string %s: %v", testStr, err)
			}

			if result != testStr {
				t.Errorf("UTF-8 string round-trip failed: %s != %s", testStr, result)
			}
		}
	})

	t.Run("SpecialCharacters", func(t *testing.T) {
		// Test strings with special characters
		specialTests := []string{
			"\x00",                           // Null byte
			"\x01\x02\x03",                   // Control characters
			"\n\r\t",                         // Whitespace
			"\"'\\",                          // Quote and backslash
			"\x7F",                           // DEL character
			"\xFF",                           // High byte value
			string([]byte{0x80, 0x81, 0x82}), // Non-UTF-8 bytes
		}

		for _, testStr := range specialTests {
			data, err := Marshal(testStr)
			if err != nil {
				t.Fatalf("Failed to marshal special string %q: %v", testStr, err)
			}

			var result string
			err = Unmarshal(data, &result)
			if err != nil {
				t.Fatalf("Failed to unmarshal special string %q: %v", testStr, err)
			}

			if result != testStr {
				t.Errorf("Special string round-trip failed: %q != %q", testStr, result)
			}
		}
	})

	t.Run("Random", func(t *testing.T) {
		// Test random strings
		for i := 0; i < 100; i++ {
			length := bt.random.Intn(1 << 17) // Up to 128KB
			input := bt.randText(length)

			data, err := Marshal(input)
			if err != nil {
				t.Fatalf("Failed to marshal random string of length %d: %v", length, err)
			}

			var result string
			err = Unmarshal(data, &result)
			if err != nil {
				t.Fatalf("Failed to unmarshal random string of length %d: %v", length, err)
			}

			if result != input {
				t.Errorf("Random string round-trip failed for length %d", length)
			}
		}
	})

	t.Run("Arrays", func(t *testing.T) {
		// Test string arrays
		bt.assertArrayEncodeDecode([]string{}, "20")
		bt.assertArrayEncodeDecode([]string{""}, "21c0")
		bt.assertArrayEncodeDecode([]string{"a"}, "21c161")
		bt.assertArrayEncodeDecode([]string{"hello", "world"}, "22c568656c6c6fc5776f726c64")

		// Test larger string arrays
		largeArray := make([]string, 1000)
		for i := range largeArray {
			largeArray[i] = fmt.Sprintf("string_%d", i)
		}

		data, err := Marshal(largeArray)
		if err != nil {
			t.Fatalf("Failed to marshal large string array: %v", err)
		}

		var result []string
		err = Unmarshal(data, &result)
		if err != nil {
			t.Fatalf("Failed to unmarshal large string array: %v", err)
		}

		if len(result) != len(largeArray) {
			t.Fatalf("Large string array length mismatch: expected %d, got %d", len(largeArray), len(result))
		}

		for i, expected := range largeArray {
			if result[i] != expected {
				t.Errorf("Large string array element %d mismatch: expected %s, got %s", i, expected, result[i])
				break // Don't spam on failure
			}
		}
	})

	t.Run("EdgeCases", func(t *testing.T) {
		// Test edge cases
		edgeTests := []struct {
			name string
			str  string
		}{
			{"Empty", ""},
			{"SingleByte", "a"},
			{"Boundary59", strings.Repeat("x", 59)},
			{"Boundary60", strings.Repeat("x", 60)},
			{"Boundary314", strings.Repeat("x", 314)},
			{"Boundary315", strings.Repeat("x", 315)},
			{"Large", strings.Repeat("x", 10000)},
		}

		for _, test := range edgeTests {
			t.Run(test.name, func(t *testing.T) {
				data, err := Marshal(test.str)
				if err != nil {
					t.Fatalf("Failed to marshal %s: %v", test.name, err)
				}

				var result string
				err = Unmarshal(data, &result)
				if err != nil {
					t.Fatalf("Failed to unmarshal %s: %v", test.name, err)
				}

				if result != test.str {
					t.Errorf("%s round-trip failed", test.name)
				}
			})
		}
	})

	t.Run("EncodingEfficiency", func(t *testing.T) {
		// Test that string encoding is efficient
		tests := []struct {
			str         string
			maxOverhead int // Maximum expected overhead in bytes
		}{
			{"", 1},                           // Empty string should be 1 byte
			{"a", 2},                          // Single char should be 2 bytes
			{strings.Repeat("x", 59), 60},     // 59 chars should be 60 bytes
			{strings.Repeat("x", 60), 62},     // 60 chars should be 62 bytes
			{strings.Repeat("x", 255), 257},   // 255 chars should be 257 bytes
			{strings.Repeat("x", 1000), 1003}, // 1000 chars should be ~1003 bytes
		}

		for _, test := range tests {
			data, err := Marshal(test.str)
			if err != nil {
				t.Fatalf("Failed to marshal string: %v", err)
			}

			if len(data) > test.maxOverhead {
				t.Errorf("String encoding inefficient: %d chars encoded to %d bytes (max expected %d)",
					len(test.str), len(data), test.maxOverhead)
			}
		}
	})

	t.Run("Comparison", func(t *testing.T) {
		// Compare YAJBE string encoding with JSON
		testStrings := []string{
			"",
			"hello",
			"hello world",
			strings.Repeat("test", 100),
			"unicode: 🚀🎉💫",
			bt.randText(1000),
		}

		for _, str := range testStrings {
			yajbeSize, jsonSize, cborSize := bt.compareWithOthers(str)
			t.Logf("String len=%d: YAJBE=%d, JSON=%d (%.2f%%), CBOR=%d (%.2f%%)",
				len(str), yajbeSize, jsonSize, float64(jsonSize)/float64(yajbeSize)*100, cborSize, float64(cborSize)/float64(yajbeSize)*100)
		}
	})
}

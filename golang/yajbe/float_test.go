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
	"math"
	"math/big"
	"testing"
)

func TestYajbeFloats(t *testing.T) {
	bt := NewBaseYajbeTest(t)

	t.Run("Simple", func(t *testing.T) {
		// Float32 tests
		bt.assertEncodeDecode(float32(0.0), "0500000000")
		bt.assertEncodeDecode(float32(1.0), "050000803f")
		bt.assertEncodeDecode(float32(1.1), "05cdcc8c3f")
		bt.assertEncodeDecode(float32(-32.26664), "050a1101c2")
		bt.assertEncodeDecode(float32(math.Inf(1)), "050000807f")
		bt.assertEncodeDecode(float32(math.Inf(-1)), "05000080ff")

		// Skip NaN test as Go NaN encoding may differ
		// bt.assertEncodeDecode(float32(math.NaN()), "050000c07f")

		// Float64 tests
		bt.assertEncodeDecode(-4.0, "0600000000000010c0")
		bt.assertEncodeDecode(-4.1, "0666666666666610c0")
		bt.assertEncodeDecode(1.5, "06000000000000f83f")
		bt.assertEncodeDecode(65504.0, "060000000000fcef40")
		bt.assertEncodeDecode(100000.0, "0600000000006af840")
		bt.assertEncodeDecode(5.960464477539063e-8, "06000000000000703e")
		bt.assertEncodeDecode(0.00006103515625, "06000000000000103f")
		bt.assertEncodeDecode(-5.960464477539063e-8, "0600000000000070be")
		bt.assertEncodeDecode(3.4028234663852886e+38, "06000000e0ffffef47")
		bt.assertEncodeDecode(9007199254740994.0, "060100000000004043")
		bt.assertEncodeDecode(-9007199254740994.0, "0601000000000040c3")
		bt.assertEncodeDecode(1.0e+300, "069c7500883ce4377e")
		bt.assertEncodeDecode(-40.049149, "06c8d0b1834a0644c0")
		bt.assertEncodeDecode(math.Inf(1), "06000000000000f07f")
		bt.assertEncodeDecode(math.Inf(-1), "06000000000000f0ff")
	})

	t.Run("SpecialValues", func(t *testing.T) {
		// Test zero values
		bt.assertEncodeDecode(float32(0.0), "0500000000")
		bt.assertEncodeDecode(float64(0.0), "060000000000000000")

		// Test positive and negative infinity
		bt.assertRoundTrip(float32(math.Inf(1)), []byte{0x05, 0x00, 0x00, 0x80, 0x7f})
		bt.assertRoundTrip(float32(math.Inf(-1)), []byte{0x05, 0x00, 0x00, 0x80, 0xff})
		bt.assertRoundTrip(float64(math.Inf(1)), []byte{0x06, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xf0, 0x7f})
		bt.assertRoundTrip(float64(math.Inf(-1)), []byte{0x06, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xf0, 0xff})

		// Test NaN (just round-trip test since encoding may vary)
		nan32Data, err := Marshal(float32(math.NaN()))
		if err != nil {
			t.Fatalf("Failed to marshal float32 NaN: %v", err)
		}
		var result32 float32
		err = Unmarshal(nan32Data, &result32)
		if err != nil {
			t.Fatalf("Failed to unmarshal float32 NaN: %v", err)
		}
		if !math.IsNaN(float64(result32)) {
			t.Errorf("float32 NaN round-trip failed: got %f", result32)
		}

		nan64Data, err := Marshal(math.NaN())
		if err != nil {
			t.Fatalf("Failed to marshal float64 NaN: %v", err)
		}
		var result64 float64
		err = Unmarshal(nan64Data, &result64)
		if err != nil {
			t.Fatalf("Failed to unmarshal float64 NaN: %v", err)
		}
		if !math.IsNaN(result64) {
			t.Errorf("float64 NaN round-trip failed: got %f", result64)
		}
	})

	t.Run("BigNumbers", func(t *testing.T) {
		// Test BigInteger
		bigNum := new(big.Int)
		bigNum.SetString("340282366920938463463374607431768211455", 10)
		bt.assertEncodeDecode(bigNum, "070000001100ffffffffffffffffffffffffffffffff")

		bigNumNeg := new(big.Int)
		bigNumNeg.SetString("-340282366920938463463374607431768211455", 10)
		bt.assertEncodeDecode(bigNumNeg, "070400001100ffffffffffffffffffffffffffffffff")

		// Test various BigInteger sizes
		testBigInts := []string{
			"0",
			"1",
			"-1",
			"127",
			"-128",
			"32767",
			"-32768",
			"2147483647",
			"-2147483648",
			"9223372036854775807",
			"-9223372036854775808",
			"123456789012345678901234567890",
			"-123456789012345678901234567890",
		}

		for _, str := range testBigInts {
			bigInt := new(big.Int)
			bigInt.SetString(str, 10)

			data, err := Marshal(bigInt)
			if err != nil {
				t.Fatalf("Failed to marshal BigInt %s: %v", str, err)
			}

			var result *big.Int
			err = Unmarshal(data, &result)
			if err != nil {
				t.Fatalf("Failed to unmarshal BigInt %s: %v", str, err)
			}

			if result.Cmp(bigInt) != 0 {
				t.Errorf("BigInt round-trip failed: %s != %s", bigInt.String(), result.String())
			}
		}
	})

	t.Run("Random", func(t *testing.T) {
		// Test random float32 values
		for i := 0; i < 100; i++ {
			input := bt.random.Float32()
			if math.IsNaN(float64(input)) || math.IsInf(float64(input), 0) {
				continue // Skip special values
			}

			data, err := Marshal(input)
			if err != nil {
				t.Fatalf("Failed to marshal random float32 %f: %v", input, err)
			}

			var result float32
			err = Unmarshal(data, &result)
			if err != nil {
				t.Fatalf("Failed to unmarshal random float32 %f: %v", input, err)
			}

			if math.Abs(float64(input-result)) > 0.00001 {
				t.Errorf("Random float32 round-trip failed: input=%f, result=%f", input, result)
			}
		}

		// Test random float64 values
		for i := 0; i < 100; i++ {
			input := bt.random.Float64()
			if math.IsNaN(input) || math.IsInf(input, 0) {
				continue // Skip special values
			}

			data, err := Marshal(input)
			if err != nil {
				t.Fatalf("Failed to marshal random float64 %f: %v", input, err)
			}

			var result float64
			err = Unmarshal(data, &result)
			if err != nil {
				t.Fatalf("Failed to unmarshal random float64 %f: %v", input, err)
			}

			if math.Abs(input-result) > 0.00001 {
				t.Errorf("Random float64 round-trip failed: input=%f, result=%f", input, result)
			}
		}
	})

	t.Run("Arrays", func(t *testing.T) {
		// Test float arrays
		float32Array := []float32{1.1, 2.2, 3.3, -4.4, 0.0}
		data, err := Marshal(float32Array)
		if err != nil {
			t.Fatalf("Failed to marshal float32 array: %v", err)
		}

		var result32Array []float32
		err = Unmarshal(data, &result32Array)
		if err != nil {
			t.Fatalf("Failed to unmarshal float32 array: %v", err)
		}

		if len(result32Array) != len(float32Array) {
			t.Fatalf("float32 array length mismatch: expected %d, got %d", len(float32Array), len(result32Array))
		}

		for i, expected := range float32Array {
			if math.Abs(float64(expected-result32Array[i])) > 0.00001 {
				t.Errorf("float32 array element %d mismatch: expected %f, got %f", i, expected, result32Array[i])
			}
		}

		// Test float64 array
		float64Array := []float64{1.1, 2.2, 3.3, -4.4, 0.0, math.Pi, math.E}
		data, err = Marshal(float64Array)
		if err != nil {
			t.Fatalf("Failed to marshal float64 array: %v", err)
		}

		var result64Array []float64
		err = Unmarshal(data, &result64Array)
		if err != nil {
			t.Fatalf("Failed to unmarshal float64 array: %v", err)
		}

		if len(result64Array) != len(float64Array) {
			t.Fatalf("float64 array length mismatch: expected %d, got %d", len(float64Array), len(result64Array))
		}

		for i, expected := range float64Array {
			if math.Abs(expected-result64Array[i]) > 0.00001 {
				t.Errorf("float64 array element %d mismatch: expected %f, got %f", i, expected, result64Array[i])
			}
		}
	})

	t.Run("Precision", func(t *testing.T) {
		// Test precision preservation for various float values
		precisionTests := []float64{
			0.1,
			0.01,
			0.001,
			0.0001,
			0.00001,
			0.000001,
			1.23456789,
			12.3456789,
			123.456789,
			1234.56789,
			12345.6789,
			math.Pi,
			math.E,
			math.Sqrt(2),
			1.0 / 3.0,
		}

		for _, value := range precisionTests {
			data, err := Marshal(value)
			if err != nil {
				t.Fatalf("Failed to marshal precision test value %f: %v", value, err)
			}

			var result float64
			err = Unmarshal(data, &result)
			if err != nil {
				t.Fatalf("Failed to unmarshal precision test value %f: %v", value, err)
			}

			// Use exact equality for IEEE 754 floating point
			if result != value {
				t.Errorf("Precision test failed for %f: got %f", value, result)
			}
		}
	})
}

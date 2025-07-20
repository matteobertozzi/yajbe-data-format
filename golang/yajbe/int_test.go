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
	"math"
	"testing"
)

func TestYajbeInts(t *testing.T) {
	bt := NewBaseYajbeTest(t)

	t.Run("Simple", func(t *testing.T) {
		// Positive ints
		bt.assertEncodeDecode(1, "40")
		bt.assertEncodeDecode(7, "46")
		bt.assertEncodeDecode(24, "57")
		bt.assertEncodeDecode(25, "5800")
		bt.assertEncodeDecode(127, "5866")
		bt.assertEncodeDecode(0xff, "58e6")
		bt.assertEncodeDecode(0xffff, "59e6ff")
		bt.assertEncodeDecode(0xffffff, "5ae6ffff")
		bt.assertEncodeDecode(int64(0xffffffff), "5be6ffffff")
		bt.assertEncodeDecode(int64(0xffffffffff), "5ce6ffffffff")
		bt.assertEncodeDecode(int64(0xffffffffffff), "5de6ffffffffff")
		bt.assertEncodeDecode(int64(0x1fffffffffffff), "5ee6ffffffffff1f")
		bt.assertEncodeDecode(int64(0xffffffffffffff), "5ee6ffffffffffff")

		bt.assertEncodeDecode(100, "584b")
		bt.assertEncodeDecode(1000, "59cf03")
		bt.assertEncodeDecode(int64(1000000), "5a27420f")
		bt.assertEncodeDecode(int64(1000000000000), "5ce70fa5d4e8")
		bt.assertEncodeDecode(int64(100000000000000), "5de73f7a10f35a")

		// Negative ints
		bt.assertEncodeDecode(0, "60")
		bt.assertEncodeDecode(-1, "61")
		bt.assertEncodeDecode(-7, "67")
		bt.assertEncodeDecode(-23, "77")
		bt.assertEncodeDecode(-24, "7800")
		bt.assertEncodeDecode(-25, "7801")
		bt.assertEncodeDecode(-0xff, "78e7")
		bt.assertEncodeDecode(-0xffff, "79e7ff")
		bt.assertEncodeDecode(-0xffffff, "7ae7ffff")
		bt.assertEncodeDecode(int64(-0xffffffff), "7be7ffffff")
		bt.assertEncodeDecode(int64(-0xffffffffff), "7ce7ffffffff")
		bt.assertEncodeDecode(int64(-0xffffffffffff), "7de7ffffffffff")

		bt.assertEncodeDecode(-100, "784c")
		bt.assertEncodeDecode(-1000, "79d003")
		bt.assertEncodeDecode(int64(-1000000), "7a28420f")
		bt.assertEncodeDecode(int64(-1000000000000), "7ce80fa5d4e8")
		bt.assertEncodeDecode(int64(-100000000000000), "7de83f7a10f35a")
	})

	t.Run("EdgeCases", func(t *testing.T) {
		// Test boundary values
		bt.assertEncodeDecode(math.MaxInt8, "5866")
		bt.assertEncodeDecode(math.MinInt8, "7868")
		bt.assertEncodeDecode(math.MaxInt16, "59e67f")
		bt.assertEncodeDecode(math.MinInt16, "79e87f")
		bt.assertEncodeDecode(math.MaxInt32, "5be6ffff7f")
		bt.assertEncodeDecode(math.MinInt32, "7be8ffff7f")

		// Test values around encoding boundaries
		bt.assertEncodeDecode(23, "56")
		bt.assertEncodeDecode(24, "57")
		bt.assertEncodeDecode(25, "5800")
		bt.assertEncodeDecode(26, "5801")

		bt.assertEncodeDecode(-22, "76")
		bt.assertEncodeDecode(-23, "77")
		bt.assertEncodeDecode(-24, "7800")
		bt.assertEncodeDecode(-25, "7801")
	})

	t.Run("SmallIntegers", func(t *testing.T) {
		// Test all small integers that fit in the header byte
		for i := -23; i <= 24; i++ {
			data, err := Marshal(i)
			if err != nil {
				t.Fatalf("Failed to marshal %d: %v", i, err)
			}

			var result int
			err = Unmarshal(data, &result)
			if err != nil {
				t.Fatalf("Failed to unmarshal %d: %v", i, err)
			}

			if result != i {
				t.Errorf("Round-trip failed for %d: got %d", i, result)
			}
		}
	})

	t.Run("Random", func(t *testing.T) {
		// Test random integers
		for i := 0; i < 1000; i++ {
			input := bt.random.Int()
			data, err := Marshal(input)
			if err != nil {
				t.Fatalf("Failed to marshal random int %d: %v", input, err)
			}

			var result int
			err = Unmarshal(data, &result)
			if err != nil {
				t.Fatalf("Failed to unmarshal random int %d: %v", input, err)
			}

			if result != input {
				t.Errorf("Random int round-trip failed: input=%d, result=%d", input, result)
			}
		}

		// Test random int64s
		for i := 0; i < 1000; i++ {
			input := bt.random.Int63()
			data, err := Marshal(input)
			if err != nil {
				t.Fatalf("Failed to marshal random int64 %d: %v", input, err)
			}

			var result int64
			err = Unmarshal(data, &result)
			if err != nil {
				t.Fatalf("Failed to unmarshal random int64 %d: %v", input, err)
			}

			if result != input {
				t.Errorf("Random int64 round-trip failed: input=%d, result=%d", input, result)
			}
		}
	})

	t.Run("Arrays", func(t *testing.T) {
		// Test integer arrays
		bt.assertArrayEncodeDecode([]int{}, "20")
		bt.assertArrayEncodeDecode([]int{1}, "2140")
		bt.assertArrayEncodeDecode([]int{2, 2}, "224141")
		bt.assertArrayEncodeDecode([]int{0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, "2a"+
			"60606060606060606060")

		// Test arrays with different value ranges
		smallInts := bt.randIntBlock(10)
		data, err := Marshal(smallInts)
		if err != nil {
			t.Fatalf("Failed to marshal int array: %v", err)
		}

		var result []int
		err = Unmarshal(data, &result)
		if err != nil {
			t.Fatalf("Failed to unmarshal int array: %v", err)
		}

		if len(result) != len(smallInts) {
			t.Fatalf("Array length mismatch: expected %d, got %d", len(smallInts), len(result))
		}

		for i, expected := range smallInts {
			if result[i] != expected {
				t.Errorf("Array element %d mismatch: expected %d, got %d", i, expected, result[i])
			}
		}
	})

	t.Run("LargeArrays", func(t *testing.T) {
		// Test larger arrays to verify encoding efficiency
		sizes := []int{100, 1000, 10000}

		for _, size := range sizes {
			t.Run(fmt.Sprintf("Size%d", size), func(t *testing.T) {
				ints := bt.randIntBlock(size)

				data, err := Marshal(ints)
				if err != nil {
					t.Fatalf("Failed to marshal array of size %d: %v", size, err)
				}

				var result []int
				err = Unmarshal(data, &result)
				if err != nil {
					t.Fatalf("Failed to unmarshal array of size %d: %v", size, err)
				}

				if len(result) != len(ints) {
					t.Fatalf("Large array length mismatch: expected %d, got %d", len(ints), len(result))
				}

				for i, expected := range ints {
					if result[i] != expected {
						t.Errorf("Large array element %d mismatch: expected %d, got %d", i, expected, result[i])
						break // Don't spam on failure
					}
				}
			})
		}
	})

	t.Run("TypeConversions", func(t *testing.T) {
		// Test that different integer types can be round-tripped
		var (
			int8Val   int8   = 42
			int16Val  int16  = 1000
			int32Val  int32  = 1000000
			int64Val  int64  = 1000000000000
			uintVal   uint   = 42
			uint8Val  uint8  = 255
			uint16Val uint16 = 65535
			uint32Val uint32 = 4294967295
			uint64Val uint64 = 18446744073709551615
		)

		// Test signed types
		data, _ := Marshal(int8Val)
		var resultInt8 int8
		err := Unmarshal(data, &resultInt8)
		if err != nil || resultInt8 != int8Val {
			t.Errorf("int8 conversion failed: %v, %d != %d", err, resultInt8, int8Val)
		}

		data, _ = Marshal(int16Val)
		var resultInt16 int16
		err = Unmarshal(data, &resultInt16)
		if err != nil || resultInt16 != int16Val {
			t.Errorf("int16 conversion failed: %v, %d != %d", err, resultInt16, int16Val)
		}

		data, _ = Marshal(int32Val)
		var resultInt32 int32
		err = Unmarshal(data, &resultInt32)
		if err != nil || resultInt32 != int32Val {
			t.Errorf("int32 conversion failed: %v, %d != %d", err, resultInt32, int32Val)
		}

		data, _ = Marshal(int64Val)
		var resultInt64 int64
		err = Unmarshal(data, &resultInt64)
		if err != nil || resultInt64 != int64Val {
			t.Errorf("int64 conversion failed: %v, %d != %d", err, resultInt64, int64Val)
		}

		// Test unsigned types (they should be converted to signed for YAJBE)
		data, _ = Marshal(uintVal)
		var resultUint uint
		err = Unmarshal(data, &resultUint)
		if err != nil || resultUint != uintVal {
			t.Errorf("uint conversion failed: %v, %d != %d", err, resultUint, uintVal)
		}

		data, _ = Marshal(uint8Val)
		var resultUint8 uint8
		err = Unmarshal(data, &resultUint8)
		if err != nil || resultUint8 != uint8Val {
			t.Errorf("uint8 conversion failed: %v, %d != %d", err, resultUint8, uint8Val)
		}

		data, _ = Marshal(uint16Val)
		var resultUint16 uint16
		err = Unmarshal(data, &resultUint16)
		if err != nil || resultUint16 != uint16Val {
			t.Errorf("uint16 conversion failed: %v, %d != %d", err, resultUint16, uint16Val)
		}

		data, _ = Marshal(uint32Val)
		var resultUint32 uint32
		err = Unmarshal(data, &resultUint32)
		if err != nil || resultUint32 != uint32Val {
			t.Errorf("uint32 conversion failed: %v, %d != %d", err, resultUint32, uint32Val)
		}

		// Note: uint64 with max value may not round-trip perfectly due to signed representation
		smallUint64 := uint64(1000000)
		data, _ = Marshal(smallUint64)
		var resultSmallUint64 uint64
		err = Unmarshal(data, &resultSmallUint64)
		if err != nil || resultSmallUint64 != smallUint64 {
			t.Errorf("small uint64 conversion failed: %v, %d != %d", err, resultSmallUint64, smallUint64)
		}

		data, _ = Marshal(uint64Val)
		var resultUint64 uint64
		err = Unmarshal(data, &resultUint64)
		if err != nil || resultUint64 != uint64Val {
			t.Errorf("uint32 conversion failed: %v, %d != %d", err, resultUint64, uint64Val)
		}
	})
}

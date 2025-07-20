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
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntSimple(t *testing.T) {
	// Positive ints
	assertEncodeDecode(t, int64(1), "40")
	assertEncodeDecode(t, int64(7), "46")
	assertEncodeDecode(t, int64(24), "57")
	assertEncodeDecode(t, int64(25), "5800")
	assertEncodeDecode(t, int64(127), "5866")
	assertEncodeDecode(t, int64(128), "5867")
	assertEncodeDecode(t, int64(0xff), "58e6")
	assertEncodeDecode(t, int64(0xffff), "59e6ff")
	assertEncodeDecode(t, int64(0xffffff), "5ae6ffff")
	assertEncodeDecode(t, int64(0xffffffff), "5be6ffffff")
	assertEncodeDecode(t, int64(0xffffffffff), "5ce6ffffffff")
	assertEncodeDecode(t, int64(0xffffffffffff), "5de6ffffffffff")
	assertEncodeDecode(t, int64(0x1fffffffffffff), "5ee6ffffffffff1f")
	assertEncodeDecode(t, int64(0xffffffffffffff), "5ee6ffffffffffff")
	assertEncodeDecode(t, int64(0xfffffffffffffff), "5fe6ffffffffffff0f")
	assertEncodeDecode(t, int64(0x7fffffffffffffff), "5fe6ffffffffffff7f")
	
	// Test decode only for boundary values
	assertDecode(t, "5ee6ffffffffffff", int64(0xffffffffffffff))
	assertDecode(t, "5fe6ffffffffffff0f", int64(0xfffffffffffffff))
	assertDecode(t, "5fe6ffffffffffff7f", int64(0x7fffffffffffffff))

	assertEncodeDecode(t, int64(100), "584b")
	assertEncodeDecode(t, int64(1000), "59cf03")
	assertEncodeDecode(t, int64(1000000), "5a27420f")
	assertEncodeDecode(t, int64(1000000000000), "5ce70fa5d4e8")
	assertEncodeDecode(t, int64(100000000000000), "5de73f7a10f35a")

	// Negative ints
	assertEncodeDecode(t, int64(0), "60")
	assertEncodeDecode(t, int64(-1), "61")
	assertEncodeDecode(t, int64(-7), "67")
	assertEncodeDecode(t, int64(-23), "77")
	assertEncodeDecode(t, int64(-24), "7800")
	assertEncodeDecode(t, int64(-25), "7801")
	assertEncodeDecode(t, int64(-0xff), "78e7")
	assertEncodeDecode(t, int64(-0xffff), "79e7ff")
	assertEncodeDecode(t, int64(-0xffffff), "7ae7ffff")
	assertEncodeDecode(t, int64(-0xffffffff), "7be7ffffff")
	assertEncodeDecode(t, int64(-0xffffffffff), "7ce7ffffffff")
	assertEncodeDecode(t, int64(-0xffffffffffff), "7de7ffffffffff")
	assertEncodeDecode(t, int64(-0x1fffffffffffff), "7ee7ffffffffff1f")
	
	// Test decode only for boundary values
	assertDecode(t, "7ee7ffffffffffff", int64(-0xffffffffffffff))
	assertDecode(t, "7fe7ffffffffffff0f", int64(-0xfffffffffffffff))
	assertDecode(t, "7fe7ffffffffffff7f", int64(-0x7fffffffffffffff))

	assertEncodeDecode(t, int64(-100), "784c")
	assertEncodeDecode(t, int64(-1000), "79d003")
	assertEncodeDecode(t, int64(-1000000), "7a28420f")
	assertEncodeDecode(t, int64(-1000000000000), "7ce80fa5d4e8")
	assertEncodeDecode(t, int64(-100000000000000), "7de83f7a10f35a")
}

func TestSmallInlineInt(t *testing.T) {
	expected := []string{
		"790001",
		"78ff", "78fe", "78fd", "78fc", "78fb", "78fa", "78f9", "78f8", "78f7", "78f6", "78f5", "78f4", "78f3", "78f2", "78f1", "78f0",
		"78ef", "78ee", "78ed", "78ec", "78eb", "78ea", "78e9", "78e8", "78e7", "78e6", "78e5", "78e4", "78e3", "78e2", "78e1", "78e0",
		"78df", "78de", "78dd", "78dc", "78db", "78da", "78d9", "78d8", "78d7", "78d6", "78d5", "78d4", "78d3", "78d2", "78d1", "78d0",
		"78cf", "78ce", "78cd", "78cc", "78cb", "78ca", "78c9", "78c8", "78c7", "78c6", "78c5", "78c4", "78c3", "78c2", "78c1", "78c0",
		"78bf", "78be", "78bd", "78bc", "78bb", "78ba", "78b9", "78b8", "78b7", "78b6", "78b5", "78b4", "78b3", "78b2", "78b1", "78b0",
		"78af", "78ae", "78ad", "78ac", "78ab", "78aa", "78a9", "78a8", "78a7", "78a6", "78a5", "78a4", "78a3", "78a2", "78a1", "78a0",
		"789f", "789e", "789d", "789c", "789b", "789a", "7899", "7898", "7897", "7896", "7895", "7894", "7893", "7892", "7891", "7890",
		"788f", "788e", "788d", "788c", "788b", "788a", "7889", "7888", "7887", "7886", "7885", "7884", "7883", "7882", "7881", "7880",
		"787f", "787e", "787d", "787c", "787b", "787a", "7879", "7878", "7877", "7876", "7875", "7874", "7873", "7872", "7871", "7870",
		"786f", "786e", "786d", "786c", "786b", "786a", "7869", "7868", "7867", "7866", "7865", "7864", "7863", "7862", "7861", "7860",
		"785f", "785e", "785d", "785c", "785b", "785a", "7859", "7858", "7857", "7856", "7855", "7854", "7853", "7852", "7851", "7850",
		"784f", "784e", "784d", "784c", "784b", "784a", "7849", "7848", "7847", "7846", "7845", "7844", "7843", "7842", "7841", "7840",
		"783f", "783e", "783d", "783c", "783b", "783a", "7839", "7838", "7837", "7836", "7835", "7834", "7833", "7832", "7831", "7830",
		"782f", "782e", "782d", "782c", "782b", "782a", "7829", "7828", "7827", "7826", "7825", "7824", "7823", "7822", "7821", "7820",
		"781f", "781e", "781d", "781c", "781b", "781a", "7819", "7818", "7817", "7816", "7815", "7814", "7813", "7812", "7811", "7810",
		"780f", "780e", "780d", "780c", "780b", "780a", "7809", "7808", "7807", "7806", "7805", "7804", "7803", "7802", "7801", "7800",
		"77", "76", "75", "74", "73", "72", "71", "70", "6f", "6e", "6d", "6c", "6b", "6a", "69", "68", "67", "66", "65", "64", "63", "62", "61", "60",
		"40", "41", "42", "43", "44", "45", "46", "47", "48", "49", "4a", "4b", "4c", "4d", "4e", "4f", "50", "51", "52", "53", "54", "55", "56", "57", "5800",
		"5801", "5802", "5803", "5804", "5805", "5806", "5807", "5808", "5809", "580a", "580b", "580c", "580d", "580e", "580f", "5810",
		"5811", "5812", "5813", "5814", "5815", "5816", "5817", "5818", "5819", "581a", "581b", "581c", "581d", "581e", "581f", "5820",
		"5821", "5822", "5823", "5824", "5825", "5826", "5827", "5828", "5829", "582a", "582b", "582c", "582d", "582e", "582f", "5830",
		"5831", "5832", "5833", "5834", "5835", "5836", "5837", "5838", "5839", "583a", "583b", "583c", "583d", "583e", "583f", "5840",
		"5841", "5842", "5843", "5844", "5845", "5846", "5847", "5848", "5849", "584a", "584b", "584c", "584d", "584e", "584f", "5850",
		"5851", "5852", "5853", "5854", "5855", "5856", "5857", "5858", "5859", "585a", "585b", "585c", "585d", "585e", "585f", "5860",
		"5861", "5862", "5863", "5864", "5865", "5866", "5867", "5868", "5869", "586a", "586b", "586c", "586d", "586e", "586f", "5870",
		"5871", "5872", "5873", "5874", "5875", "5876", "5877", "5878", "5879", "587a", "587b", "587c", "587d", "587e", "587f", "5880",
		"5881", "5882", "5883", "5884", "5885", "5886", "5887", "5888", "5889", "588a", "588b", "588c", "588d", "588e", "588f", "5890",
		"5891", "5892", "5893", "5894", "5895", "5896", "5897", "5898", "5899", "589a", "589b", "589c", "589d", "589e", "589f", "58a0",
		"58a1", "58a2", "58a3", "58a4", "58a5", "58a6", "58a7", "58a8", "58a9", "58aa", "58ab", "58ac", "58ad", "58ae", "58af", "58b0",
		"58b1", "58b2", "58b3", "58b4", "58b5", "58b6", "58b7", "58b8", "58b9", "58ba", "58bb", "58bc", "58bd", "58be", "58bf", "58c0",
		"58c1", "58c2", "58c3", "58c4", "58c5", "58c6", "58c7", "58c8", "58c9", "58ca", "58cb", "58cc", "58cd", "58ce", "58cf", "58d0",
		"58d1", "58d2", "58d3", "58d4", "58d5", "58d6", "58d7", "58d8", "58d9", "58da", "58db", "58dc", "58dd", "58de", "58df", "58e0",
		"58e1", "58e2", "58e3", "58e4", "58e5", "58e6", "58e7", "58e8", "58e9", "58ea", "58eb", "58ec", "58ed", "58ee", "58ef", "58f0",
		"58f1", "58f2", "58f3", "58f4", "58f5", "58f6", "58f7", "58f8", "58f9", "58fa", "58fb", "58fc", "58fd", "58fe", "58ff",
		"590001",
	}

	value := int64(-280)
	for _, hexStr := range expected {
		assertEncodeDecode(t, value, hexStr)
		value++
	}
}

func TestRandIntEncodeDecode(t *testing.T) {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	
	for i := 0; i < 1000; i++ {
		input := int64(rng.Intn(2147483647))
		enc, err := Marshal(input)
		require.NoError(t, err)
		
		var decoded interface{}
		err = Unmarshal(enc, &decoded)
		require.NoError(t, err)
		
		// For JSON compatibility, integers decode as float64 when target is interface{}
		assert.Equal(t, float64(input), decoded)
	}
}

func TestRandLongEncodeDecode(t *testing.T) {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	
	for i := 0; i < 1000; i++ {
		input := int64(rng.Float64() * 9223372036854775807)
		enc, err := Marshal(input)
		require.NoError(t, err)
		
		var decoded interface{}
		err = Unmarshal(enc, &decoded)
		require.NoError(t, err)
		
		// For JSON compatibility, integers decode as float64 when target is interface{}
		assert.Equal(t, float64(input), decoded)
	}
}

func TestRandIntArrayEncodeDecode(t *testing.T) {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	
	for k := 0; k < 32; k++ {
		length := rng.Intn(1 << 15)
		input := make([]interface{}, length)
		for i := 0; i < length; i++ {
			input[i] = int64(rng.Intn(2147483647))
		}
		
		enc, err := Marshal(input)
		require.NoError(t, err)
		
		var decoded interface{}
		err = Unmarshal(enc, &decoded)
		require.NoError(t, err)
		
		// Handle type conversion - integers decode as float64 for JSON compatibility
		if decodedSlice, ok := decoded.([]interface{}); ok {
			assert.Equal(t, len(input), len(decodedSlice))
			for i, inputVal := range input {
				expectedVal := float64(inputVal.(int64)) // Convert to float64 for comparison
				assert.Equal(t, expectedVal, decodedSlice[i])
			}
		} else {
			t.Errorf("Expected []interface{}, got %T", decoded)
		}
	}
}

func TestRandLongArrayEncodeDecode(t *testing.T) {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	
	for k := 0; k < 32; k++ {
		length := rng.Intn(1 << 15)
		input := make([]interface{}, length)
		for i := 0; i < length; i++ {
			input[i] = int64(rng.Float64() * 9223372036854775807)
		}
		
		enc, err := Marshal(input)
		require.NoError(t, err)
		
		var decoded interface{}
		err = Unmarshal(enc, &decoded)
		require.NoError(t, err)
		
		// Handle type conversion - integers decode as float64 for JSON compatibility
		if decodedSlice, ok := decoded.([]interface{}); ok {
			assert.Equal(t, len(input), len(decodedSlice))
			for i, inputVal := range input {
				expectedVal := float64(inputVal.(int64)) // Convert to float64 for comparison
				assert.Equal(t, expectedVal, decodedSlice[i])
			}
		} else {
			t.Errorf("Expected []interface{}, got %T", decoded)
		}
	}
}
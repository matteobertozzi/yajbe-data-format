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
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBytesSimple(t *testing.T) {
	assertEncodeDecode(t, make([]byte, 0), "80")
	assertEncodeDecode(t, make([]byte, 1), "8100")
	assertEncodeDecode(t, make([]byte, 3), "83000000")
	assertEncodeDecode(t, make([]byte, 59), "bb"+strings.Repeat("00", 59))
	assertEncodeDecode(t, make([]byte, 60), "bc01"+strings.Repeat("00", 60))
	assertEncodeDecode(t, make([]byte, 127), "bc44"+strings.Repeat("00", 127))
	assertEncodeDecode(t, make([]byte, 0xff), "bcc4"+strings.Repeat("00", 255))
	assertEncodeDecode(t, make([]byte, 256), "bcc5"+strings.Repeat("00", 256))
	assertEncodeDecode(t, make([]byte, 314), "bcff"+strings.Repeat("00", 314))
	assertEncodeDecode(t, make([]byte, 315), "bd0001"+strings.Repeat("00", 315))
	assertEncodeDecode(t, make([]byte, 0xffff), "bdc4ff"+strings.Repeat("00", 0xffff))
	assertEncodeDecode(t, make([]byte, 0xfffff), "bec4ff0f"+strings.Repeat("00", 0xfffff))

	// Test encoding only for ByteData equivalents (same as Uint8List)
	assertEncode(t, make([]byte, 0), "80")
	assertEncode(t, make([]byte, 1), "8100")
	assertEncode(t, make([]byte, 3), "83000000")
	assertEncode(t, make([]byte, 59), "bb"+strings.Repeat("00", 59))
	assertEncode(t, make([]byte, 60), "bc01"+strings.Repeat("00", 60))
	assertEncode(t, make([]byte, 127), "bc44"+strings.Repeat("00", 127))
}

func TestBytesSmallLength(t *testing.T) {
	// Test byte arrays 0-59 bytes (1 byte overhead)
	for i := 0; i < 60; i++ {
		input := make([]byte, i)
		enc, err := Marshal(input)
		require.NoError(t, err)
		assert.Equal(t, 1+i, len(enc))
		
		var decoded []byte
		err = Unmarshal(enc, &decoded)
		require.NoError(t, err)
		assert.Equal(t, input, decoded)
	}

	// Test byte arrays 60-314 bytes (2 byte overhead)
	for i := 60; i <= 314; i++ {
		input := make([]byte, i)
		enc, err := Marshal(input)
		require.NoError(t, err)
		assert.Equal(t, 2+i, len(enc))
		
		var decoded []byte
		err = Unmarshal(enc, &decoded)
		require.NoError(t, err)
		assert.Equal(t, input, decoded)
	}

	// Test byte arrays 315-0xfff bytes (3 byte overhead)
	for i := 315; i <= 0xfff; i++ {
		input := make([]byte, i)
		enc, err := Marshal(input)
		require.NoError(t, err)
		assert.Equal(t, 3+i, len(enc))
		
		var decoded []byte
		err = Unmarshal(enc, &decoded)
		require.NoError(t, err)
		assert.Equal(t, input, decoded)
	}
}

func TestBytesRandEncodeDecode(t *testing.T) {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	
	for i := 0; i < 100; i++ {
		length := rng.Intn(1 << 16)
		input := make([]byte, length)
		for k := 0; k < length; k++ {
			input[k] = byte(rng.Intn(0xff))
		}
		
		enc, err := Marshal(input)
		require.NoError(t, err)
		
		var decoded []byte
		err = Unmarshal(enc, &decoded)
		require.NoError(t, err)
		assert.Equal(t, input, decoded)
	}
}
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

const textChars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func randText(length int) string {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	var sb strings.Builder
	for i := 0; i < length; i++ {
		wordLength := 4 + rng.Intn(8)
		for w := 0; w < wordLength; w++ {
			sb.WriteByte(textChars[rng.Intn(len(textChars))])
		}
		sb.WriteByte(' ')
	}
	return sb.String()
}

func TestStringSimple(t *testing.T) {
	assertEncodeDecode(t, "", "c0")
	assertEncodeDecode(t, "a", "c161")
	assertEncodeDecode(t, "abc", "c3616263")
	assertEncodeDecode(t, strings.Repeat("x", 59), "fb"+strings.Repeat("78", 59))
	assertEncodeDecode(t, strings.Repeat("y", 60), "fc01"+strings.Repeat("79", 60))
	assertEncodeDecode(t, strings.Repeat("y", 127), "fc44"+strings.Repeat("79", 127))
	assertEncodeDecode(t, strings.Repeat("y", 255), "fcc4"+strings.Repeat("79", 255))
	assertEncodeDecode(t, strings.Repeat("z", 0x100), "fcc5"+strings.Repeat("7a", 256))
	assertEncodeDecode(t, strings.Repeat("z", 314), "fcff"+strings.Repeat("7a", 314))
	assertEncodeDecode(t, strings.Repeat("z", 315), "fd0001"+strings.Repeat("7a", 315))
	assertEncodeDecode(t, strings.Repeat("z", 0xffff), "fdc4ff"+strings.Repeat("7a", 0xffff))
	assertEncodeDecode(t, strings.Repeat("k", 0xfffff), "fec4ff0f"+strings.Repeat("6b", 0xfffff))
	assertEncodeDecode(t, strings.Repeat("k", 0xffffff), "fec4ffff"+strings.Repeat("6b", 0xffffff))
	assertEncodeDecode(t, strings.Repeat("k", 0x1000000), "fec5ffff"+strings.Repeat("6b", 0x1000000))
	assertEncodeDecode(t, strings.Repeat("k", 0x1000123), "ffe8000001"+strings.Repeat("6b", 0x1000123))
}

func TestSmallString(t *testing.T) {
	// Test strings 0-59 characters (1 byte overhead)
	for i := 0; i < 60; i++ {
		input := strings.Repeat("x", i)
		enc, err := Marshal(input)
		require.NoError(t, err)
		assert.Equal(t, 1+i, len(enc))
		
		var decoded string
		err = Unmarshal(enc, &decoded)
		require.NoError(t, err)
		assert.Equal(t, input, decoded)
	}

	// Test strings 60-314 characters (2 byte overhead)
	for i := 60; i <= 314; i++ {
		input := strings.Repeat("x", i)
		enc, err := Marshal(input)
		require.NoError(t, err)
		assert.Equal(t, 2+i, len(enc))
		
		var decoded string
		err = Unmarshal(enc, &decoded)
		require.NoError(t, err)
		assert.Equal(t, input, decoded)
	}

	// Test strings 315-0x1fff characters (3 byte overhead)
	for i := 315; i <= 0x1fff; i++ {
		input := strings.Repeat("x", i)
		enc, err := Marshal(input)
		require.NoError(t, err)
		assert.Equal(t, 3+i, len(enc))
		
		var decoded string
		err = Unmarshal(enc, &decoded)
		require.NoError(t, err)
		assert.Equal(t, input, decoded)
	}
}

func TestRandStringEncodeDecode(t *testing.T) {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	
	for i := 0; i < 10; i++ {
		length := rng.Intn(1 << 14)
		input := randText(length)
		
		enc, err := Marshal(input)
		require.NoError(t, err)
		
		var decoded string
		err = Unmarshal(enc, &decoded)
		require.NoError(t, err)
		assert.Equal(t, input, decoded)
	}
}
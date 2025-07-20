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

func TestBoolSimple(t *testing.T) {
	assertEncodeDecode(t, false, "02")
	assertEncodeDecode(t, true, "03")
}

func TestBoolArray(t *testing.T) {
	// All true small array (7 elements)
	allTrueSmall := make([]bool, 7)
	for i := range allTrueSmall {
		allTrueSmall[i] = true
	}
	assertEncodeDecode(t, allTrueSmall, "2703030303030303")

	// All true large array (310 elements)
	allTrueLarge := make([]bool, 310)
	for i := range allTrueLarge {
		allTrueLarge[i] = true
	}
	expectedLarge := "2c2c01" + strings.Repeat("03", 310)
	assertEncodeDecode(t, allTrueLarge, expectedLarge)

	// All false small array (4 elements)
	allFalseSmall := make([]bool, 4)
	assertEncodeDecode(t, allFalseSmall, "2402020202")

	// All false large array (128 elements)
	allFalseLarge := make([]bool, 128)
	expectedFalseLarge := "2b76" + strings.Repeat("02", 128)
	assertEncodeDecode(t, allFalseLarge, expectedFalseLarge)

	// Mixed small array (10 elements)
	mixSmall := make([]bool, 10)
	for i := 0; i < len(mixSmall); i++ {
		mixSmall[i] = (i & 2) == 0
	}
	assertEncodeDecode(t, mixSmall, "2a03030202030302020303")

	// Mixed large array (128 elements)
	mixLarge := make([]bool, 128)
	for i := 0; i < len(mixLarge); i++ {
		mixLarge[i] = (i & 3) == 0
	}
	expectedMixLarge := "2b76" + strings.Repeat("03020202", 32)
	assertEncodeDecode(t, mixLarge, expectedMixLarge)
}

func TestBoolRandArray(t *testing.T) {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	
	for k := 0; k < 32; k++ {
		length := rng.Intn(1 << 16)
		items := make([]bool, length)
		for i := 0; i < length; i++ {
			items[i] = rng.Float64() > 0.5
		}
		
		enc, err := Marshal(items)
		require.NoError(t, err)
		
		var decoded interface{}
		err = Unmarshal(enc, &decoded)
		require.NoError(t, err)
		
		// Handle type conversion - Go decodes bool arrays as []interface{} for interface{}
		if decodedSlice, ok := decoded.([]interface{}); ok {
			assert.Equal(t, len(items), len(decodedSlice))
			for i, expectedVal := range items {
				assert.Equal(t, expectedVal, decodedSlice[i])
			}
		} else {
			assert.Equal(t, items, decoded)
		}
	}
}
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

// TestSpecific127And128 tests the specific encoding for 127 and 128 values that were mentioned in commits
func TestSpecific127And128(t *testing.T) {
	// From Dart test: 127 should encode as "5866", 128 should encode as "5867"
	
	// Test 127
	enc, err := Marshal(int64(127))
	require.NoError(t, err)
	actual := hex.EncodeToString(enc)
	assert.Equal(t, "5866", actual, "127 encoding mismatch")
	
	// Test 128
	enc, err = Marshal(int64(128))
	require.NoError(t, err)
	actual = hex.EncodeToString(enc)
	assert.Equal(t, "5867", actual, "128 encoding mismatch")
	
	// Test decoding
	data, err := hex.DecodeString("5866")
	require.NoError(t, err)
	var decoded interface{}
	err = Unmarshal(data, &decoded)
	require.NoError(t, err)
	assert.Equal(t, float64(127), decoded, "127 decoding mismatch")
	
	data, err = hex.DecodeString("5867")
	require.NoError(t, err)
	err = Unmarshal(data, &decoded)
	require.NoError(t, err)
	assert.Equal(t, float64(128), decoded, "128 decoding mismatch")
}
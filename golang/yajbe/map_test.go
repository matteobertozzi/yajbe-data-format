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
	"testing"
)

func TestMapSimple(t *testing.T) {
	// Test decode-only for special map encodings
	assertDecode(t, "30", map[string]interface{}{})
	assertDecode(t, "3f01", map[string]interface{}{})
	assertDecode(t, "3f81614001", map[string]interface{}{"a": float64(1)}) // int decodes as float64
	assertDecode(t, "3f8161c2764101", map[string]interface{}{"a": "vA"})
	assertDecode(t, "3f81612340414201", map[string]interface{}{"a": []interface{}{float64(1), float64(2), float64(3)}})
	assertDecode(t, "3f81613f816c234041420101", map[string]interface{}{"a": map[string]interface{}{"l": []interface{}{float64(1), float64(2), float64(3)}}})
	assertDecode(t, "3f81613f816c3f817840010101", map[string]interface{}{"a": map[string]interface{}{"l": map[string]interface{}{"x": float64(1)}}})

	assertDecode(t, "3f816140836f626a0001", map[string]interface{}{"a": float64(1), "obj": nil})
	assertDecode(t, "3f816140836f626a3fa041a1000101", map[string]interface{}{"a": float64(1), "obj": map[string]interface{}{"a": float64(2), "obj": nil}})
	assertDecode(t, "3f816140836f626a3fa041a13fa042a100010101", map[string]interface{}{"a": float64(1), "obj": map[string]interface{}{"a": float64(2), "obj": map[string]interface{}{"a": float64(3), "obj": nil}}})

	// Test encode/decode for regular maps
	assertEncodeDecode(t, map[string]interface{}{"a": int64(1), "b": int64(2)}, "32816140816241")
	assertEncodeDecode(t, map[string]interface{}{"a": int64(1), "b": int64(2), "c": int64(3)}, "33816140816241816342")
	assertEncodeDecode(t, map[string]interface{}{"a": int64(1), "b": int64(2), "c": int64(3), "d": int64(4)}, "34816140816241816342816443")
	assertEncodeDecode(t, map[string]interface{}{"a": []interface{}{int64(1), int64(2), int64(3)}}, "31816123404142")
	assertEncodeDecode(t, map[string]interface{}{"a": map[string]interface{}{"l": []interface{}{int64(1), int64(2), int64(3)}}}, "31816131816c23404142")
	assertEncodeDecode(t, map[string]interface{}{"a": map[string]interface{}{"l": map[string]interface{}{"x": int64(1)}}}, "31816131816c31817840")
}

func TestMapTypes(t *testing.T) {
	input := map[string]interface{}{
		"aaa": int64(1),
		"bbb": map[string]interface{}{"k": int64(10)},
		"ccc": 2.3,
		"ddd": []interface{}{"a", "b"},
		"eee": []interface{}{"a", map[string]interface{}{"k": int64(10)}, "b"},
		"fff": map[string]interface{}{"a": map[string]interface{}{"k": []interface{}{"z", "d"}}},
		"ggg": "foo",
	}
	
	assertEncodeDecode(t, input, "3783616161408362626231816b49836363630666666666666602408364646422c161c1628365656523c16131a249c1628366666631816131a222c17ac16483676767c3666f6f")
	
	// Test decode-only version with special encoding
	expectedForDecode := map[string]interface{}{
		"aaa": float64(1), // int decodes as float64
		"bbb": map[string]interface{}{"k": float64(10)}, // int decodes as float64
		"ccc": 2.3,
		"ddd": []interface{}{"a", "b"},
		"eee": []interface{}{"a", map[string]interface{}{"k": float64(10)}, "b"}, // int decodes as float64
		"fff": map[string]interface{}{"a": map[string]interface{}{"k": []interface{}{"z", "d"}}},
		"ggg": "foo",
	}
	assertDecode(t, "3f8361616140836262623f816b4901836363630666666666666602408364646422c161c1628365656523c1613fa24901c162836666663f81613fa222c17ac164010183676767c3666f6f01", expectedForDecode)
}
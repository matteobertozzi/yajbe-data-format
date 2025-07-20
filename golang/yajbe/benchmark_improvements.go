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

// Benchmark for the improved encoder/decoder generic functions

func BenchmarkEncoderGeneric(b *testing.B) {
	testData := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Marshal(testData)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDecoderGeneric(b *testing.B) {
	testData := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	encoded, err := Marshal(testData)
	if err != nil {
		b.Fatal(err)
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var decoded []int
		err := Unmarshal(encoded, &decoded)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkStringMapEncoding(b *testing.B) {
	testData := map[string]int{
		"key1": 1, "key2": 2, "key3": 3, "key4": 4, "key5": 5,
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Marshal(testData)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkStringMapDecoding(b *testing.B) {
	testData := map[string]int{
		"key1": 1, "key2": 2, "key3": 3, "key4": 4, "key5": 5,
	}
	encoded, err := Marshal(testData)
	if err != nil {
		b.Fatal(err)
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var decoded map[string]int
		err := Unmarshal(encoded, &decoded)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkErrorHandling(b *testing.B) {
	// Benchmark the improved error handling
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var invalidTarget int
		err := Unmarshal([]byte{0x00}, &invalidTarget) // null into int should error
		if err == nil {
			b.Fatal("Expected error but got nil")
		}
	}
}
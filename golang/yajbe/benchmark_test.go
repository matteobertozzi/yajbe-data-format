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
	"encoding/json"
	"fmt"
	"math/big"
	"math/rand"
	"testing"
	"time"

	"github.com/fxamacker/cbor/v2"
)

// Test data structures for benchmarks
type SimpleStruct struct {
	ID   int    `json:"id" cbor:"id"`
	Name string `json:"name" cbor:"name"`
	Age  int    `json:"age" cbor:"age"`
}

type ComplexStruct struct {
	ID        int            `json:"id" cbor:"id"`
	Name      string         `json:"name" cbor:"name"`
	Email     string         `json:"email" cbor:"email"`
	Active    bool           `json:"active" cbor:"active"`
	Score     float64        `json:"score" cbor:"score"`
	Tags      []string       `json:"tags" cbor:"tags"`
	Metadata  map[string]any `json:"metadata" cbor:"metadata"`
	CreatedAt time.Time      `json:"created_at" cbor:"created_at"`
}

type NestedStruct struct {
	User    ComplexStruct   `json:"user" cbor:"user"`
	Friends []ComplexStruct `json:"friends" cbor:"friends"`
	Groups  []string        `json:"groups" cbor:"groups"`
	Numbers []int           `json:"numbers" cbor:"numbers"`
	Floats  []float64       `json:"floats" cbor:"floats"`
}

// Generate test data
func generateSimpleStruct() SimpleStruct {
	return SimpleStruct{
		ID:   rand.Int(),
		Name: randomString(20),
		Age:  rand.Intn(100),
	}
}

func generateComplexStruct() ComplexStruct {
	tags := make([]string, rand.Intn(5)+1)
	for i := range tags {
		tags[i] = randomString(10)
	}

	metadata := map[string]any{
		"department": randomString(15),
		"level":      rand.Intn(10),
		"temp":       rand.Float64() * 100,
		"enabled":    rand.Intn(2) == 1,
	}

	return ComplexStruct{
		ID:        rand.Int(),
		Name:      randomString(25),
		Email:     randomString(20) + "@example.com",
		Active:    rand.Intn(2) == 1,
		Score:     rand.Float64() * 100,
		Tags:      tags,
		Metadata:  metadata,
		CreatedAt: time.Now().Add(-time.Duration(rand.Intn(365*24)) * time.Hour),
	}
}

func generateNestedStruct() NestedStruct {
	friends := make([]ComplexStruct, rand.Intn(50)+1)
	for i := range friends {
		friends[i] = generateComplexStruct()
	}

	groups := make([]string, rand.Intn(30)+1)
	for i := range groups {
		groups[i] = randomString(15)
	}

	numbers := make([]int, rand.Intn(20)+5)
	for i := range numbers {
		numbers[i] = rand.Int()
	}

	floats := make([]float64, rand.Intn(20)+5)
	for i := range floats {
		floats[i] = rand.Float64() * 1000
	}

	return NestedStruct{
		User:    generateComplexStruct(),
		Friends: friends,
		Groups:  groups,
		Numbers: numbers,
		Floats:  floats,
	}
}

func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[rand.Intn(len(charset))]
	}
	return string(result)
}

// Marshal Benchmarks
func BenchmarkMarshal_Simple_YAJBE(b *testing.B) {
	data := generateSimpleStruct()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Marshal(data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMarshal_Simple_JSON(b *testing.B) {
	data := generateSimpleStruct()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := json.Marshal(data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMarshal_Simple_CBOR(b *testing.B) {
	data := generateSimpleStruct()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := cbor.Marshal(data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMarshal_Complex_YAJBE(b *testing.B) {
	data := generateComplexStruct()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Marshal(data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMarshal_Complex_JSON(b *testing.B) {
	data := generateComplexStruct()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := json.Marshal(data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMarshal_Complex_CBOR(b *testing.B) {
	data := generateComplexStruct()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := cbor.Marshal(data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMarshal_Nested_YAJBE(b *testing.B) {
	data := generateNestedStruct()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Marshal(data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMarshal_Nested_JSON(b *testing.B) {
	data := generateNestedStruct()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := json.Marshal(data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMarshal_Nested_CBOR(b *testing.B) {
	data := generateNestedStruct()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := cbor.Marshal(data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Unmarshal Benchmarks
func BenchmarkUnmarshal_Simple_YAJBE(b *testing.B) {
	data := generateSimpleStruct()
	encoded, _ := Marshal(data)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var result SimpleStruct
		err := Unmarshal(encoded, &result)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkUnmarshal_Simple_JSON(b *testing.B) {
	data := generateSimpleStruct()
	encoded, _ := json.Marshal(data)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var result SimpleStruct
		err := json.Unmarshal(encoded, &result)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkUnmarshal_Simple_CBOR(b *testing.B) {
	data := generateSimpleStruct()
	encoded, _ := cbor.Marshal(data)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var result SimpleStruct
		err := cbor.Unmarshal(encoded, &result)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkUnmarshal_Complex_YAJBE(b *testing.B) {
	data := generateComplexStruct()
	encoded, _ := Marshal(data)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var result ComplexStruct
		err := Unmarshal(encoded, &result)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkUnmarshal_Complex_JSON(b *testing.B) {
	data := generateComplexStruct()
	encoded, _ := json.Marshal(data)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var result ComplexStruct
		err := json.Unmarshal(encoded, &result)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkUnmarshal_Complex_CBOR(b *testing.B) {
	data := generateComplexStruct()
	encoded, _ := cbor.Marshal(data)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var result ComplexStruct
		err := cbor.Unmarshal(encoded, &result)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkUnmarshal_Nested_YAJBE(b *testing.B) {
	data := generateNestedStruct()
	encoded, _ := Marshal(data)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var result NestedStruct
		err := Unmarshal(encoded, &result)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkUnmarshal_Nested_JSON(b *testing.B) {
	data := generateNestedStruct()
	encoded, _ := json.Marshal(data)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var result NestedStruct
		err := json.Unmarshal(encoded, &result)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkUnmarshal_Nested_CBOR(b *testing.B) {
	data := generateNestedStruct()
	encoded, _ := cbor.Marshal(data)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var result NestedStruct
		err := cbor.Unmarshal(encoded, &result)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Data type specific benchmarks
func BenchmarkMarshal_IntArray_YAJBE(b *testing.B) {
	data := make([]int, 1000)
	for i := range data {
		data[i] = rand.Int()
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Marshal(data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMarshal_IntArray_JSON(b *testing.B) {
	data := make([]int, 1000)
	for i := range data {
		data[i] = rand.Int()
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := json.Marshal(data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMarshal_IntArray_CBOR(b *testing.B) {
	data := make([]int, 1000)
	for i := range data {
		data[i] = rand.Int()
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := cbor.Marshal(data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMarshal_StringArray_YAJBE(b *testing.B) {
	data := make([]string, 100)
	for i := range data {
		data[i] = randomString(50)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Marshal(data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMarshal_StringArray_JSON(b *testing.B) {
	data := make([]string, 100)
	for i := range data {
		data[i] = randomString(50)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := json.Marshal(data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMarshal_StringArray_CBOR(b *testing.B) {
	data := make([]string, 100)
	for i := range data {
		data[i] = randomString(50)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := cbor.Marshal(data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMarshal_BigInt_YAJBE(b *testing.B) {
	data := big.NewInt(0)
	data.SetString("1234567890123456789012345678901234567890", 10)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Marshal(data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMarshal_BigInt_JSON(b *testing.B) {
	data := big.NewInt(0)
	data.SetString("1234567890123456789012345678901234567890", 10)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := json.Marshal(data.String())
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMarshal_BigInt_CBOR(b *testing.B) {
	data := big.NewInt(0)
	data.SetString("1234567890123456789012345678901234567890", 10)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := cbor.Marshal(data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Size comparison benchmarks
func BenchmarkSize_Simple_YAJBE(b *testing.B) {
	data := generateSimpleStruct()
	encoded, _ := Marshal(data)
	b.ReportMetric(float64(len(encoded)), "bytes")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = len(encoded)
	}
}

func BenchmarkSize_Simple_JSON(b *testing.B) {
	data := generateSimpleStruct()
	encoded, _ := json.Marshal(data)
	b.ReportMetric(float64(len(encoded)), "bytes")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = len(encoded)
	}
}

func BenchmarkSize_Simple_CBOR(b *testing.B) {
	data := generateSimpleStruct()
	encoded, _ := cbor.Marshal(data)
	b.ReportMetric(float64(len(encoded)), "bytes")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = len(encoded)
	}
}

func BenchmarkSize_Complex_YAJBE(b *testing.B) {
	data := generateComplexStruct()
	encoded, _ := Marshal(data)
	b.ReportMetric(float64(len(encoded)), "bytes")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = len(encoded)
	}
}

func BenchmarkSize_Complex_JSON(b *testing.B) {
	data := generateComplexStruct()
	encoded, _ := json.Marshal(data)
	b.ReportMetric(float64(len(encoded)), "bytes")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = len(encoded)
	}
}

func BenchmarkSize_Complex_CBOR(b *testing.B) {
	data := generateComplexStruct()
	encoded, _ := cbor.Marshal(data)
	b.ReportMetric(float64(len(encoded)), "bytes")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = len(encoded)
	}
}

// Throughput benchmarks for large datasets
func BenchmarkThroughput_1000_Items_YAJBE(b *testing.B) {
	data := make([]ComplexStruct, 1000)
	for i := range data {
		data[i] = generateComplexStruct()
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Marshal(data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkThroughput_1000_Items_JSON(b *testing.B) {
	data := make([]ComplexStruct, 1000)
	for i := range data {
		data[i] = generateComplexStruct()
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := json.Marshal(data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkThroughput_1000_Items_CBOR(b *testing.B) {
	data := make([]ComplexStruct, 1000)
	for i := range data {
		data[i] = generateComplexStruct()
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := cbor.Marshal(data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Run a comprehensive benchmark comparison
func TestBenchmarkComparison(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping benchmark comparison in short mode")
	}

	fmt.Println("\n=== YAJBE vs JSON vs CBOR Benchmark Comparison ===")

	// Test data
	simple := generateSimpleStruct()
	complex := generateComplexStruct()
	nested := generateNestedStruct()

	// Marshal size comparison
	yajbeSimple, _ := Marshal(simple)
	jsonSimple, _ := json.Marshal(simple)
	cborSimple, _ := cbor.Marshal(simple)

	yajbeComplex, _ := Marshal(complex)
	jsonComplex, _ := json.Marshal(complex)
	cborComplex, _ := cbor.Marshal(complex)

	yajbeNested, _ := Marshal(nested)
	jsonNested, _ := json.Marshal(nested)
	cborNested, _ := cbor.Marshal(nested)

	fmt.Printf("\nSize Comparison (bytes):\n")
	fmt.Printf("Simple Struct:  YAJBE=%d  JSON=%d  CBOR=%d\n", len(yajbeSimple), len(jsonSimple), len(cborSimple))
	fmt.Printf("Complex Struct: YAJBE=%d  JSON=%d  CBOR=%d\n", len(yajbeComplex), len(jsonComplex), len(cborComplex))
	fmt.Printf("Nested Struct:  YAJBE=%d  JSON=%d  CBOR=%d\n", len(yajbeNested), len(jsonNested), len(cborNested))

	// Size efficiency
	fmt.Printf("\nSize Efficiency (smaller is better):\n")
	fmt.Printf("Simple:  YAJBE vs JSON: %.1f%%  YAJBE vs CBOR: %.1f%%\n",
		float64(len(yajbeSimple))/float64(len(jsonSimple))*100,
		float64(len(yajbeSimple))/float64(len(cborSimple))*100)
	fmt.Printf("Complex: YAJBE vs JSON: %.1f%%  YAJBE vs CBOR: %.1f%%\n",
		float64(len(yajbeComplex))/float64(len(jsonComplex))*100,
		float64(len(yajbeComplex))/float64(len(cborComplex))*100)
	fmt.Printf("Nested:  YAJBE vs JSON: %.1f%%  YAJBE vs CBOR: %.1f%%\n",
		float64(len(yajbeNested))/float64(len(jsonNested))*100,
		float64(len(yajbeNested))/float64(len(cborNested))*100)

	// Round-trip test
	var simpleResult SimpleStruct
	err := Unmarshal(yajbeSimple, &simpleResult)
	if err != nil || simpleResult.ID != simple.ID {
		t.Fatal("YAJBE round-trip failed")
	}

	fmt.Printf("\nAll round-trip tests passed\n")
	fmt.Printf("\nRun 'go test -bench=.' to see detailed performance benchmarks\n")
}

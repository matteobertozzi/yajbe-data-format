// +build ignore

// Standalone benchmark runner
package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/fxamacker/cbor/v2"
	yajbe "./yajbe" // You'll need to adjust this import
)

// Copy the same structures and generators from benchmark_test.go
type Person struct {
	ID       int64             `json:"id" cbor:"id"`
	Name     string            `json:"name" cbor:"name"`
	Email    string            `json:"email" cbor:"email"`
	Age      int               `json:"age" cbor:"age"`
	Active   bool              `json:"active" cbor:"active"`
	Balance  float64           `json:"balance" cbor:"balance"`
	Tags     []string          `json:"tags" cbor:"tags"`
	Metadata map[string]string `json:"metadata" cbor:"metadata"`
}

func generatePerson(id int64) Person {
	rand.Seed(id)
	tags := make([]string, rand.Intn(5)+1)
	for i := range tags {
		tags[i] = fmt.Sprintf("tag%d", i)
	}
	
	metadata := make(map[string]string)
	for i := 0; i < rand.Intn(3)+1; i++ {
		metadata[fmt.Sprintf("key%d", i)] = fmt.Sprintf("value%d", i)
	}
	
	return Person{
		ID:       id,
		Name:     fmt.Sprintf("Person %d", id),
		Email:    fmt.Sprintf("person%d@example.com", id),
		Age:      rand.Intn(80) + 18,
		Active:   rand.Intn(2) == 1,
		Balance:  rand.Float64() * 100000,
		Tags:     tags,
		Metadata: metadata,
	}
}

func generatePersons(count int) []Person {
	persons := make([]Person, count)
	for i := 0; i < count; i++ {
		persons[i] = generatePerson(int64(i))
	}
	return persons
}

func main() {
	fmt.Println("YAJBE vs JSON vs CBOR Benchmark")
	fmt.Println(strings.Repeat("=", 60))
	
	// Test different dataset sizes
	sizes := []int{10, 100, 1000}
	iterations := 1000
	
	fmt.Printf("%-10s %-8s %-10s %-12s %-8s %-8s\n", 
		"Dataset", "Format", "Operation", "Ops/sec", "Size", "Ratio")
	fmt.Println(strings.Repeat("-", 70))
	
	for _, size := range sizes {
		data := generatePersons(size)
		datasetName := fmt.Sprintf("persons_%d", size)
		
		runBenchmark(datasetName, data, iterations)
		fmt.Println()
	}
}

func runBenchmark(dataset string, data interface{}, iterations int) {
	// JSON Benchmark
	jsonEncoded := benchmarkFormat("JSON", dataset, data, iterations,
		func(d interface{}) ([]byte, error) { return json.Marshal(d) },
		func(data []byte, v interface{}) error { return json.Unmarshal(data, v) })
	
	// CBOR Benchmark  
	cborEncoded := benchmarkFormat("CBOR", dataset, data, iterations,
		func(d interface{}) ([]byte, error) { return cbor.Marshal(d) },
		func(data []byte, v interface{}) error { return cbor.Unmarshal(data, v) })
	
	// Would need YAJBE implementation here
	// yajbeEncoded := benchmarkFormat("YAJBE", dataset, data, iterations,
	//     func(d interface{}) ([]byte, error) { return yajbe.Marshal(d) },
	//     func(data []byte, v interface{}) error { return yajbe.Unmarshal(data, v) })
	
	// Show compression ratios
	jsonSize := len(jsonEncoded)
	cborSize := len(cborEncoded)
	
	fmt.Printf("Compression vs JSON:\n")
	fmt.Printf("  CBOR:  %.2f%% (%.2fx smaller)\n", 
		float64(cborSize)/float64(jsonSize)*100, float64(jsonSize)/float64(cborSize))
}

func benchmarkFormat(format, dataset string, data interface{}, iterations int,
	encoder func(interface{}) ([]byte, error),
	decoder func([]byte, interface{}) error) []byte {
	
	// Encoding benchmark
	start := time.Now()
	var encoded []byte
	var err error
	for i := 0; i < iterations; i++ {
		encoded, err = encoder(data)
		if err != nil {
			panic(err)
		}
	}
	encodeDuration := time.Since(start)
	
	// Decoding benchmark
	start = time.Now()
	for i := 0; i < iterations; i++ {
		var decoded interface{}
		err = decoder(encoded, &decoded)
		if err != nil {
			panic(err)
		}
	}
	decodeDuration := time.Since(start)
	
	encodeOpsPerSec := float64(iterations) / encodeDuration.Seconds()
	decodeOpsPerSec := float64(iterations) / decodeDuration.Seconds()
	
	fmt.Printf("%-10s %-8s %-10s %12.0f %8d %8s\n", 
		dataset, format, "Encode", encodeOpsPerSec, len(encoded), "")
	fmt.Printf("%-10s %-8s %-10s %12.0f %8s %8s\n", 
		"", format, "Decode", decodeOpsPerSec, "", "")
	
	return encoded
}
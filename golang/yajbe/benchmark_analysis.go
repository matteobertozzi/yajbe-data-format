// +build benchmarks

// Benchmark analysis and reporting
package yajbe

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/fxamacker/cbor/v2"
)

// ComprehensiveBenchmarkResult stores detailed benchmark results
type ComprehensiveBenchmarkResult struct {
	Dataset       string
	Format        string
	Operation     string
	Duration      time.Duration
	Throughput    float64 // ops/sec
	OutputSize    int
	Compression   float64 // vs JSON
	AllocsPerOp   int64
	BytesPerOp    int64
}

// RunComprehensiveBenchmark runs a detailed benchmark comparing all formats
func RunComprehensiveBenchmark(t *testing.T) {
	datasets := map[string]interface{}{
		"person_1":     generatePerson(1),
		"persons_10":   generatePersons(10),
		"persons_100":  generatePersons(100),
		"persons_1000": generatePersons(1000),
		"log_1":        generateLogEntry(1),
		"logs_10":      generateLogEntries(10),
		"logs_100":     generateLogEntries(100),
		"logs_1000":    generateLogEntries(1000),
		"nested_small": generateNestedMap(2, 3),
		"nested_med":   generateNestedMap(3, 4),
		"nested_large": generateNestedMap(4, 5),
	}
	
	var results []ComprehensiveBenchmarkResult
	
	for datasetName, data := range datasets {
		// JSON
		jsonResults := benchmarkFormatDetailed("JSON", datasetName, data,
			func(d interface{}) ([]byte, error) { return json.Marshal(d) },
			func(data []byte, v interface{}) error { return json.Unmarshal(data, v) })
		results = append(results, jsonResults...)
		
		// CBOR  
		cborResults := benchmarkFormatDetailed("CBOR", datasetName, data,
			func(d interface{}) ([]byte, error) { return cbor.Marshal(d) },
			func(data []byte, v interface{}) error { return cbor.Unmarshal(data, v) })
		results = append(results, cborResults...)
		
		// YAJBE
		yajbeResults := benchmarkFormatDetailed("YAJBE", datasetName, data,
			func(d interface{}) ([]byte, error) { return Marshal(d) },
			func(data []byte, v interface{}) error { return Unmarshal(data, v) })
		results = append(results, yajbeResults...)
	}
	
	printBenchmarkReport(results)
}

func benchmarkFormatDetailed(format, dataset string, data interface{},
	encoder func(interface{}) ([]byte, error),
	decoder func([]byte, interface{}) error) []ComprehensiveBenchmarkResult {
	
	const iterations = 1000
	var results []ComprehensiveBenchmarkResult
	
	// Encoding benchmark
	start := time.Now()
	var lastEncoded []byte
	var err error
	for i := 0; i < iterations; i++ {
		lastEncoded, err = encoder(data)
		if err != nil {
			panic(err)
		}
	}
	encodeDuration := time.Since(start)
	
	// Get JSON size for compression ratio
	jsonData, _ := json.Marshal(data)
	compression := float64(len(lastEncoded)) / float64(len(jsonData))
	
	results = append(results, ComprehensiveBenchmarkResult{
		Dataset:     dataset,
		Format:      format,
		Operation:   "Encode",
		Duration:    encodeDuration,
		Throughput:  float64(iterations) / encodeDuration.Seconds(),
		OutputSize:  len(lastEncoded),
		Compression: compression,
	})
	
	// Decoding benchmark
	start = time.Now()
	for i := 0; i < iterations; i++ {
		var decoded interface{}
		err = decoder(lastEncoded, &decoded)
		if err != nil {
			panic(err)
		}
	}
	decodeDuration := time.Since(start)
	
	results = append(results, ComprehensiveBenchmarkResult{
		Dataset:     dataset,
		Format:      format,
		Operation:   "Decode",
		Duration:    decodeDuration,
		Throughput:  float64(iterations) / decodeDuration.Seconds(),
		OutputSize:  len(lastEncoded),
		Compression: compression,
	})
	
	return results
}

func printBenchmarkReport(results []ComprehensiveBenchmarkResult) {
	fmt.Print("\n" + strings.Repeat("=", 100) + "\n")
	fmt.Printf("COMPREHENSIVE BENCHMARK REPORT\n")
	fmt.Print(strings.Repeat("=", 100) + "\n\n")
	
	// Group by dataset for easier reading
	datasets := make(map[string][]ComprehensiveBenchmarkResult)
	for _, result := range results {
		datasets[result.Dataset] = append(datasets[result.Dataset], result)
	}
	
	fmt.Printf("%-15s %-6s %-7s %10s %12s %8s %8s\n", 
		"Dataset", "Format", "Op", "Ops/sec", "Duration(µs)", "Size", "Ratio")
	fmt.Print(strings.Repeat("-", 100) + "\n")
	
	for datasetName, datasetResults := range datasets {
		fmt.Printf("\n%s:\n", strings.ToUpper(datasetName))
		
		for _, result := range datasetResults {
			duration := result.Duration.Microseconds() / 1000 // per 1000 ops
			fmt.Printf("%-15s %-6s %-7s %10.0f %12d %8d %8.2f\n",
				"", result.Format, result.Operation, 
				result.Throughput, duration, result.OutputSize, result.Compression)
		}
	}
	
	// Summary statistics
	fmt.Print("\n" + strings.Repeat("-", 100) + "\n")
	fmt.Printf("SUMMARY:\n")
	printSummaryStats(results)
}

func printSummaryStats(results []ComprehensiveBenchmarkResult) {
	// Calculate averages by format and operation
	stats := make(map[string]map[string][]float64)
	
	for _, result := range results {
		if stats[result.Format] == nil {
			stats[result.Format] = make(map[string][]float64)
		}
		stats[result.Format][result.Operation+"_throughput"] = append(
			stats[result.Format][result.Operation+"_throughput"], result.Throughput)
		stats[result.Format][result.Operation+"_compression"] = append(
			stats[result.Format][result.Operation+"_compression"], result.Compression)
	}
	
	fmt.Printf("\nAverage Performance (ops/sec):\n")
	fmt.Printf("%-8s %12s %12s\n", "Format", "Encode", "Decode")
	fmt.Print(strings.Repeat("-", 40) + "\n")
	
	for format := range stats {
		encodeAvg := average(stats[format]["Encode_throughput"])
		decodeAvg := average(stats[format]["Decode_throughput"])
		fmt.Printf("%-8s %12.0f %12.0f\n", format, encodeAvg, decodeAvg)
	}
	
	fmt.Printf("\nAverage Compression Ratio (vs JSON):\n")
	fmt.Printf("%-8s %12s\n", "Format", "Ratio")
	fmt.Print(strings.Repeat("-", 25) + "\n")
	
	for format := range stats {
		compressionAvg := average(stats[format]["Encode_compression"])
		if compressionAvg > 0 {
			fmt.Printf("%-8s %12.2f\n", format, compressionAvg)
		}
	}
}

func average(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

// TestComprehensiveBenchmark runs the detailed benchmark analysis
func TestComprehensiveBenchmark(t *testing.T) {
	fmt.Println("Running comprehensive benchmark analysis...")
	RunComprehensiveBenchmark(t)
}
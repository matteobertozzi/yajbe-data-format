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
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestYajbeDataSets tests encoding/decoding from ../../test-data folder
func TestYajbeDataSets(t *testing.T) {
	bt := NewBaseYajbeTest(t)
	testDataDir := "/Users/th30z/Projects/Mine/yajbe-format/test-data"

	// Find all JSON test files
	jsonFiles, err := findJSONTestFiles(testDataDir)
	require.NoError(t, err, "Failed to find test data files")

	if len(jsonFiles) == 0 {
		t.Skip("No test data files found in " + testDataDir)
	}

	totalTests := 0
	passedTests := 0
	var compressionStats []CompressionStat

	for _, file := range jsonFiles {
		println(file)
		t.Run(filepath.Base(file), func(t *testing.T) {
			totalTests++

			// Load and parse JSON data
			data, err := loadJSONFile(file)
			if err != nil {
				t.Logf("Skipping file %s: %v", file, err)
				return
			}

			// Test YAJBE round-trip
			err = bt.testRoundTrip(data)
			if err != nil {
				t.Errorf("Round-trip failed for %s: %v", file, err)
				return
			}

			// Collect compression statistics
			yajbeSize, jsonSize, cborSize := bt.compareWithOthers(data)
			stat := CompressionStat{
				File:  filepath.Base(file),
				YAJBE: yajbeSize,
				JSON:  jsonSize,
				CBOR:  cborSize,
			}
			compressionStats = append(compressionStats, stat)

			fmt.Printf("Compression: YAJBE=%d bytes, JSON=%d (%.2f%%) bytes CBOR=%d bytes (%.2f%%)",
				yajbeSize, jsonSize, float64(jsonSize)/float64(yajbeSize)*100, cborSize, float64(cborSize)/float64(yajbeSize)*100)

			passedTests++
		})
	}

	// Print summary statistics
	t.Logf("\n=== YAJBE Dataset Test Results ===")
	t.Logf("Total tests: %d, Passed: %d, Failed: %d", totalTests, passedTests, totalTests-passedTests)

	if len(compressionStats) > 0 {
		printCompressionSummary(t, compressionStats)
	}
}

// CompressionStat holds compression statistics for a test file
type CompressionStat struct {
	File  string
	YAJBE int
	JSON  int
	CBOR  int
	Ratio float64
}

// findJSONTestFiles finds all JSON files in the test data directory
func findJSONTestFiles(dir string) ([]string, error) {
	var files []string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip files that can't be accessed
		}

		if !info.IsDir() {
			name := info.Name()
			// Include .json and .json.gz files, but exclude XML files
			if strings.HasSuffix(name, ".json") ||
				(strings.HasSuffix(name, ".json.gz") && !strings.Contains(name, ".xml")) {
				files = append(files, path)
			}
		}

		return nil
	})

	return files, err
}

// loadJSONFile loads and parses a JSON file (supports gzip compression)
func loadJSONFile(filename string) (any, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var reader io.Reader = file

	// Handle gzip compressed files
	if strings.HasSuffix(filename, ".gz") {
		gzReader, err := gzip.NewReader(file)
		if err != nil {
			return nil, err
		}
		defer gzReader.Close()
		reader = gzReader
	}

	var data any
	decoder := json.NewDecoder(reader)
	err = decoder.Decode(&data)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// testRoundTrip tests that data can be encoded and decoded correctly
func (bt *BaseYajbeTest) testRoundTrip(data any) error {
	// Marshal to YAJBE
	yajbeData, err := Marshal(data)
	if err != nil {
		return fmt.Errorf("marshal failed: %w", err)
	}

	// Unmarshal from YAJBE
	var result any
	err = Unmarshal(yajbeData, &result)
	if err != nil {
		return fmt.Errorf("unmarshal failed: %w", err)
	}

	// Verify round-trip by re-encoding both and comparing
	originalJSON, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("original JSON marshal failed: %w", err)
	}

	resultJSON, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("result JSON marshal failed: %w", err)
	}

	if string(originalJSON) != string(resultJSON) {
		return fmt.Errorf("round-trip data mismatch")
	}

	return nil
}

// printCompressionSummary prints a summary of compression statistics
func printCompressionSummary(t *testing.T, stats []CompressionStat) {
	t.Logf("\n=== Compression Statistics ===")

	totalYAJBE := 0
	totalJSON := 0

	for _, stat := range stats {
		totalYAJBE += stat.YAJBE
		totalJSON += stat.JSON
		t.Logf("%-30s: YAJBE=%7d, JSON=%7d, Ratio=%5.1f%%",
			stat.File, stat.YAJBE, stat.JSON, stat.Ratio*100)
	}

	overallRatio := float64(totalYAJBE) / float64(totalJSON)
	compressionPercent := (1.0 - overallRatio) * 100

	t.Logf("%-30s: YAJBE=%7d, JSON=%7d, Ratio=%5.1f%%",
		"TOTAL", totalYAJBE, totalJSON, overallRatio*100)
	t.Logf("Overall compression: %.1f%% smaller than JSON", compressionPercent)

	// Find best and worst compression ratios
	bestRatio := 1.0
	worstRatio := 0.0
	bestFile := ""
	worstFile := ""

	for _, stat := range stats {
		if stat.Ratio < bestRatio {
			bestRatio = stat.Ratio
			bestFile = stat.File
		}
		if stat.Ratio > worstRatio {
			worstRatio = stat.Ratio
			worstFile = stat.File
		}
	}

	t.Logf("Best compression:  %.1f%% (%s)", bestRatio*100, bestFile)
	t.Logf("Worst compression: %.1f%% (%s)", worstRatio*100, worstFile)
}

// TestSpecificDatasets tests specific known datasets
func TestSpecificDatasets(t *testing.T) {
	bt := NewBaseYajbeTest(t)
	testDataDir := "/Users/th30z/Projects/Mine/yajbe-format/test-data"

	// Test specific files that should exist
	specificFiles := []string{
		"data.json",
		"array-same-obj.json",
		"array-same-obj-float.json",
	}

	for _, filename := range specificFiles {
		t.Run(filename, func(t *testing.T) {
			fullPath := filepath.Join(testDataDir, filename)

			// Check if file exists
			if _, err := os.Stat(fullPath); os.IsNotExist(err) {
				t.Skipf("Test file %s not found", fullPath)
				return
			}

			// Load and test the file
			data, err := loadJSONFile(fullPath)
			require.NoError(t, err, "Failed to load test file")

			// Test round-trip
			err = bt.testRoundTrip(data)
			require.NoError(t, err, "Round-trip test failed")

			// Test compression
			yajbeSize, jsonSize, cborSize := bt.compareWithOthers(data)
			t.Logf("File: %s, YAJBE: %d bytes, JSON: %d bytes (%.2f%%), CBOR: %d bytes (%.2f%%)",
				filename, yajbeSize, jsonSize, float64(jsonSize)/float64(yajbeSize)*100, cborSize, float64(cborSize)/float64(yajbeSize)*100)
		})
	}
}

// TestDatasetPerformance tests performance on real datasets
func TestDatasetPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	//bt := NewBaseYajbeTest(t)
	testDataDir := "../../../../test-data"

	// Find a reasonably sized test file
	files, err := findJSONTestFiles(testDataDir)
	require.NoError(t, err)

	var testFile string
	for _, file := range files {
		if strings.Contains(file, "data.json") && !strings.Contains(file, ".gz") {
			testFile = file
			break
		}
	}

	if testFile == "" {
		t.Skip("No suitable test file found for performance test")
	}

	// Load test data
	data, err := loadJSONFile(testFile)
	require.NoError(t, err)

	// Performance test
	t.Run("Marshal", func(t *testing.T) {
		for i := 0; i < 100; i++ {
			_, err := Marshal(data)
			require.NoError(t, err)
		}
	})

	// Test unmarshal performance
	yajbeData, err := Marshal(data)
	require.NoError(t, err)

	t.Run("Unmarshal", func(t *testing.T) {
		for i := 0; i < 100; i++ {
			var result any
			err := Unmarshal(yajbeData, &result)
			require.NoError(t, err)
		}
	})
}

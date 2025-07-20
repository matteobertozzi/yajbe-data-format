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
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testEncodeDecodeFile(t *testing.T, path string, filename string) {
	fullPath := filepath.Join(path, filename)
	
	var rawJSON []byte
	var err error
	
	if strings.HasSuffix(filename, ".json") {
		// Read plain JSON file
		rawJSON, err = os.ReadFile(fullPath)
		require.NoError(t, err)
	} else if strings.HasSuffix(filename, ".json.gz") {
		// Read compressed JSON file
		file, err := os.Open(fullPath)
		require.NoError(t, err)
		defer file.Close()
		
		gzReader, err := gzip.NewReader(file)
		require.NoError(t, err)
		defer gzReader.Close()
		
		rawJSON, err = io.ReadAll(gzReader)
		require.NoError(t, err)
	} else {
		// Skip non-JSON files
		t.Skipf("Skipping non-JSON file: %s", filename)
		return
	}
	
	// Parse and re-encode JSON to normalize it
	startTime := time.Now()
	var obj interface{}
	err = json.Unmarshal(rawJSON, &obj)
	require.NoError(t, err, "Failed to parse JSON from %s", filename)
	
	normalizedJSON, err := json.Marshal(obj)
	require.NoError(t, err)
	jsonElapsed := time.Since(startTime)
	
	// Verify JSON round-trip
	var obj2 interface{}
	err = json.Unmarshal(normalizedJSON, &obj2)
	require.NoError(t, err)
	assert.Equal(t, obj, obj2, "JSON round-trip failed for %s", filename)
	
	t.Logf("%s: JSON decode/encode took %v, size: %d bytes", filename, jsonElapsed, len(normalizedJSON))
	
	// Test YAJBE encoding/decoding
	startTime = time.Now()
	enc, err := Marshal(obj)
	require.NoError(t, err, "Failed to encode YAJBE for %s", filename)
	
	var decoded interface{}
	err = Unmarshal(enc, &decoded)
	require.NoError(t, err, "Failed to decode YAJBE for %s", filename)
	yajbeElapsed := time.Since(startTime)
	
	// Verify YAJBE round-trip (with JSON compatibility adjustments)
	verifyYAJBERoundTrip(t, obj, decoded, filename)
	
	// Calculate SHA-256 hash of encoded data
	hash := sha256.Sum256(enc)
	hashHex := hex.EncodeToString(hash[:])
	
	t.Logf("%s: YAJBE encode/decode took %v, size: %d bytes, SHA256: %s", 
		filename, yajbeElapsed, len(enc), hashHex)
	
	// Log compression ratio
	compressionRatio := float64(len(enc)) / float64(len(normalizedJSON)) * 100
	t.Logf("%s: Compression ratio: %.2f%% (YAJBE %d bytes vs JSON %d bytes)", 
		filename, compressionRatio, len(enc), len(normalizedJSON))
}

// verifyYAJBERoundTrip handles the JSON compatibility issues when comparing objects
func verifyYAJBERoundTrip(t *testing.T, original, decoded interface{}, filename string) {
	// For YAJBE, we need to account for JSON compatibility:
	// - integers decode as float64 when unmarshaling to interface{}
	// - typed slices become []interface{}
	
	// Convert both to JSON and back to normalize types for comparison
	originalJSON, err := json.Marshal(original)
	require.NoError(t, err)
	
	var normalizedOriginal interface{}
	err = json.Unmarshal(originalJSON, &normalizedOriginal)
	require.NoError(t, err)
	
	// For complex nested structures, we compare the JSON representations
	decodedJSON, err := json.Marshal(decoded)
	require.NoError(t, err)
	
	assert.JSONEq(t, string(originalJSON), string(decodedJSON), 
		"YAJBE round-trip failed for %s", filename)
}

func TestDatasetEncodeDecode(t *testing.T) {
	testDataPath := "../test-data"
	
	// Check if test-data directory exists
	if _, err := os.Stat(testDataPath); os.IsNotExist(err) {
		t.Skip("test-data directory not found, skipping dataset tests")
		return
	}
	
	// Read directory entries
	entries, err := os.ReadDir(testDataPath)
	require.NoError(t, err)
	
	processedCount := 0
	for _, entry := range entries {
		if entry.IsDir() {
			continue // Skip subdirectories for now
		}
		
		filename := entry.Name()
		if strings.HasSuffix(filename, ".json") || strings.HasSuffix(filename, ".json.gz") {
			t.Run(filename, func(t *testing.T) {
				testEncodeDecodeFile(t, testDataPath, filename)
			})
			processedCount++
		}
	}
	
	t.Logf("Processed %d JSON dataset files", processedCount)
}

func TestDatasetSpecificFiles(t *testing.T) {
	testDataPath := "../test-data"
	
	// Test specific files that are known to exist and are good test cases
	testFiles := []string{
		"data.json",
		"array-same-obj.json", 
		"array-same-obj-float.json",
	}
	
	for _, filename := range testFiles {
		fullPath := filepath.Join(testDataPath, filename)
		if _, err := os.Stat(fullPath); err == nil {
			t.Run(filename, func(t *testing.T) {
				testEncodeDecodeFile(t, testDataPath, filename)
			})
		} else {
			t.Logf("Skipping %s (file not found)", filename)
		}
	}
}

func TestDatasetPerformanceComparison(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance tests in short mode")
	}
	
	testDataPath := "../test-data"
	filename := "data.json" // Use a simple test file
	
	fullPath := filepath.Join(testDataPath, filename)
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		t.Skip("data.json not found, skipping performance test")
		return
	}
	
	// Read and parse the test data
	rawJSON, err := os.ReadFile(fullPath)
	require.NoError(t, err)
	
	var obj interface{}
	err = json.Unmarshal(rawJSON, &obj)
	require.NoError(t, err)
	
	// Benchmark JSON encoding
	jsonStart := time.Now()
	for i := 0; i < 100; i++ {
		_, err := json.Marshal(obj)
		require.NoError(t, err)
	}
	jsonDuration := time.Since(jsonStart)
	
	// Benchmark YAJBE encoding  
	yajbeStart := time.Now()
	for i := 0; i < 100; i++ {
		_, err := Marshal(obj)
		require.NoError(t, err)
	}
	yajbeDuration := time.Since(yajbeStart)
	
	t.Logf("Performance comparison (100 iterations):")
	t.Logf("  JSON encoding: %v (avg: %v)", jsonDuration, jsonDuration/100)
	t.Logf("  YAJBE encoding: %v (avg: %v)", yajbeDuration, yajbeDuration/100)
	t.Logf("  YAJBE vs JSON ratio: %.2fx", float64(yajbeDuration)/float64(jsonDuration))
}
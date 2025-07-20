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
	"testing"

	"github.com/fxamacker/cbor/v2"
)

// BenchmarkDatasetFiles benchmarks YAJBE vs JSON vs CBOR on real dataset files
func BenchmarkDatasetFiles(b *testing.B) {
	// Load some representative test files
	testFiles := []string{
		"/Users/th30z/Projects/Mine/yajbe-format/test-data/data.json",
		"/Users/th30z/Projects/Mine/yajbe-format/test-data/array-same-obj.json",
		"/Users/th30z/Projects/Mine/yajbe-format/test-data/array-same-obj-float.json",
	}

	// Load the datasets
	datasets := make([]any, 0, len(testFiles))
	datasetNames := make([]string, 0, len(testFiles))

	for _, file := range testFiles {
		data, err := loadJSONFile(file)
		if err != nil {
			b.Logf("Skipping %s: %v", file, err)
			continue
		}
		datasets = append(datasets, data)

		// Extract just the filename for cleaner benchmark names
		name := file[len("/Users/th30z/Projects/Mine/yajbe-format/test-data/"):]
		if len(name) > 20 {
			name = name[:17] + "..."
		}
		datasetNames = append(datasetNames, name)
	}

	if len(datasets) == 0 {
		b.Skip("No dataset files available")
	}

	// Benchmark Marshal operations
	for i, data := range datasets {
		name := datasetNames[i]

		b.Run("Marshal_"+name+"_YAJBE", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, err := Marshal(data)
				if err != nil {
					b.Fatal(err)
				}
			}
		})

		b.Run("Marshal_"+name+"_JSON", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, err := json.Marshal(data)
				if err != nil {
					b.Fatal(err)
				}
			}
		})

		b.Run("Marshal_"+name+"_CBOR", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, err := cbor.Marshal(data)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}

	// Benchmark Unmarshal operations
	for i, data := range datasets {
		name := datasetNames[i]

		// Pre-encode the data for unmarshal benchmarks
		yajbeData, _ := Marshal(data)
		jsonData, _ := json.Marshal(data)
		cborData, _ := cbor.Marshal(data)

		b.Run("Unmarshal_"+name+"_YAJBE", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				var result any
				err := Unmarshal(yajbeData, &result)
				if err != nil {
					b.Fatal(err)
				}
			}
		})

		b.Run("Unmarshal_"+name+"_JSON", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				var result any
				err := json.Unmarshal(jsonData, &result)
				if err != nil {
					b.Fatal(err)
				}
			}
		})

		b.Run("Unmarshal_"+name+"_CBOR", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				var result any
				err := cbor.Unmarshal(cborData, &result)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}

	// Benchmark complete round-trip operations
	for i, data := range datasets {
		name := datasetNames[i]

		b.Run("RoundTrip_"+name+"_YAJBE", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				encoded, err := Marshal(data)
				if err != nil {
					b.Fatal(err)
				}
				var result any
				err = Unmarshal(encoded, &result)
				if err != nil {
					b.Fatal(err)
				}
			}
		})

		b.Run("RoundTrip_"+name+"_JSON", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				encoded, err := json.Marshal(data)
				if err != nil {
					b.Fatal(err)
				}
				var result any
				err = json.Unmarshal(encoded, &result)
				if err != nil {
					b.Fatal(err)
				}
			}
		})

		b.Run("RoundTrip_"+name+"_CBOR", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				encoded, err := cbor.Marshal(data)
				if err != nil {
					b.Fatal(err)
				}
				var result any
				err = cbor.Unmarshal(encoded, &result)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

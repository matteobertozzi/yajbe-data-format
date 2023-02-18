/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package io.github.matteobertozzi.yajbe.examples.bench;

import java.io.File;
import java.io.FileInputStream;
import java.io.IOException;
import java.util.ArrayList;
import java.util.function.Consumer;
import java.util.zip.GZIPInputStream;

import org.openjdk.jmh.annotations.Benchmark;
import org.openjdk.jmh.annotations.BenchmarkMode;
import org.openjdk.jmh.annotations.Mode;
import org.openjdk.jmh.annotations.Param;
import org.openjdk.jmh.annotations.Scope;
import org.openjdk.jmh.annotations.Setup;
import org.openjdk.jmh.annotations.State;
import org.openjdk.jmh.results.format.ResultFormatType;
import org.openjdk.jmh.runner.Runner;
import org.openjdk.jmh.runner.options.OptionsBuilder;

import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.databind.json.JsonMapper;
import com.fasterxml.jackson.dataformat.cbor.databind.CBORMapper;

import io.github.matteobertozzi.yajbe.YajbeMapper;
import io.github.matteobertozzi.yajbe.examples.util.ExamplesUtil;

@State(Scope.Benchmark)
@BenchmarkMode(Mode.Throughput)
public class BenchEncoding {
  private static final ObjectMapper YAJBE_MAPPER = ExamplesUtil.newObjectMapper(new YajbeMapper());
  private static final ObjectMapper JSON_MAPPER = ExamplesUtil.newObjectMapper(new JsonMapper());
  private static final ObjectMapper CBOR_MAPPER = ExamplesUtil.newObjectMapper(new CBORMapper());

  @Param("dataSetName")
  private String dataSetName;

  @Param("format")
  private String format;

  private ObjectMapper mapper;
  private JsonNode inputData;
  private byte[] encodedData;

  @Setup
  public void setup() throws IOException {
    this.mapper = switch (format) {
      case "YAJBE" -> YAJBE_MAPPER;
      case "JSON" -> JSON_MAPPER;
      case "CBOR" -> CBOR_MAPPER;
      default -> throw new IllegalArgumentException("invalid format " + format);
    };

    this.inputData = readDataSetTree(dataSetName);
    this.encodedData = mapper.writeValueAsBytes(inputData);
  }

  private static JsonNode readDataSetTree(final String dataSetPath) throws IOException {
    if (dataSetPath.endsWith(".json.gz")) {
      try (GZIPInputStream stream = new GZIPInputStream(new FileInputStream(dataSetPath))) {
        return JSON_MAPPER.readTree(stream);
      }
    }

    if (dataSetPath.endsWith(".json")) {
      return JSON_MAPPER.readTree(new File(dataSetPath));
    }

    throw new IllegalArgumentException("unsupported file " + dataSetPath);
  }

  @Benchmark
  public byte[] test_encode() throws IOException {
    return mapper.writeValueAsBytes(inputData);
  }

  @Benchmark
  public JsonNode test_decode() throws IOException {
    return mapper.readTree(encodedData);
  }

  @Benchmark
  public JsonNode test_encode_decode() throws IOException {
    final byte[] enc = mapper.writeValueAsBytes(inputData);
    return mapper.readTree(enc);
  }

  @Benchmark
  public JsonNode test_encode_decode_with_gz() throws IOException {
    final byte[] enc = ExamplesUtil.compress(mapper.writeValueAsBytes(inputData));
    return mapper.readTree(ExamplesUtil.decompress(enc));
  }

  private static void foreachTestData(final File rootDir, final Consumer<File> consumer) {
    final File[] files = rootDir.listFiles();
    if (files == null || files.length == 0) return;

    for (final File file: files) {
      if (file.isDirectory()) {
        foreachTestData(file, consumer);
      } else if (file.getName().endsWith(".json")) {
        consumer.accept(file);
      } else if (file.getName().endsWith(".json.gz")) {
        consumer.accept(file);
      }
    }
  }

  public static void main(final String[] args) throws Exception {
    final ArrayList<String> dataSetPaths = new ArrayList<>();
    foreachTestData(new File("../../test-data/"), f -> dataSetPaths.add(f.getAbsolutePath()));

    new Runner(new OptionsBuilder()
      .include(BenchEncoding.class.getSimpleName())
      .param("dataSetName", dataSetPaths.toArray(new String[0]))
      .param("format", "JSON", "CBOR", "YAJBE")
      .result("results.csv")
      .resultFormat(ResultFormatType.CSV)
      .build()
    ).run();
  }
}

/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package io.github.matteobertozzi.yajbe;

import static org.junit.jupiter.api.Assertions.assertArrayEquals;
import static org.junit.jupiter.api.Assertions.assertEquals;

import java.io.File;
import java.io.FileInputStream;
import java.io.IOException;
import java.util.ArrayList;
import java.util.stream.Stream;
import java.util.zip.GZIPInputStream;

import org.junit.jupiter.params.ParameterizedTest;
import org.junit.jupiter.params.provider.MethodSource;

import com.fasterxml.jackson.databind.JsonNode;

public class TestYajbeDataSets extends BaseYajbeTest {
  @ParameterizedTest
  @MethodSource("dataSetPaths")
  public void testEncodeDecode(final File file) throws IOException {
    final JsonNode inputData = parseFile(file);

    // round 1 - encode/decode and check if the decoded data is the same as the input data
    final byte[] enc1 = YAJBE_MAPPER.writeValueAsBytes(inputData);
    final JsonNode dec1 = YAJBE_MAPPER.readValue(enc1, JsonNode.class);
    assertEquals(inputData, dec1);

    // round 2 - encode/decode from the decoded data of the step before. the result should be the same
    final byte[] enc2 = YAJBE_MAPPER.writeValueAsBytes(dec1);
    final JsonNode dec2 = YAJBE_MAPPER.readValue(enc2, JsonNode.class);
    assertArrayEquals(enc1, enc2);
    assertEquals(dec1, dec2);

    // just check the size comparison between JSON and YAJBE
    final byte[] json = JSON_MAPPER.writeValueAsBytes(inputData);
    System.out.printf("test-data %s -> JSON:%s -> YAJBE:%s -> SizeDiff:%s%%%n",
      file, humanSize(json.length), humanSize(enc1.length), Math.round((1.0 - ((double)enc1.length / json.length)) * 100));
  }

  private JsonNode parseFile(final File file) throws IOException {
    if (file.getName().endsWith(".json.gz")) {
      try (GZIPInputStream stream = new GZIPInputStream(new FileInputStream(file))) {
        return JSON_MAPPER.readTree(stream);
      }
    }

    if (file.getName().endsWith(".json")) {
      return JSON_MAPPER.readTree(file);
    }

    throw new IllegalArgumentException("unsupported file " + file);
  }

  private static Stream<File> dataSetPaths() {
    final ArrayList<File> testFiles = new ArrayList<>();
    fetchTestDataSets(testFiles, new File("../../test-data"));
    System.out.println("Datasets: " + testFiles);
    return testFiles.stream();
  }

  private static void fetchTestDataSets(final ArrayList<File> testFiles, final File rootDir) {
    final File[] dirFiles = rootDir.listFiles();
    if (dirFiles == null) return;

    for (final File file : dirFiles) {
      if (file.isDirectory()) {
        fetchTestDataSets(testFiles, file);
      } else {
        final String fileName = file.getName();
        if (fileName.endsWith(".json")) {
          testFiles.add(file);
        } else if (fileName.endsWith(".json.gz")) {
          testFiles.add(file);
        }
      }
    }
  }

  public static String humanSize(final long size) {
    if (size >= (1L << 30)) return String.format("%.2fGiB", (float) size / (1L << 30));
    if (size >= (1L << 20)) return String.format("%.2fMiB", (float) size / (1L << 20));
    if (size >= (1L << 10)) return String.format("%.2fKiB", (float) size / (1L << 10));
    return size > 0 ? size + "bytes" : "0";
  }
}

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

package tech.dnaco.yajbe.examples.bench;

import java.io.File;
import java.io.FileInputStream;
import java.io.IOException;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.ArrayList;
import java.util.zip.GZIPInputStream;

import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.databind.json.JsonMapper;
import com.fasterxml.jackson.dataformat.cbor.databind.CBORMapper;
import com.fasterxml.jackson.dataformat.xml.XmlMapper;

import tech.dnaco.yajbe.YajbeMapper;
import tech.dnaco.yajbe.examples.util.ExamplesUtil;
import tech.dnaco.yajbe.examples.util.HumansTableView;

public class TestDataSetsEncodeDecode {
  private static final ObjectMapper YAJBE_MAPPER = ExamplesUtil.newObjectMapper(new YajbeMapper());
  private static final ObjectMapper JSON_MAPPER = ExamplesUtil.newObjectMapper(new JsonMapper());
  private static final ObjectMapper CBOR_MAPPER = ExamplesUtil.newObjectMapper(new CBORMapper());
  private static final ObjectMapper XML_MAPPER = ExamplesUtil.newObjectMapper(new XmlMapper());

  record TestData (File file, JsonNode node, byte[] jsonEnc, byte[] cborEnc, byte[] yajbeEnc) {}

  @FunctionalInterface
  interface TestDataConsumer {
    void accept(File file, JsonNode node) throws Exception;
  }

  private static void foreachTestData(final File rootDir, final TestDataConsumer consumer) throws Exception {
    final File[] files = rootDir.listFiles();
    if (files == null || files.length == 0) return;

    for (final File file: files) {
      if (file.isDirectory()) {
        foreachTestData(file, consumer);
      } else if (file.getName().endsWith(".json")) {
        consumer.accept(file, JSON_MAPPER.readTree(file));
      } else if (file.getName().endsWith(".json.gz")) {
        consumer.accept(file, parseGzFile(JSON_MAPPER, file));
      } else if (file.getName().endsWith(".xml")) {
        consumer.accept(file, XML_MAPPER.readTree(file));
      } else if (file.getName().endsWith(".xml.gz")) {
        consumer.accept(file, parseGzFile(XML_MAPPER, file));
      }
    }
  }

  private static JsonNode parseGzFile(final ObjectMapper mapper, final File file) throws IOException {
    try (GZIPInputStream stream = new GZIPInputStream(new FileInputStream(file))) {
      return mapper.readTree(stream);
    }
  }

  public static JsonNode encodeDecode(final ObjectMapper mapper, final Object input) throws IOException {
    final byte[] enc = mapper.writeValueAsBytes(input);
    return mapper.readTree(enc);
  }

  public static long runEncodeDecodeBench(final ObjectMapper mapper, final int nruns, final TestData data) throws Exception {
    final String label = mapper.getFactory().getFormatName() + " encode/decode " + data.file().getName();
    return ExamplesUtil.runBench(label, nruns, () -> encodeDecode(mapper, data.node()));
  }

  public static void main(final String[] args) throws Exception {
    final ArrayList<TestData> testData = new ArrayList<>();
    foreachTestData(new File("../../test-data/"), (file, node) -> {
      final byte[] yajbeEnc = YAJBE_MAPPER.writeValueAsBytes(node);
      final byte[] jsonEnc = JSON_MAPPER.writeValueAsBytes(node);
      final byte[] cborEnc = CBOR_MAPPER.writeValueAsBytes(node);
      testData.add(new TestData(file, node, jsonEnc, cborEnc, yajbeEnc));
    });
    testData.sort((a, b) -> Long.compare(b.jsonEnc().length, a.jsonEnc().length));

    final HumansTableView results = new HumansTableView();
    results.addColumn("file", null);
    results.addColumn("JSON Size", v -> ExamplesUtil.humanSize(((Number)v).longValue()));
    results.addColumn("CBOR Size", v -> ExamplesUtil.humanSize(((Number)v).longValue()));
    results.addColumn("Yajbe Size", v -> ExamplesUtil.humanSize(((Number)v).longValue()));
    results.addColumn("JSON ops/sec", v -> ExamplesUtil.humanRate(((Number)v).longValue()));
    results.addColumn("CBOR ops/sec", v -> ExamplesUtil.humanRate(((Number)v).doubleValue()));
    results.addColumn("Yajbe ops/sec", v -> ExamplesUtil.humanRate(((Number)v).doubleValue()));
    results.addColumn("JSON time", v -> ExamplesUtil.humanTimeNanos(((Number)v).longValue()));
    results.addColumn("CBOR time", v -> ExamplesUtil.humanTimeNanos(((Number)v).longValue()));
    results.addColumn("Yajbe time", v -> ExamplesUtil.humanTimeNanos(((Number)v).longValue()));

    for (final TestData data: testData) {
      final int NRUNS = Math.max(10, (int)Math.ceil(2_000_000_000.0 / data.jsonEnc.length));

      final long jsonElapsed = runEncodeDecodeBench(JSON_MAPPER, NRUNS, data);
      final long cborElapsed = runEncodeDecodeBench(CBOR_MAPPER, NRUNS, data);
      final long yajbeElapsed = runEncodeDecodeBench(YAJBE_MAPPER, NRUNS, data);

      results.addRow(data.file().getName(),
        data.jsonEnc().length,
        data.cborEnc().length,
        data.yajbeEnc().length,
        (double)NRUNS / (jsonElapsed / 1000000000.0),
        (double)NRUNS / (cborElapsed / 1000000000.0),
        (double)NRUNS / (yajbeElapsed / 1000000000.0),
        jsonElapsed,
        cborElapsed,
        yajbeElapsed
      );
    }

    System.out.println(results.addHumanView(new StringBuilder()));
    Files.writeString(Path.of("test-datasets.csv"), results.toCsv());
  }
}

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

import java.io.IOException;
import java.util.Arrays;
import java.util.Random;
import java.util.UUID;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.databind.json.JsonMapper;
import com.fasterxml.jackson.dataformat.cbor.databind.CBORMapper;

import tech.dnaco.yajbe.YajbeMapper;
import tech.dnaco.yajbe.examples.util.ExamplesUtil;
import tech.dnaco.yajbe.examples.util.HumansTableView;

public class TestArrayEncodeDecode {
  private static final ObjectMapper YAJBE_MAPPER = ExamplesUtil.newObjectMapper(new YajbeMapper());
  private static final ObjectMapper JSON_MAPPER = ExamplesUtil.newObjectMapper(new JsonMapper());
  private static final ObjectMapper CBOR_MAPPER = ExamplesUtil.newObjectMapper(new CBORMapper());

  public static int[] zeroIntBlock() {
    final int[] block = new int[2 << 20];
    Arrays.fill(block, 0);
    return block;
  }

  public static int[] randInlineIntBlock() {
    final int[] block = new int[2 << 20];
    final Random rand = new Random();
    for (int i = 0; i < block.length; ++i) {
      block[i] = rand.nextInt(-23, 25);
    }
    return block;
  }

  public static int[] randIntBlock() {
    final int[] block = new int[2 << 20];
    final Random rand = new Random();
    for (int i = 0; i < block.length; ++i) {
      final int w = 1 + rand.nextInt(0, 4);
      final int v = rand.nextInt(0, Math.toIntExact(Math.round(Math.pow(2, (w << 3) - 1) - 1)));
      block[i] = rand.nextBoolean() ? v : -v;
    }
    return block;
  }

  private static long[] randLongBlock() {
    final long[] block = new long[1 << 20];
    final Random rand = new Random();
    for (int i = 0; i < block.length; ++i) {
      final int w = 1 + rand.nextInt(0, 8);
      final long v = rand.nextLong(0, Math.round(Math.pow(2, (w << 3) - 1) - 1));
      block[i] = rand.nextBoolean() ? v : -v;
    }
    return block;
  }

  private static String[] randStringBlock() {
    final String[] block = new String[1 << 20];
    for (int i = 0; i < block.length; ++i) {
      block[i] = (UUID.randomUUID() + UUID.randomUUID().toString()).substring(0, 50);
    }
    return block;
  }

  record TestData (String name, Class<?> classOfInput, Object input) {}

  public static Object encodeDecode(final ObjectMapper mapper, final Object input, final Class<?> classOfInput) throws IOException {
    final byte[] enc = mapper.writeValueAsBytes(input);
    return mapper.readValue(enc, classOfInput);
  }


  public static long runEncodeDecodeBench(final ObjectMapper mapper, final int nruns, final TestData data) throws Exception {
    final String label = mapper.getFactory().getFormatName() + " encode/decode " + data.name();
    return ExamplesUtil.runBench(label, nruns, () -> encodeDecode(mapper, data.input(), data.classOfInput()));
  }

  public static void main(final String[] args) throws Exception {
    final TestData[] testData = new TestData[] {
      new TestData("zero[2M]", int[].class, zeroIntBlock()),
      new TestData("inlineInt[2M]", int[].class, randInlineIntBlock()),
      new TestData("int[2M]", int[].class, randIntBlock()),
      new TestData("long[1M]", long[].class, randLongBlock()),
      new TestData("string[1M]", String[].class, randStringBlock()),
    };

    final HumansTableView results = new HumansTableView();
    results.addColumn("test", null);
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
      final int NRUNS = 100;

      final byte[] jsonEnc = JSON_MAPPER.writeValueAsBytes(data.input());
      final byte[] cborEnc = CBOR_MAPPER.writeValueAsBytes(data.input());
      final byte[] yajbeEnc = YAJBE_MAPPER.writeValueAsBytes(data.input());

      final long jsonElapsed = runEncodeDecodeBench(JSON_MAPPER, NRUNS, data);
      final long cborElapsed = runEncodeDecodeBench(CBOR_MAPPER, NRUNS, data);
      final long yajbeElapsed = runEncodeDecodeBench(YAJBE_MAPPER, NRUNS, data);

      results.addRow(data.name(),
        jsonEnc.length,
        cborEnc.length,
        yajbeEnc.length,
        (double)NRUNS / (jsonElapsed / 1000000000.0),
        (double)NRUNS / (cborElapsed / 1000000000.0),
        (double)NRUNS / (yajbeElapsed / 1000000000.0),
        jsonElapsed,
        cborElapsed,
        yajbeElapsed
      );
    }

    System.out.println(results.addHumanView(new StringBuilder()));
  }
}

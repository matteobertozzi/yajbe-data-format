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

import java.nio.file.Files;
import java.nio.file.Path;
import java.util.ArrayList;
import java.util.Arrays;
import java.util.List;
import java.util.Map;
import java.util.Random;
import java.util.UUID;

import io.github.matteobertozzi.yajbe.examples.util.AbstractTestEncodeDecode;
import io.github.matteobertozzi.yajbe.examples.util.HumansTableView;

public class TestArrayEncodeDecode extends AbstractTestEncodeDecode {
  private static final Random rand = new Random();

  public static int[] zeroIntBlock() {
    final int[] block = new int[2 << 20];
    Arrays.fill(block, 0);
    return block;
  }

  public static int[] randInlineIntBlock() {
    final int[] block = new int[2 << 20];
    for (int i = 0; i < block.length; ++i) {
      block[i] = rand.nextInt(-23, 25);
    }
    return block;
  }

  public static int[] randIntBlock() {
    final int[] block = new int[2 << 20];
    for (int i = 0; i < block.length; ++i) {
      final int w = 1 + rand.nextInt(0, 4);
      final int v = rand.nextInt(0, Math.toIntExact(Math.round(Math.pow(2, (w << 3) - 1) - 1)));
      block[i] = rand.nextBoolean() ? v : -v;
    }
    return block;
  }

  private static long[] randLongBlock() {
    final long[] block = new long[1 << 20];
    for (int i = 0; i < block.length; ++i) {
      final int w = 1 + rand.nextInt(0, 8);
      final long v = rand.nextLong(0, Math.round(Math.pow(2, (w << 3) - 1) - 1));
      block[i] = rand.nextBoolean() ? v : -v;
    }
    return block;
  }

  private static float[] randFloatBlock() {
    final float[] block = new float[2 << 20];
    for (int i = 0; i < block.length; ++i) {
      block[i] = rand.nextFloat();
    }
    return block;
  }

  private static double[] randDoubleBlock() {
    final double[] block = new double[1 << 20];
    for (int i = 0; i < block.length; ++i) {
      block[i] = rand.nextDouble();
    }
    return block;
  }

  record DataObject (boolean boolValue, int intValue, long longValue, float floatValue, double doubleValue, String strValue, List<DataObject> items) {}
  private static DataObject[] randDataObjects() {
    final DataObject[] block = new DataObject[128];
    for (int i = 0; i < block.length; ++i) {
      block[i] = randDataObject(0);
    }
    return block;
  }

  private static DataObject randDataObject(final int level) {
    final int nitems = level < 9 && rand.nextBoolean() ? 1 + rand.nextInt(9) : 0;
    final ArrayList<DataObject> items = new ArrayList<>(nitems);
    for (int i = 0; i < nitems; ++i) {
      items.add(randDataObject(level + 1));
    }

    return new DataObject(rand.nextBoolean(),
      rand.nextInt(), rand.nextLong(),
      rand.nextFloat(), rand.nextDouble(),
      "hello",
      (nitems > 0 ? items : List.of())
    );
  }

  private static ArrayList<Map<String, Object>> randMapObjects() {
    final int NITEMS = 128;
    final ArrayList<Map<String, Object>> block = new ArrayList<>(NITEMS);
    for (int i = 0; i < NITEMS; ++i) {
      block.add(randMapObject(0));
    }
    return block;
  }

  private static Map<String, Object> randMapObject(final int level) {
    final int nitems = level < 9 && rand.nextBoolean() ? 1 + rand.nextInt(9) : 0;
    final ArrayList<Map<String, Object>> items = new ArrayList<>(nitems);
    for (int i = 0; i < nitems; ++i) {
      items.add(randMapObject(level + 1));
    }

    return Map.of("boolValue", rand.nextBoolean(),
      "intValue", 0,
      //"longValue", rand.nextLong(),
      //"floatValue", rand.nextFloat(),
      //"doubleValue", rand.nextDouble(),
      //"strValue", "hello",
      "items", (nitems > 0 ? items : List.of())
    );
  }

  private static String[] randStringBlock() {
    final String[] block = new String[1 << 20];
    for (int i = 0; i < block.length; ++i) {
      block[i] = (UUID.randomUUID() + UUID.randomUUID().toString()).substring(0, 50);
    }
    return block;
  }

  public static void main(final String[] args) throws Exception {
    final List<TestData> testData = List.of(
      new TestData("zero[2M]", int[].class, zeroIntBlock()),
      new TestData("inlineInt[2M]", int[].class, randInlineIntBlock()),
      new TestData("int[2M]", int[].class, randIntBlock()),
      new TestData("long[1M]", long[].class, randLongBlock()),
      new TestData("float[2M]", float[].class, randFloatBlock()),
      new TestData("double[1M]", double[].class, randDoubleBlock()),
      new TestData("string[1M]", String[].class, randStringBlock()),
      new TestData("object[1k]", DataObject[].class, randDataObjects()),
      new TestData("map[1k]", Map[].class, randMapObjects())
    );

    final HumansTableView results = runSeparateEncodeDecode(testData);
    //final HumansTableView results = runEncodeDecode(testData);
    System.out.println(results.addHumanView(new StringBuilder()));
    Files.writeString(Path.of("test-array.csv"), results.toCsv());
  }
}

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

import java.io.IOException;
import java.math.BigDecimal;
import java.math.BigInteger;
import java.util.ArrayList;
import java.util.HashSet;
import java.util.HexFormat;
import java.util.List;
import java.util.Set;

import org.junit.jupiter.api.Assertions;
import org.junit.jupiter.api.Test;

public class TestYajbeArrays extends BaseYajbeTest {
  record DataObject (boolean boolValue, int intValue, long longValue, float floatValue, double doubleValue,
    BigInteger bigInt, BigDecimal bigDecimal, String strValue) {}

  @Test
  public void testSimple() throws IOException {
    assertEquals(List.of(), YAJBE_MAPPER.readValue(new byte[] { 0x20 }, List.class));
    assertEquals(List.of(), YAJBE_MAPPER.readValue(new byte[] { 0x2f, 0x01 }, List.class));
    assertArrayEquals(new Object[0], YAJBE_MAPPER.readValue(new byte[] { 0x20 }, Object[].class));
    assertArrayEquals(new Object[0], YAJBE_MAPPER.readValue(new byte[] { 0x2f, 0x01 }, Object[].class));

    assertArrayDecode("2f01", new int[0]);
    assertArrayDecode("2f6001", new int[1]);
    assertArrayDecode("2f606001", new int[2]);
    assertArrayDecode("2f4001", new int[] { 1 });
    assertArrayDecode("2f414101", new int[] { 2, 2 });

    assertArrayEncodeDecode(new int[0], "20");
    assertArrayEncodeDecode(new int[] { 1 }, "2140");
    assertArrayEncodeDecode(new int[] { 2, 2 }, "224141");
    assertArrayEncodeDecode(new int[10], "2a60606060606060606060");
    assertArrayEncodeDecode(new int[11], "2b016060606060606060606060");
    assertArrayEncodeDecode(new int[0xff], "2bf5" + "60".repeat(0xff));
    assertArrayEncodeDecode(new int[265], "2bff" + "60".repeat(265));
    assertArrayEncodeDecode(new int[0xffff], "2cf5ff" + "60".repeat(0xffff));
    assertArrayEncodeDecode(new int[0xffffff], "2df5ffff" + "60".repeat(0xffffff));
    //assertArrayEncodeDecode(new int[0x1fffffff], "2e1fffff" + "60".repeat(0x1fffffff));

    assertArrayEncodeDecode(new String[0], String[].class, "20");
    assertArrayEncodeDecode(new String[1], String[].class, "2100");
    assertArrayEncodeDecode(new String[] { "" }, String[].class, "21c0");
    assertArrayEncodeDecode(new String[] { "a" }, String[].class, "21c161");
    assertArrayEncodeDecode(new String[0xff], String[].class, "2bf5" + "00".repeat(0xff));
    assertArrayEncodeDecode(new String[265], String[].class, "2bff" + "00".repeat(265));
    assertArrayEncodeDecode(new String[0xffff], String[].class, "2cf5ff" + "00".repeat(0xffff));
    assertArrayEncodeDecode(new String[0xffffff], String[].class, "2df5ffff" + "00".repeat(0xffffff));
  }

  @Test
  public void testSmallArrayLength() throws IOException {
    for (int i = 0; i < 10; ++i) {
      final int[] input = new int[i];
      final byte[] enc = YAJBE_MAPPER.writeValueAsBytes(input);
      assertEquals(1 + i, enc.length);
      assertArrayEquals(input, YAJBE_MAPPER.readValue(enc, int[].class));
    }

    for (int i = 11; i <= 265; ++i) {
      final int[] input = new int[i];
      final byte[] enc = YAJBE_MAPPER.writeValueAsBytes(input);
      assertEquals(2 + i, enc.length);
      assertArrayEquals(input, YAJBE_MAPPER.readValue(enc, int[].class));
    }

    for (int i = 266; i <= 0x1fff; ++i) {
      final int[] input = new int[i];
      final byte[] enc = YAJBE_MAPPER.writeValueAsBytes(input);
      assertEquals(3 + i, enc.length);
      assertArrayEquals(input, YAJBE_MAPPER.readValue(enc, int[].class));
    }
  }

  @Test
  public void testRandStringSet() throws IOException {
    final HashSet<String> setData = new HashSet<>();
    for (int i = 0; i < 16; ++i) {
      setData.add("k" + RANDOM.nextLong());

      final byte[] encoded = YAJBE_MAPPER.writeValueAsBytes(setData);
      final String[] decoded = YAJBE_MAPPER.readValue(encoded, String[].class);
      Assertions.assertEquals(setData, Set.of(decoded));

      Assertions.assertArrayEquals(encoded, YAJBE_MAPPER.writeValueAsBytes(decoded));
    }
  }

  @Test
  public void testRandObjectArray() throws IOException {
    final ArrayList<DataObject> data = new ArrayList<>();
    for (int k = 0; k < 32; ++k) {
      data.add(new DataObject(RANDOM.nextBoolean(),
        RANDOM.nextInt(), RANDOM.nextLong(),
        RANDOM.nextFloat(), RANDOM.nextDouble(),
        new BigInteger(RANDOM.nextInt(256), RANDOM),
        new BigDecimal(new BigInteger(RANDOM.nextInt(256), RANDOM), RANDOM.nextInt(0xff)),
        randText(RANDOM.nextInt(0xffff))
      ));

      final byte[] encoded = YAJBE_MAPPER.writeValueAsBytes(data);
      final DataObject[] decoded = YAJBE_MAPPER.readValue(encoded, DataObject[].class);
      Assertions.assertEquals(data, List.of(decoded));

      Assertions.assertArrayEquals(encoded, YAJBE_MAPPER.writeValueAsBytes(decoded));
    }
  }

  public void assertArrayEncodeDecode(final int[] input, final String expectedEnc) throws IOException {
    final byte[] enc = YAJBE_MAPPER.writeValueAsBytes(input);
    assertHexEquals(expectedEnc, enc);
    assertArrayEquals(input, YAJBE_MAPPER.readValue(enc, int[].class));
  }

  public void assertArrayDecode(final String expectedEnc, final int[] input) throws IOException {
    assertArrayEquals(input, YAJBE_MAPPER.readValue(HexFormat.of().parseHex(expectedEnc), int[].class));
  }
}

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

import static org.junit.jupiter.api.Assertions.assertEquals;

import java.io.IOException;

import org.junit.jupiter.api.Test;

public class TestYajbeStrings extends BaseYajbeTest {
  @Test
  public void testSimple() throws IOException {
    assertEncodeDecode("", String.class, "c0");
    assertEncodeDecode("a", String.class, "c161");
    assertEncodeDecode("abc", String.class, "c3616263");
    assertEncodeDecode("x".repeat(59), String.class, "fb" + "78".repeat(59));
    assertEncodeDecode("y".repeat(60), String.class, "fc01" + "79".repeat(60));
    assertEncodeDecode("y".repeat(127), String.class, "fc44" + "79".repeat(127));
    assertEncodeDecode("y".repeat(255), String.class, "fcc4" + "79".repeat(255));
    assertEncodeDecode("z".repeat(0x100), String.class, "fcc5" + "7a".repeat(256));
    assertEncodeDecode("z".repeat(314), String.class, "fcff" + "7a".repeat(314));
    assertEncodeDecode("z".repeat(315), String.class, "fd0001" + "7a".repeat(315));
    assertEncodeDecode("z".repeat(0xffff), String.class, "fdc4ff" + "7a".repeat(0xffff));
    assertEncodeDecode("k".repeat(0xfffff), String.class, "fec4ff0f" + "6b".repeat(0xfffff));
    assertEncodeDecode("k".repeat(0xffffff), String.class, "fec4ffff" + "6b".repeat(0xffffff));
    assertEncodeDecode("k".repeat(0x1000000), String.class, "fec5ffff" + "6b".repeat(0x1000000));
    assertEncodeDecode("k".repeat(0x1000123), String.class, "ffe8000001" + "6b".repeat(0x1000123));
  }

  @Test
  public void testSmallStringLength() throws IOException {
    for (int i = 0; i < 60; ++i) {
      final String input = "x".repeat(i);
      final byte[] enc = YAJBE_MAPPER.writeValueAsBytes(input);
      assertEquals(1 + i, enc.length);
      assertEquals(input, YAJBE_MAPPER.readValue(enc, String.class));
    }

    for (int i = 60; i <= 314; ++i) {
      final String input = "x".repeat(i);
      final byte[] enc = YAJBE_MAPPER.writeValueAsBytes(input);
      assertEquals(2 + i, enc.length);
      assertEquals(input, YAJBE_MAPPER.readValue(enc, String.class));
    }

    for (int i = 315; i <= 0x1fff; ++i) {
      final String input = "x".repeat(i);
      final byte[] enc = YAJBE_MAPPER.writeValueAsBytes(input);
      assertEquals(3 + i, enc.length);
      assertEquals(input, YAJBE_MAPPER.readValue(enc, String.class));
    }
  }

  @Test
  public void testRandEncodeDecode() throws IOException {
    for (int i = 0; i < 100; ++i) {
      final int length = RANDOM.nextInt(1 << 17);
      final String input = randText(length);
      final byte[] enc = YAJBE_MAPPER.writeValueAsBytes(input);
      assertEquals(input, YAJBE_MAPPER.readValue(enc, String.class));
    }
  }
}

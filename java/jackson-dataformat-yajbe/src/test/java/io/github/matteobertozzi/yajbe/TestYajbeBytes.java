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

import org.junit.jupiter.api.Test;

public class TestYajbeBytes extends BaseYajbeTest {
  @Test
  public void testSimple() throws IOException {
    assertEncodeDecode(new byte[0], "80");
    assertEncodeDecode(new byte[1], "8100");
    assertEncodeDecode(new byte[3], "83000000");
    assertEncodeDecode(new byte[59], "bb" + "00".repeat(59));
    assertEncodeDecode(new byte[60], "bc01" + "00".repeat(60));
    assertEncodeDecode(new byte[127], "bc44" + "00".repeat(127));
    assertEncodeDecode(new byte[0xff], "bcc4" + "00".repeat(255));
    assertEncodeDecode(new byte[256], "bcc5" + "00".repeat(256));
    assertEncodeDecode(new byte[314], "bcff" + "00".repeat(314));
    assertEncodeDecode(new byte[315], "bd0001" + "00".repeat(315));
    assertEncodeDecode(new byte[0xffff], "bdc4ff" + "00".repeat(0xffff));
    assertEncodeDecode(new byte[0xfffff], "bec4ff0f" + "00".repeat(0xfffff));
    assertEncodeDecode(new byte[0xffffff], "bec4ffff" + "00".repeat(0xffffff));
    assertEncodeDecode(new byte[0x1000000], "bec5ffff" + "00".repeat(0x1000000));
  }

  @Test
  public void testSmallStringLength() throws IOException {
    for (int i = 0; i < 60; ++i) {
      final byte[] input = new byte[i];
      final byte[] enc = YAJBE_MAPPER.writeValueAsBytes(input);
      assertEquals(1 + i, enc.length);
      assertArrayEquals(input, YAJBE_MAPPER.readValue(enc, byte[].class));
    }

    for (int i = 60; i <= 314; ++i) {
      final byte[] input = new byte[i];
      final byte[] enc = YAJBE_MAPPER.writeValueAsBytes(input);
      assertEquals(2 + i, enc.length);
      assertArrayEquals(input, YAJBE_MAPPER.readValue(enc, byte[].class));
    }

    for (int i = 315; i <= 0x1fff; ++i) {
      final byte[] input = new byte[i];
      final byte[] enc = YAJBE_MAPPER.writeValueAsBytes(input);
      assertEquals(3 + i, enc.length);
      assertArrayEquals(input, YAJBE_MAPPER.readValue(enc, byte[].class));
    }
  }

  @Test
  public void testRandEncodeDecode() throws IOException {
    for (int i = 0; i < 100; ++i) {
      final int length = RANDOM.nextInt(1 << 20);
      final byte[] input = new byte[length];
      RANDOM.nextBytes(input);
      final byte[] enc = YAJBE_MAPPER.writeValueAsBytes(input);
      assertArrayEquals(input, YAJBE_MAPPER.readValue(enc, byte[].class));
    }
  }

  public void assertEncodeDecode(final byte[] input, final String expectedEnc) throws IOException {
    final byte[] enc = YAJBE_MAPPER.writeValueAsBytes(input);
    assertHexEquals(expectedEnc, enc);
    assertArrayEquals(input, YAJBE_MAPPER.readValue(enc, byte[].class));
  }
}

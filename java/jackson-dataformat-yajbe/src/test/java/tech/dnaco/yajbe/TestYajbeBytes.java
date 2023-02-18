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

package tech.dnaco.yajbe;

import static org.junit.jupiter.api.Assertions.assertArrayEquals;

import java.io.IOException;

import org.junit.jupiter.api.Test;

public class TestYajbeBytes extends BaseYajbeTest {
  @Test
  public void testSimple() throws IOException {
    assertEncodeDecode(new byte[0], "80");
    assertEncodeDecode(new byte[1], "8100");
    assertEncodeDecode(new byte[3], "83000000");
    assertEncodeDecode(new byte[59], "bb" + "00".repeat(59));
    assertEncodeDecode(new byte[60], "bc3c" + "00".repeat(60));
    assertEncodeDecode(new byte[127], "bc7f" + "00".repeat(127));
    assertEncodeDecode(new byte[0xff], "bcff" + "00".repeat(255));
    assertEncodeDecode(new byte[0x100], "bd0001" + "00".repeat(256));
    assertEncodeDecode(new byte[0xffff], "bdffff" + "00".repeat(0xffff));
    assertEncodeDecode(new byte[0xfffff], "beffff0f" + "00".repeat(0xfffff));
    assertEncodeDecode(new byte[0xffffff], "beffffff" + "00".repeat(0xffffff));
    assertEncodeDecode(new byte[0x1000000], "bf00000001" + "00".repeat(0x1000000));
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

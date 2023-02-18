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
import static org.junit.jupiter.api.Assertions.assertEquals;

import java.io.IOException;
import java.util.Arrays;

import org.junit.jupiter.api.Test;

public class TestYajbeBool extends BaseYajbeTest {
  @Test
  public void testSimple() throws IOException {
    final byte[] FALSE_BYTES = new byte[] { 0x02 };
    final byte[] TRUE_BYTES = new byte[] { 0x03 };

    assertArrayEquals(FALSE_BYTES, YAJBE_MAPPER.writeValueAsBytes(false));
    assertArrayEquals(TRUE_BYTES, YAJBE_MAPPER.writeValueAsBytes(true));

    assertEquals(false, YAJBE_MAPPER.readValue(FALSE_BYTES, boolean.class));
    assertEquals(true, YAJBE_MAPPER.readValue(TRUE_BYTES, boolean.class));
  }

  @Test
  public void testArray() throws IOException {
    // Array overflow length is 11
    final boolean[] allTrueSmall = new boolean[7];
    Arrays.fill(allTrueSmall, true);
    assertEncodeDecode(allTrueSmall, "2703030303030303");

    final boolean[] allTrueLarge = new boolean[310];
    Arrays.fill(allTrueLarge, true);
    assertEncodeDecode(allTrueLarge, "2c360103030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303");

    final boolean[] allFalseSmall = new boolean[4];
    Arrays.fill(allFalseSmall, false);
    assertEncodeDecode(allFalseSmall, "2402020202");

    final boolean[] allFalseLarge = new boolean[128];
    Arrays.fill(allFalseLarge, false);
    assertEncodeDecode(allFalseLarge, "2b800202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202");

    final boolean[] mixSmall = new boolean[10];
    for (int i = 0; i < mixSmall.length; ++i) mixSmall[i] = (i & 2) == 0;
    assertEncodeDecode(mixSmall, "2a03030202030302020303");

    final boolean[] mixLarge = new boolean[128];
    for (int i = 0; i < mixLarge.length; ++i) mixLarge[i] = (i & 3) == 0;
    assertEncodeDecode(mixLarge, "2b800302020203020202030202020302020203020202030202020302020203020202030202020302020203020202030202020302020203020202030202020302020203020202030202020302020203020202030202020302020203020202030202020302020203020202030202020302020203020202030202020302020203020202");
  }

  @Test
  public void testRandArray() throws IOException {
    for (int k = 0; k < 32; ++k) {
      final int length = RANDOM.nextInt(1 << 20);
      final boolean[] items = new boolean[length];
      for (int i = 0; i < length; ++i) {
        items[i] = RANDOM.nextBoolean();
      }
      final byte[] enc = YAJBE_MAPPER.writeValueAsBytes(items);
      assertArrayEquals(items, YAJBE_MAPPER.readValue(enc, boolean[].class));
    }
  }

  private void assertEncodeDecode(final boolean[] input, final String expectedEnc) throws IOException {
    final byte[] enc = YAJBE_MAPPER.writeValueAsBytes(input);
    assertHexEquals(expectedEnc, enc);
    assertArrayEquals(input, YAJBE_MAPPER.readValue(enc, boolean[].class));
  }
}

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
    assertEncodeDecode("y".repeat(60), String.class, "fc3c" + "79".repeat(60));
    assertEncodeDecode("y".repeat(127), String.class, "fc7f" + "79".repeat(127));
    assertEncodeDecode("y".repeat(0xff), String.class, "fcff" + "79".repeat(255));
    assertEncodeDecode("z".repeat(0x100), String.class, "fd0001" + "7a".repeat(256));
    assertEncodeDecode("z".repeat(0xffff), String.class, "fdffff" + "7a".repeat(0xffff));
    assertEncodeDecode("k".repeat(0xfffff), String.class, "feffff0f" + "6b".repeat(0xfffff));
    assertEncodeDecode("k".repeat(0xffffff), String.class, "feffffff" + "6b".repeat(0xffffff));
    assertEncodeDecode("k".repeat(0x1000000), String.class, "ff00000001" + "6b".repeat(0x1000000));
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

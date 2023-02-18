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
import java.util.HashMap;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;

import org.junit.jupiter.api.Test;

public class TestYajbeMaps extends BaseYajbeTest {
  record DataObject (int a, DataObject obj) {}

  @Test
  public void testSimple() throws IOException {
    assertEquals(Map.of(), YAJBE_MAPPER.readValue(new byte[] { 0x30 }, Map.class));
    assertEquals(Map.of(), YAJBE_MAPPER.readValue(new byte[] { 0x30, 0x01 }, Map.class));

    assertEncodeDecode(Map.of(), Map.class, "3f01");
    assertEncodeDecode(Map.of("a", 1), Map.class, "3f81614001");
    assertEncodeDecode(Map.of("a", "vA"), Map.class, "3f8161c2764101");
    assertEncodeDecode(Map.of("a", List.of(1, 2, 3)), Map.class, "3f81612340414201");
    assertEncodeDecode(Map.of("a", Map.of("l", List.of(1, 2, 3))), Map.class, "3f81613f816c234041420101");
    assertEncodeDecode(Map.of("a", Map.of("l", Map.of("x", 1))), Map.class, "3f81613f816c3f817840010101");

    assertEncodeDecode(new DataObject(1, null), DataObject.class, "3f816140836f626a0001");
    assertEncodeDecode(new DataObject(1, new DataObject(2, null)), DataObject.class, "3f816140836f626a3fa041a1000101");
    assertEncodeDecode(new DataObject(1, new DataObject(2, new DataObject(3, null))), DataObject.class, "3f816140836f626a3fa041a13fa042a100010101");
  }

  @Test
  public void testRand() throws IOException {
    for (int k = 0; k < 32; ++k) {
      final HashMap<String, Object> input = new HashMap<>();
      input.put(generateFieldName(1, 12), RANDOM.nextBoolean());
      input.put(generateFieldName(1, 12), RANDOM.nextInt());
      input.put(generateFieldName(1, 12), RANDOM.nextFloat());
      input.put(generateFieldName(1, 12), randText(32));
      input.put(generateFieldName(1, 12), List.of("1", "2"));
      input.put(generateFieldName(1, 12), Map.of("k", 10, "x", 20));

      final byte[] enc = YAJBE_MAPPER.writeValueAsBytes(input);
      assertEquals(input, YAJBE_MAPPER.readValue(enc, Map.class));
    }
  }

  @Test
  public void testLongMap() throws IOException {
    final LinkedHashMap<String, Integer> input = new LinkedHashMap<>();
    for (int i = 0; i < 0xfff; ++i) {
      input.put("k" + i, i);
      final byte[] enc = YAJBE_MAPPER.writeValueAsBytes(input);
      assertEquals(input, YAJBE_MAPPER.readValue(enc, Map.class));
    }
  }
}

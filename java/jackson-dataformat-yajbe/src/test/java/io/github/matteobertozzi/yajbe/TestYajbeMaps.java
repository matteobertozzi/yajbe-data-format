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
import java.util.HexFormat;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;

import org.junit.jupiter.api.Test;

import com.fasterxml.jackson.databind.cfg.ContextAttributes;

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

    assertDecode("32816140816241", Map.class, Map.of("a", 1, "b", 2));
    assertDecode("33816140816241816342", Map.class, Map.of("a", 1, "b", 2, "c", 3));
    assertDecode("34816140816241816342816443", Map.class, Map.of("a", 1, "b", 2, "c", 3, "d", 4));
    assertDecode("31816123404142", Map.class, Map.of("a", List.of(1, 2, 3)));
    assertDecode("31816131816c23404142", Map.class, Map.of("a", Map.of("l", List.of(1, 2, 3))));
    assertDecode("31816131816c31817840", Map.class, Map.of("a", Map.of("l", Map.of("x", 1))));
  }

  @Test
  public void testStack() throws IOException {
    final LinkedHashMap<String, Object> input = new LinkedHashMap<>();
    input.put("aaa", 1);
    input.put("bbb", Map.of("k", 10));
    input.put("ccc", 2.3);
    input.put("ddd", List.of("a", "b"));
    input.put("eee", List.of("a", Map.of("k", 10), "b"));
    input.put("fff", Map.of("a", Map.of("k", List.of("z", "d"))));
    input.put("ggg", "foo");
    assertEncodeDecode(input, Map.class, "3f8361616140836262623f816b4901836363630666666666666602408364646422c161c1628365656523c1613fa24901c162836666663f81613fa222c17ac164010183676767c3666f6f01");
    assertDecode("3783616161408362626231816b49836363630666666666666602408364646422c161c1628365656523c16131a249c1628366666631816131a222c17ac16483676767c3666f6f", Map.class, input);
  }

  @Test
  public void testProvidedFields() throws IOException {
    final ContextAttributes attrs = ContextAttributes.getEmpty()
      .withSharedAttribute(YajbeMapper.CONFIG_MAP_FIELD_NAMES, new String[] { "hello", "world" });

    final LinkedHashMap<String, Integer> input = new LinkedHashMap<>();
    input.put("world", 2);
    input.put("hello", 1);

    // encode/decode with fields already present in the map. the names will not be in the encoded data
    final byte[] enc = YAJBE_MAPPER.writer(attrs).writeValueAsBytes(input);
    assertEquals("3fa141a04001", HexFormat.of().formatHex(enc));
    final Object dec = YAJBE_MAPPER.reader(attrs).readValue(enc, LinkedHashMap.class);
    assertEquals(input, dec);
    final Object decx = YAJBE_MAPPER.reader(attrs).readValue(HexFormat.of().parseHex("32a141a040"), LinkedHashMap.class);
    assertEquals(input, decx);

    // encode/decode adding a fields not in the base list
    input.put("something new", 3);
    final byte[] enc2 = YAJBE_MAPPER.writer(attrs).writeValueAsBytes(input);
    assertEquals("3fa141a0408d736f6d657468696e67206e65774201", HexFormat.of().formatHex(enc2));
    final Object dec2 = YAJBE_MAPPER.reader(attrs).readValue(enc2, LinkedHashMap.class);
    assertEquals(input, dec2);
    final Object dec2x = YAJBE_MAPPER.reader(attrs).readValue(HexFormat.of().parseHex("33a141a0408d736f6d657468696e67206e657742"), LinkedHashMap.class);
    assertEquals(input, dec2x);
  }

  @Test
  public void testStringLength() throws IOException {
    for (int i = 1; i < 30; ++i) {
      final byte[] x = YAJBE_MAPPER.writeValueAsBytes(Map.of("x".repeat(i), 1));
      assertEquals(3 + i + 1, x.length);
    }
    for (int i = 31; i <= 284; ++i) {
      final byte[] x = YAJBE_MAPPER.writeValueAsBytes(Map.of("x".repeat(i), 1));
      assertEquals(3 + i + 2, x.length);
    }
    for (int i = 285; i < 0xffff; ++i) {
      final byte[] x = YAJBE_MAPPER.writeValueAsBytes(Map.of("x".repeat(i), 1));
      assertEquals(3 + i + 3, x.length);
    }
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

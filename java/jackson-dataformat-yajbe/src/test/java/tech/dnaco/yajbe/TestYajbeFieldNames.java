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

import java.io.ByteArrayOutputStream;
import java.io.IOException;
import java.util.ArrayList;
import java.util.HexFormat;
import java.util.List;

import org.junit.jupiter.api.Test;

public class TestYajbeFieldNames extends BaseYajbeTest {
  @Test
  public void testSimple() throws Exception {
    testEncodeDecode(List.of(
      "aaaaa", "bbbbb", "aaaaa", "aaabb", "aaacc"
    ), "856161616161856262626262a0c2036262c2036363");

    testEncodeDecode(List.of(
      "aaaaa", "aaabbb", "aaaccc", "ddd", "dddeee", "dddffeee"
    ), "856161616161c303626262c30363636383646464c303656565e203036666");

    testEncodeDecode(List.of(
      "1234", "1st_place_medal", "2nd_place_medal", "3rd_place_medal",
      "arrow_backward", "arrow_double_down", "arrow_double_up", "arrow_down",
      "arrow_down_small", "arrow_forward", "arrow_heading_down", "arrow_heading_up",
      "arrow_left", "arrow_lower_left", "arrow_lower_right", "arrow_right",
      "code", "ciqual_food_name_tags", "cities_tags", "codes_tags",
      "1st_place_medal", "2nd_place_medal", "3rd_place_medal"
    ), "84313233348f3173745f706c6163655f6d6564616ce3000c326e64e2000d33728e6172726f775f6261636b77617264cb06646f75626c655f646f776ec20d7570c208776ec60a5f736d616c6cc706666f7277617264cc0668656164696e675f646f776ec20e7570c4066c656674e407056f776572c50c7269676874e0060584636f64659563697175616c5f666f6f645f6e616d655f74616773e4020574696573e201076f64a1a2a3");
  }

  @Test
  public void testRandFieldNames() throws IOException {
    final ArrayList<String> fields = new ArrayList<>(1000);
    for (int i = 0; i < 1000; ++i) {
      fields.add(generateFieldName(3, 32));
    }
    testEncodeDecode(fields, null);
  }

  @Test
  public void testPrefixNames() throws IOException {
    final ArrayList<String> fields = new ArrayList<>(1000);
    for (int i = 0; i < 1000; ++i) {
      fields.add(String.format("key-%03d", i));
    }
    testEncodeDecode(fields, null);
  }

  @Test
  public void testPrefixAndSuffixNames() throws IOException {
    final ArrayList<String> fields = new ArrayList<>(1000);
    for (int i = 0; i < 1000; ++i) {
      fields.add(String.format("key-%03d-foo", i));
    }
    testEncodeDecode(fields, null);
  }

  private static void testEncodeDecode(final List<String> fieldNames, final String expectedHex) throws IOException {
    try (ByteArrayOutputStream baos = new ByteArrayOutputStream()) {
      final YajbeWriterStream writer = new YajbeWriterStream(baos, new byte[128]);
      final YajbeFieldNameWriter fieldsWriter = new YajbeFieldNameWriter(writer);
      for (final String fieldName: fieldNames) {
        fieldsWriter.write(fieldName);
      }
      writer.flush();
      final byte[] data = baos.toByteArray();

      if (expectedHex != null) {
        assertEquals(expectedHex, HexFormat.of().formatHex(data));
      }

      final YajbeReader reader = YajbeReader.fromBytes(data);
      final YajbeFieldNameReader fieldsReader = new YajbeFieldNameReader(reader);
      for (int i = 0; reader.peek() >= 0; ++i) {
        final String field = fieldsReader.read();
        assertEquals(fieldNames.get(i), field);
      }
    }
  }
}

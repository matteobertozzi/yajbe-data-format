/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the 'License'); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an 'AS IS' BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

import { assertEquals } from 'jsr:@std/assert';
import { encodeHex } from 'jsr:@std/encoding';
import { FieldNameReader, FieldNameWriter, InMemoryBytesReader, InMemoryBytesWriter } from './yajbe.ts';

function testFieldNamesEncodeDecode(fieldNames: string[], expectedHex: string): void {
  const writer = new InMemoryBytesWriter();
  const fieldsWriter = new FieldNameWriter(writer, new TextEncoder());
  for (const fieldName of fieldNames) {
    fieldsWriter.encodeString(fieldName);
  }
  writer.flush();
  assertEquals(encodeHex(writer.slice()), expectedHex);

  const reader = new InMemoryBytesReader(writer.slice());
  const fieldsReader = new FieldNameReader(reader, new TextDecoder());
  for (const fieldName of fieldNames) {
    assertEquals(fieldsReader.decodeString(), fieldName);
  }
}

Deno.test("fieldNames.testSimple", () => {
  testFieldNamesEncodeDecode([
    "aaaaa", "bbbbb", "aaaaa", "aaabb", "aaacc"
  ], "856161616161856262626262a0c2036262c2036363");

  testFieldNamesEncodeDecode([
    "aaaaa", "aaabbb", "aaaccc", "ddd", "dddeee", "dddffeee"
  ], "856161616161c303626262c30363636383646464c303656565e203036666");

  testFieldNamesEncodeDecode([
    "1234", "1st_place_medal", "2nd_place_medal", "3rd_place_medal",
    "arrow_backward", "arrow_double_down", "arrow_double_up", "arrow_down",
    "arrow_down_small", "arrow_forward", "arrow_heading_down", "arrow_heading_up",
    "arrow_left", "arrow_lower_left", "arrow_lower_right", "arrow_right",
    "code", "ciqual_food_name_tags", "cities_tags", "codes_tags",
    "1st_place_medal", "2nd_place_medal", "3rd_place_medal"
  ], "84313233348f3173745f706c6163655f6d6564616ce3000c326e64e2000d33728e6172726f775f6261636b77617264cb06646f75626c655f646f776ec20d7570c208776ec60a5f736d616c6cc706666f7277617264cc0668656164696e675f646f776ec20e7570c4066c656674e407056f776572c50c7269676874e0060584636f64659563697175616c5f666f6f645f6e616d655f74616773e4020574696573e201076f64a1a2a3");
});

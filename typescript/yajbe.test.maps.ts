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

import { assertEquals } from 'https://deno.land/std/testing/asserts.ts';
import * as hex from 'https://deno.land/std@0.178.0/encoding/hex.ts';
import * as YAJBE from './yajbe.ts';

function assertEncodeDecode(input: unknown, expectedHex: string) {
  const enc = YAJBE.encode(input);
  assertEquals(new TextDecoder().decode(hex.encode(enc)), expectedHex);
  assertEquals(YAJBE.decode(enc), input);
}

function assertEncode(input: unknown, expectedHex: string) {
  const enc = YAJBE.encode(input);
  assertEquals(new TextDecoder().decode(hex.encode(enc)), expectedHex);
}

function assertDecode(expectedHex: string, input: unknown) {
  const enc = hex.decode(new TextEncoder().encode(expectedHex));
  assertEquals(YAJBE.decode(enc), input);
}


Deno.test("testSimple", () => {
  assertDecode("30", {});
  assertDecode("3f01", {});
  assertDecode("3f81614001", {"a": 1});
  assertDecode("3f8161c2764101", {"a": "vA"});
  assertDecode("3f81612340414201", {"a": [1, 2, 3]});
  assertDecode("3f81613f816c234041420101", {"a": {"l": [1, 2, 3]}});
  assertDecode("3f81613f816c3f817840010101", {"a": {"l": {"x": 1}}});

  assertDecode("3f816140836f626a0001", {a: 1, obj: null});
  assertDecode("3f816140836f626a3fa041a1000101", {a: 1, obj: {a: 2, obj: null}});
  assertDecode("3f816140836f626a3fa041a13fa042a100010101", {a: 1, obj: {a: 2, obj: {a: 3, obj: null}}});

  assertEncodeDecode({"a": 1, "b": 2}, "32816140816241");
  assertEncodeDecode({"a": 1, "b": 2, "c": 3}, "33816140816241816342");
  assertEncodeDecode({"a": 1, "b": 2, "c": 3, "d": 4}, "34816140816241816342816443");
  assertEncodeDecode({"a": [1, 2, 3]}, "31816123404142");
  assertEncodeDecode({"a": {"l": [1, 2, 3]}}, "31816131816c23404142");
  assertEncodeDecode({"a": {"l": {"x": 1}}}, "31816131816c31817840");
});

Deno.test("testStack", () => {
  const input = {
    "aaa": 1,
    "bbb": {"k": 10},
    "ccc": 2.3,
    "ddd": ["a", "b"],
    "eee": ["a", {"k": 10}, "b"],
    "fff": {"a": {"k": ["z", "d"]}},
    "ggg": "foo"
  };
  assertEncodeDecode(input, "3783616161408362626231816b49836363630666666666666602408364646422c161c1628365656523c16131a249c1628366666631816131a222c17ac16483676767c3666f6f");
  assertDecode("3f8361616140836262623f816b4901836363630666666666666602408364646422c161c1628365656523c1613fa24901c162836666663f81613fa222c17ac164010183676767c3666f6f01", input);
});

Deno.test("map.testProvidedFields", () => {
  const INITIAL_FIELDS = ["hello", "world"];
  const options = { fieldNames: INITIAL_FIELDS };

  const input = new Map();
  input.set('world', 2);
  input.set('hello', 1);

  // encode/decode with fields already present in the map. the names will not be in the encoded data
  const obj = Object.fromEntries(input.entries());
  const enc = YAJBE.encode(input, options);
  assertEquals("32a141a040", new TextDecoder().decode(hex.encode(enc)));
  const dec = YAJBE.decode(enc, options);
  assertEquals(obj, dec);
  const decx = YAJBE.decode(hex.decode(new TextEncoder().encode("3fa141a04001")), options);
  assertEquals(obj, decx);

  // encode/decode adding a fields not in the base list
  input.set('something new', 3);
  const obj2 = Object.fromEntries(input.entries());
  const enc2 = YAJBE.encode(input, options);
  assertEquals("33a141a0408d736f6d657468696e67206e657742", new TextDecoder().decode(hex.encode(enc2)));
  const dec2 = YAJBE.decode(enc2, options);
  assertEquals(obj2, dec2);
  const dec2x = YAJBE.decode(hex.decode(new TextEncoder().encode("3fa141a0408d736f6d657468696e67206e65774201")), options);
  assertEquals(obj2, dec2x);
});

Deno.test("testRand", () => {
  for (let k = 0; k < 32; ++k) {
    const input = new Map();
    input.set(generateFieldName(1, 12), Math.random() > 0.5);
    input.set(generateFieldName(1, 12), Math.floor(Math.random() * 9223372036854775807));
    input.set(generateFieldName(1, 12), Math.random());
    input.set(generateFieldName(1, 12), ["1", "2"]);
    input.set(generateFieldName(1, 12), {"k": 10, "x": 20});

    const enc = YAJBE.encode(input);
    assertEquals(Object.fromEntries(input), YAJBE.decode(enc));
  }
});

Deno.test("testLongMap", () => {
  const input: Map<string, number> = new Map();
  for (let i = 0; i < 0x2ff; ++i) {
    input.set("k" + i, i);
    const enc = YAJBE.encode(input);
    assertEquals(Object.fromEntries(input), YAJBE.decode(enc));
  }
});

const FIELD_NAME_CHARS = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ_-";
function generateFieldName(minLength: number, maxLength: number): string {
  const length = minLength + Math.floor(Math.random() * (maxLength - minLength));
  const builder: string[] = [];
  for (let i = 0; i < length; ++i) {
    builder.push(FIELD_NAME_CHARS.charAt(Math.floor(Math.random() * FIELD_NAME_CHARS.length)));
  }
  return builder.join('');
}

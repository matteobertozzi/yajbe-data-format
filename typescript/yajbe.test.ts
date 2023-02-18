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

import { assertEquals, assertAlmostEquals } from "https://deno.land/std/testing/asserts.ts";
import { Buffer } from "https://deno.land/std/io/buffer.ts";

import { FieldNameReader, FieldNameWriter, InMemoryBytesReader, InMemoryBytesWriter } from "./yajbe.ts";
import * as YAJBE from './yajbe.ts';

function encode16(buf: Uint8Array, off?: number, len?: number): string {
  const alphabet = '0123456789abcdef';
  const length = len ?? buf.length;
  const offset = off ?? 0;
  const result: string[] = [];
  for (let i = 0; i < length; ++i) {
    const val = buf[offset + i] & 0xff;
    result.push(alphabet[(val >> 4) & 0xf]);
    result.push(alphabet[val & 0xf]);
  }
  return result.join('');
}

function decode16(hex: string): Uint8Array {
  const buf = new Uint8Array(hex.length / 2);
  let index = 0;
  for (let i = 0; i < hex.length; i += 2) {
    const v = hex.substring(i, i + 2);
    buf[index++] = parseInt(v, 16);
  }
  return buf;
}

function assertEncode(input: unknown, expectedHex: string): Uint8Array {
  const enc = YAJBE.encode(input, {bufSize: expectedHex.length / 2});
  assertEquals(encode16(enc), expectedHex);
  return enc;
}

function assertEncodeDecode(input: unknown, expectedHex: string): void {
  const enc = assertEncode(input, expectedHex);
  const dec = YAJBE.decode(enc);
  assertEquals(dec, input);
}

function assertDecode(hexData: string, expectedObj: unknown): void {
  const dec = YAJBE.decode(decode16(hexData));
  assertEquals(dec, expectedObj);
}

function assertDecodeFloat(hexData: string, expectedObj: number): void {
  const dec: number = YAJBE.decode(decode16(hexData));
  assertAlmostEquals(dec, expectedObj, 0.000001);
}

function assertEncodeDecodeFloat(input: number, expectedHex: string): void {
  const enc = YAJBE.encode(input);
  assertEquals(encode16(enc), expectedHex);
  const dec: number = YAJBE.decode(enc);
  assertAlmostEquals(dec, input, 0.000001);
}

Deno.test("bool.testSimple", () => {
  assertEncodeDecode(false, "02");
  assertEncodeDecode(true, "03");
});

Deno.test("null.testSimple", () => {
  assertEncodeDecode(null, "00");
  assertEncodeDecode([null], "2100");
  assertEncodeDecode([null, null], "220000");
});

Deno.test("int.testSimple", () => {
  // positive ints
  assertEncodeDecode(1, "40");
  assertEncodeDecode(7, "46");
  assertEncodeDecode(24, "57");
  assertEncodeDecode(25, "5819");
  assertEncodeDecode(0xff, "58ff");
  assertEncodeDecode(0xffff, "59ffff");
  assertEncodeDecode(0xffffff, "5affffff");
  assertEncodeDecode(0xffffffff, "5bffffffff");
  assertEncodeDecode(0xffffffffff, "5cffffffffff");
  assertEncodeDecode(0xffffffffffff, "5dffffffffffff");
  assertEncodeDecode(0x1fffffffffffff, "5effffffffffff1f"); // Number.MAX_SAFE_INTEGER
  //assertEncodeDecode(0xffffffffffffff, "5effffffffffffff");
  //assertEncodeDecode(0xfffffffffffffff, "5fffffffffffffff0f");
  //assertEncodeDecode(0x7fffffffffffffff, "5fffffffffffffff7f");

  assertEncodeDecode(100, "5864");
  assertEncodeDecode(1000, "59e803");
  assertEncodeDecode(1000000, "5a40420f");
  assertEncodeDecode(1000000000000, "5c0010a5d4e8");
  assertEncodeDecode(100000000000000, "5d00407a10f35a");

  // negative ints
  assertEncodeDecode(0, "60");
  assertEncodeDecode(-1, "61");
  assertEncodeDecode(-7, "67");
  assertEncodeDecode(-23, "77");
  assertEncodeDecode(-24, "7818");
  assertEncodeDecode(-25, "7819");
  assertEncodeDecode(-0xff, "78ff");
  assertEncodeDecode(-0xffff, "79ffff");
  assertEncodeDecode(-0xffffff, "7affffff");
  assertEncodeDecode(-0xffffffff, "7bffffffff");
  assertEncodeDecode(-0xffffffffff, "7cffffffffff");
  assertEncodeDecode(-0xffffffffffff, "7dffffffffffff");
  assertEncodeDecode(-0x1fffffffffffff, "7effffffffffff1f"); // Number.MIN_SAFE_INTEGER
  //assertEncodeDecode(-0xffffffffffffff, "7effffffffffffff");
  //assertEncodeDecode(-0xfffffffffffffff, "7fffffffffffffff0f");
  //assertEncodeDecode(-0x7fffffffffffffff, "7fffffffffffffff7f");

  assertEncodeDecode(-100, "7864");
  assertEncodeDecode(-1000, "79e803");
  assertEncodeDecode(-1000000, "7a40420f");
  assertEncodeDecode(-1000000000000, "7c0010a5d4e8");
  assertEncodeDecode(-100000000000000, "7d00407a10f35a");
});

Deno.test("float.testSimple", () => {
  assertDecodeFloat("0500000000", 0.0);
  assertDecodeFloat("050000803f", 1.0);
  assertDecodeFloat("05cdcc8c3f", 1.1);
  assertDecodeFloat("050a1101c2", -32.26664);
  assertDecodeFloat("050000807f", Infinity);
  assertDecodeFloat("050000c07f", NaN);
  assertDecodeFloat("05000080ff", -Infinity);

  assertDecodeFloat("060000000000000080", -0.0);
  assertDecodeFloat("0600000000000010c0", -4.0);
  assertEncodeDecodeFloat(-4.1, "0666666666666610c0");
  assertEncodeDecode(1.5, "06000000000000f83f");
  assertEncodeDecode(65504.0, "59e0ff");
  assertDecodeFloat("060000000000fcef40", 65504.0);
  assertEncodeDecode(100000.0, "5aa08601");
  assertDecodeFloat("0600000000006af840", 100000.0);
  assertEncodeDecode(5.960464477539063e-8, "06000000000000703e");
  assertEncodeDecode(0.00006103515625, "06000000000000103f");
  assertEncodeDecode(-5.960464477539063e-8, "0600000000000070be");
  assertEncodeDecode(3.4028234663852886e+38, "06000000e0ffffef47");
  assertEncodeDecode(9007199254740994.0, "060100000000004043");
  assertEncodeDecode(-9007199254740994.0, "0601000000000040c3");
  assertEncodeDecode(1.0e+300, "069c7500883ce4377e");
  assertEncodeDecode(-40.049149, "06c8d0b1834a0644c0");

  //assertEncodeDecode(new BigInteger("-340282366920938463463374607431768211455"), "070400001100ffffffffffffffffffffffffffffffff");
  //assertEncodeDecode(new BigInteger("340282366920938463463374607431768211455"),  "070000001100ffffffffffffffffffffffffffffffff");
  //assertEncodeDecode(new BigDecimal(new BigInteger("-1234567"), -12), "07840c070312d687");
  //assertEncodeDecode(new BigDecimal(new BigInteger("-1234567"), 12), "07040c070312d687");
  //assertEncodeDecode(new BigDecimal(new BigInteger("1234567"), -12), "07800c070312d687");
  //assertEncodeDecode(new BigDecimal(new BigInteger("1234567"), 12), "07000c070312d687");
});

Deno.test("string.testSimple", () => {
  assertEncodeDecode("", "c0");
  assertEncodeDecode("a", "c161");
  assertEncodeDecode("abc", "c3616263");
  assertEncodeDecode("x".repeat(59), "fb" + "78".repeat(59));
  assertEncodeDecode("y".repeat(60), "fc3c" + "79".repeat(60));
  assertEncodeDecode("y".repeat(127), "fc7f" + "79".repeat(127));
  assertEncodeDecode("y".repeat(0xff), "fcff" + "79".repeat(255));
  assertEncodeDecode("z".repeat(0x100), "fd0001" + "7a".repeat(256));
  assertEncodeDecode("z".repeat(0xffff), "fdffff" + "7a".repeat(0xffff));
  assertEncodeDecode("k".repeat(0xfffff), "feffff0f" + "6b".repeat(0xfffff));
  assertEncodeDecode("k".repeat(0xffffff), "feffffff" + "6b".repeat(0xffffff));
  assertEncodeDecode("k".repeat(0x1000000), "ff00000001" + "6b".repeat(0x1000000));
});

Deno.test("bytes.testSimple", () => {
  assertEncodeDecode(new Uint8Array(0), "80");
  assertEncodeDecode(new Uint8Array(1), "8100");
  assertEncodeDecode(new Uint8Array(3), "83000000");
  assertEncodeDecode(new Uint8Array(59), "bb" + "00".repeat(59));
  assertEncodeDecode(new Uint8Array(60), "bc3c" + "00".repeat(60));
  assertEncodeDecode(new Uint8Array(127), "bc7f" + "00".repeat(127));
  assertEncodeDecode(new Uint8Array(0xff), "bcff" + "00".repeat(255));
  assertEncodeDecode(new Uint8Array(0x100), "bd0001" + "00".repeat(256));
  assertEncodeDecode(new Uint8Array(0xffff), "bdffff" + "00".repeat(0xffff));
  assertEncodeDecode(new Uint8Array(0xfffff), "beffff0f" + "00".repeat(0xfffff));
  //assertEncodeDecode(new Uint8Array(0xffffff), "beffffff" + "00".repeat(0xffffff)); // slow
  //assertEncodeDecode(new Uint8Array(0x1000000), "bf00000001" + "00".repeat(0x1000000)); // slow
});

Deno.test("array.testSimple", () => {
  assertDecode("2f01", []);
  assertEncodeDecode([], "20");
  assertEncodeDecode(newIntArray(0, 0), "20");
  assertEncodeDecode(newIntArray(1, 1), "2140");
  assertEncodeDecode(newIntArray(10, 0), "2a60606060606060606060");
  assertEncodeDecode(newIntArray(11, 0), "2b0b6060606060606060606060");
  assertEncodeDecode(newIntArray(0xff, 0), "2bff" + "60".repeat(0xff));
  assertEncodeDecode(newIntArray(0xffff, 0), "2cffff" + "60".repeat(0xffff));
  //assertEncodeDecode(newIntArray(0xffffff, 0), "2dffffff" + "60".repeat(0xffffff)); // slow
});

Deno.test("map.testSimple", () => {
  assertEncodeDecode({}, "30");
  assertEncode(new Map(), "30");
  assertDecode("3f01", {});

  assertEncodeDecode({"a": 1}, "31816140");
  assertEncodeDecode({"a": "vA"}, "318161c27641");
  assertEncodeDecode({"a": [1, 2, 3]}, "31816123404142");
  assertEncodeDecode({"a": {"l": [1, 2, 3]}}, "31816131816c23404142");
  assertEncodeDecode({"a": {"l": {"x": 1}}}, "31816131816c31817840");

  assertDecode("3f81614001", {"a": 1});
  assertDecode("3f8161c2764101", {"a": "vA"});
  assertDecode("3f81612340414201", {"a": [1, 2, 3]});
  assertDecode("3f81613f816c234041420101", {"a": {"l": [1, 2, 3]}});
  assertDecode("3f81613f816c3f817840010101", {"a": {"l": {"x": 1}}});

  assertDecode("3f816140836f626a0001", {a: 1, obj: null});
  assertDecode("3f816140836f626a3fa041a1000101", {a: 1, obj: {a: 2, obj: null}});
  assertDecode("3f816140836f626a3fa041a13fa042a100010101", {a: 1, obj: {a: 2, obj: {a: 3, obj: null}}});
});

function testFieldNamesEncodeDecode(fieldNames: string[], expectedHex: string): void {
  const writer = new InMemoryBytesWriter();
  const fieldsWriter = new FieldNameWriter(writer, new TextEncoder());
  for (const fieldName of fieldNames) {
    fieldsWriter.encodeString(fieldName);
  }
  writer.flush();
  assertEquals(encode16(writer.slice()), expectedHex);

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

Deno.test('data-sets.encodeDecode', async () => {
  const path = '../test-data/';
  for await (const entry of Deno.readDir(path)) {
    let rawJson: string;
    if (entry.name.endsWith('.json')) {
      rawJson = await Deno.readTextFile(path + entry.name);
    } else if (entry.name.endsWith('.json.gz')) {
      const gz = await Deno.readFile(path + entry.name);
      const jsonStream = new Blob([gz]).stream().pipeThrough(new DecompressionStream('gzip'));
      const buffer = new Buffer();
      for await (const chunk of jsonStream) {
        await buffer.write(chunk);
      }
      rawJson = new TextDecoder().decode(buffer.bytes());
    } else {
      continue;
    }

    console.log(entry);
    const obj = JSON.parse(rawJson);
    const enc = YAJBE.encode(obj, {bufSize: rawJson.length});
    const dec = YAJBE.decode(enc);
    assertEquals(obj, dec);
  }
});

function newIntArray(length: number, value: number): Array<number> {
  const array = new Array(length);
  for (let i = 0; i < length; ++i) {
    array[i] = value;
  }
  return array;
}

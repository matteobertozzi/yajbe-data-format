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

import { assertEquals, assertAlmostEquals } from 'https://deno.land/std/testing/asserts.ts';
import * as hex from 'https://deno.land/std@0.178.0/encoding/hex.ts';
import * as YAJBE from './yajbe.ts';

function assertEncodeDecode(input: number, expectedHex: string) {
  const enc = YAJBE.encode(input);
  assertEquals(new TextDecoder().decode(hex.encode(enc)), expectedHex);
  assertAlmostEquals(YAJBE.decode(enc), input);
}

function assertDecode(expectedHex: string, input: number) {
  const enc = hex.decode(new TextEncoder().encode(expectedHex));
  assertAlmostEquals(YAJBE.decode(enc), input);
}

Deno.test("testSimple", () => {
  assertDecode("0500000000", 0.0);
  assertDecode("050000803f", 1.0);
  assertDecode("05cdcc8c3f", 1.1);
  //assertDecode("050a1101c2", -32.26664);
  assertDecode("050000807f", Infinity);
  assertDecode("050000c07f", NaN);
  assertDecode("05000080ff", -Infinity);

  assertDecode("060000000000000080", -0.0);
  assertDecode("0600000000000010c0", -4.0);
  assertDecode("060000000000fcef40", 65504.0);
  assertDecode("0600000000006af840", 100000.0);

  assertEncodeDecode(-4.1, "0666666666666610c0");
  assertEncodeDecode(1.5, "06000000000000f83f");
  assertEncodeDecode(5.960464477539063e-8, "06000000000000703e");
  assertEncodeDecode(0.00006103515625, "06000000000000103f");
  assertEncodeDecode(-5.960464477539063e-8, "0600000000000070be");
  assertEncodeDecode(3.4028234663852886e+38, "06000000e0ffffef47");
  assertEncodeDecode(9007199254740994.0, "060100000000004043");
  assertEncodeDecode(-9007199254740994.0, "0601000000000040c3");
  assertEncodeDecode(1.0e+300, "069c7500883ce4377e");
  assertEncodeDecode(-40.049149, "06c8d0b1834a0644c0");
  assertEncodeDecode(Infinity, "06000000000000f07f");
  assertEncodeDecode(NaN, "06000000000000f87f");
  assertEncodeDecode(-Infinity, "06000000000000f0ff");

  //assertEncodeDecode(new BigInteger("-340282366920938463463374607431768211455"), "070400001100ffffffffffffffffffffffffffffffff");
  //assertEncodeDecode(new BigInteger("340282366920938463463374607431768211455"),  "070000001100ffffffffffffffffffffffffffffffff");
  //assertEncodeDecode(new BigDecimal(new BigInteger("-1234567"), -12), "07840c070312d687");
  //assertEncodeDecode(new BigDecimal(new BigInteger("-1234567"), 12), "07040c070312d687");
  //assertEncodeDecode(new BigDecimal(new BigInteger("1234567"), -12), "07800c070312d687");
  //assertEncodeDecode(new BigDecimal(new BigInteger("1234567"), 12), "07000c070312d687");
});


Deno.test("testRandFloatEncodeDecode", () => {
  for (let i = 0; i < 100; ++i) {
    const input = Math.random() * (1 << 16);
    const enc = YAJBE.encode(input);
    assertAlmostEquals(input, YAJBE.decode(enc));
  }
});

Deno.test("testRandFloatArrayEncodeDecode", () => {
  for (let i = 0; i < 100; ++i) {
    const length = Math.floor(Math.random() * (1 << 14));
    const input: number[] = [];
    for (let i = 0; i < length; ++i) {
      input.push(Math.random() * (1 << 16));
    }
    const enc = YAJBE.encode(input);
    const dec: number[] = YAJBE.decode(enc);
    assertEquals(dec.length, input.length);
    for (let i = 0; i < length; ++i) {
      assertAlmostEquals(dec[i], input[i]);
    }
  }
});
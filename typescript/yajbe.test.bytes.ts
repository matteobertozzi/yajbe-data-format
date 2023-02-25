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

Deno.test('testSimple', () => {
  assertEncodeDecode(new Uint8Array(0), "80");
  assertEncodeDecode(new Uint8Array(1), "8100");
  assertEncodeDecode(new Uint8Array(3), "83000000");
  assertEncodeDecode(new Uint8Array(59), "bb" + "00".repeat(59));
  assertEncodeDecode(new Uint8Array(60), "bc01" + "00".repeat(60));
  assertEncodeDecode(new Uint8Array(127), "bc44" + "00".repeat(127));
  assertEncodeDecode(new Uint8Array(0xff), "bcc4" + "00".repeat(255));
  assertEncodeDecode(new Uint8Array(256), "bcc5" + "00".repeat(256));
  assertEncodeDecode(new Uint8Array(314), "bcff" + "00".repeat(314));
  assertEncodeDecode(new Uint8Array(315), "bd0001" + "00".repeat(315));
  assertEncodeDecode(new Uint8Array(0xffff), "bdc4ff" + "00".repeat(0xffff));
  assertEncodeDecode(new Uint8Array(0xfffff), "bec4ff0f" + "00".repeat(0xfffff));
  //assertEncodeDecode(new Uint8Array(0xffffff), "bec4ffff" + "00".repeat(0xffffff)); // too slow
  //assertEncodeDecode(new Uint8Array(0x1000000), "bec5ffff" + "00".repeat(0x1000000)); // too slow
});

Deno.test('testSmallLength', () => {
  for (let i = 0; i < 60; ++i) {
    const input = new Uint8Array(i);
    const enc = YAJBE.encode(input);
    assertEquals(1 + i, enc.length);
    assertEquals(input, YAJBE.decode(enc));
  }

  for (let i = 60; i <= 314; ++i) {
    const input = new Uint8Array(i);
    const enc = YAJBE.encode(input);
    assertEquals(2 + i, enc.length);
    assertEquals(input, YAJBE.decode(enc));
  }

  for (let i = 315; i <= 0xfff; ++i) {
    const input = new Uint8Array(i);
    const enc = YAJBE.encode(input);
    assertEquals(3 + i, enc.length);
    assertEquals(input, YAJBE.decode(enc));
  }
});

Deno.test('testRandEncodeDecode', () => {
  for (let i = 0; i < 100; ++i) {
    const length = Math.floor(Math.random() * (1 << 16));
    const input = new Uint8Array(length);
    crypto.getRandomValues(input);
    const enc = YAJBE.encode(input);
    assertEquals(input, YAJBE.decode(enc));
  }
});

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

function assertArrayEncodeDecode(input: unknown, expectedHex: string) {
  const enc = YAJBE.encode(input);
  assertEquals(new TextDecoder().decode(hex.encode(enc)), expectedHex);
  assertEquals(YAJBE.decode(enc), input);
}

function assertArrayDecode(expectedHex: string, input: unknown) {
  const enc = hex.decode(new TextEncoder().encode(expectedHex));
  assertEquals(YAJBE.decode(enc), input);
}

function newArray<T>(length: number, defaultValue: T): Array<T> {
  const arr = new Array(length);
  arr.fill(defaultValue);
  return arr;
}

Deno.test('testSimple', () => {
    assertArrayDecode("20", []);
    assertArrayDecode("2f01", []);
    assertArrayDecode("2f6001", [0]);
    assertArrayDecode("2f606001", [0, 0]);
    assertArrayDecode("2f4001", [1]);
    assertArrayDecode("2f414101", [2, 2]);

    assertArrayEncodeDecode([1], "2140");
    assertArrayEncodeDecode([2, 2], "224141");
    assertArrayEncodeDecode(newArray(10, 0), "2a60606060606060606060");
    assertArrayEncodeDecode(newArray(11, 0), "2b016060606060606060606060");
    assertArrayEncodeDecode(newArray(0xff, 0), "2bf5" + "60".repeat(0xff));
    assertArrayEncodeDecode(newArray(265, 0), "2bff" + "60".repeat(265));
    assertArrayEncodeDecode(newArray(0xffff, 0), "2cf5ff" + "60".repeat(0xffff));
    //assertArrayEncodeDecode(newArray(0xffffff, 0), "2df5ffff" + "60".repeat(0xffffff));
    //assertArrayEncodeDecode(newArray(0x1fffffff], "2e1fffff" + "60".repeat(0x1fffffff));

    assertArrayEncodeDecode(["a"], "21c161");
    assertArrayEncodeDecode(newArray(0xff, null), "2bf5" + "00".repeat(0xff));
    assertArrayEncodeDecode(newArray(265, null), "2bff" + "00".repeat(265));
    //assertArrayEncodeDecode(newArray(0xffff, null), "2cf5ff" + "00".repeat(0xffff));
    //assertArrayEncodeDecode(newArray(0xffffff, null), "2df5ffff" + "00".repeat(0xffffff));
});

Deno.test('testSmallLength', () => {
  for (let i = 0; i < 10; ++i) {
    const input = newArray(i, 0);
    const enc = YAJBE.encode(input);
    assertEquals(1 + i, enc.length);
    assertEquals(input, YAJBE.decode(enc));
  }

  for (let i = 11; i <= 265; ++i) {
    const input = newArray(i, 0);
    const enc = YAJBE.encode(input);
    assertEquals(2 + i, enc.length);
    assertEquals(input, YAJBE.decode(enc));
  }

  for (let i = 266; i <= 0xfff; ++i) {
    const input = newArray(i, 0);
    const enc = YAJBE.encode(input);
    assertEquals(3 + i, enc.length);
    assertEquals(input, YAJBE.decode(enc));
  }
});

Deno.test('testRandEncodeDecode', () => {
  for (let i = 0; i < 100; ++i) {
    const length = Math.floor(Math.random() * (1 << 10));
    const input: unknown[] = [];
    for (let i = 0; i < length; ++i) {
      input.push({
        boolValue: Math.random() > 0.5,
        intValue: Math.floor(Math.random() * 2147483648),
        floatValue: Math.random()
      });
    }
    const enc = YAJBE.encode(input);
    assertEquals(input, YAJBE.decode(enc));
  }
});


Deno.test('testRanStringSet', () => {
  const input: Set<string> = new Set();
  for (let i = 0; i < 16; ++i) {
    input.add('k' + Math.floor(Math.random() * 2147483648));

    const enc = YAJBE.encode(input);
    assertEquals(Array.from(input), YAJBE.decode(enc));
  }
});

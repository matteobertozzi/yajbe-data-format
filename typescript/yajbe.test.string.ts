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

const TEXT_CHARS = 'abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ';
function randText(length: number): string {
  const sw: string[] = [];
  for (let i = 0; i < length; ++i) {
    const wordLength = 4 + (Math.floor(Math.random() * 8));
    for (let w = 0; w < wordLength; ++w) {
      sw.push(TEXT_CHARS.charAt(Math.floor(Math.random() * TEXT_CHARS.length)));
    }
    sw.push(' ');
  }
  return sw.join('');
}

Deno.test('testSimple', () => {
  assertEncodeDecode('', 'c0');
  assertEncodeDecode('a', 'c161');
  assertEncodeDecode('abc', 'c3616263');
  assertEncodeDecode('x'.repeat(59), 'fb' + '78'.repeat(59));
  assertEncodeDecode('y'.repeat(60), 'fc01' + '79'.repeat(60));
  assertEncodeDecode('y'.repeat(127), 'fc44' + '79'.repeat(127));
  assertEncodeDecode('y'.repeat(255), 'fcc4' + '79'.repeat(255));
  assertEncodeDecode('z'.repeat(0x100), 'fcc5' + '7a'.repeat(256));
  assertEncodeDecode('z'.repeat(314), 'fcff' + '7a'.repeat(314));
  assertEncodeDecode('z'.repeat(315), 'fd0001' + '7a'.repeat(315));
  assertEncodeDecode('z'.repeat(0xffff), 'fdc4ff' + '7a'.repeat(0xffff));
  assertEncodeDecode('k'.repeat(0xfffff), 'fec4ff0f' + '6b'.repeat(0xfffff));
  assertEncodeDecode('k'.repeat(0xffffff), 'fec4ffff' + '6b'.repeat(0xffffff));
  assertEncodeDecode('k'.repeat(0x1000000), 'fec5ffff' + '6b'.repeat(0x1000000));
  assertEncodeDecode('k'.repeat(0x1000123), 'ffe8000001' + '6b'.repeat(0x1000123));
});

Deno.test('testSmallStringLength', () => {
  for (let i = 0; i < 60; ++i) {
    const input = 'x'.repeat(i);
    const enc = YAJBE.encode(input);
    assertEquals(1 + i, enc.length);
    assertEquals(input, YAJBE.decode(enc));
  }

  for (let i = 60; i <= 314; ++i) {
    const input = 'x'.repeat(i);
    const enc = YAJBE.encode(input);
    assertEquals(2 + i, enc.length);
    assertEquals(input, YAJBE.decode(enc));
  }

  for (let i = 315; i <= 0x1fff; ++i) {
    const input = 'x'.repeat(i);
    const enc = YAJBE.encode(input);
    assertEquals(3 + i, enc.length);
    assertEquals(input, YAJBE.decode(enc));
  }
});

Deno.test('testRandEncodeDecode', () => {
  for (let i = 0; i < 100; ++i) {
    const length = Math.floor(Math.random() * (1 << 17));
    const input = randText(length);
    const enc = YAJBE.encode(input);
    assertEquals(input, YAJBE.decode(enc));
  }
});

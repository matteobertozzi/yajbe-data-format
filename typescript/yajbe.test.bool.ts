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
import * as YAJBE from './yajbe.ts';

function assertEncodeDecode(input: unknown, expectedHex: string) {
  const enc = YAJBE.encode(input);
  assertEquals(encodeHex(enc), expectedHex);
  assertEquals(YAJBE.decode(enc), input);
}

Deno.test("testSimple", () => {
  assertEncodeDecode(false, '02');
  assertEncodeDecode(true, '03');
});

Deno.test("testArray", () => {
  const allTrueSmall: Array<boolean> = new Array(7);
  allTrueSmall.fill(true);
  assertEncodeDecode(allTrueSmall, "2703030303030303");

  const allTrueLarge = new Array(310);
  allTrueLarge.fill(true);
  assertEncodeDecode(allTrueLarge, "2c2c0103030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303");

  const allFalseSmall = new Array(4);
  allFalseSmall.fill(false);
  assertEncodeDecode(allFalseSmall, "2402020202");

  const allFalseLarge = new Array(128);
  allFalseLarge.fill(false);
  assertEncodeDecode(allFalseLarge, "2b760202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202");

  const mixSmall = new Array(10);
  for (let i = 0; i < mixSmall.length; ++i) mixSmall[i] = (i & 2) == 0;
  assertEncodeDecode(mixSmall, "2a03030202030302020303");

  const mixLarge = new Array(128);
  for (let i = 0; i < mixLarge.length; ++i) mixLarge[i] = (i & 3) == 0;
  assertEncodeDecode(mixLarge, "2b760302020203020202030202020302020203020202030202020302020203020202030202020302020203020202030202020302020203020202030202020302020203020202030202020302020203020202030202020302020203020202030202020302020203020202030202020302020203020202030202020302020203020202");
});

Deno.test("testRandArray", () => {
  for (let k = 0; k < 32; ++k) {
    const length = Math.floor(Math.random() * (1 << 16));
    const items = new Array(length);
    for (let i = 0; i < length; ++i) {
      items[i] = Math.random() > 0.5;
    }
    const enc = YAJBE.encode(items);
    assertEquals(items, YAJBE.decode(enc));
  }
});

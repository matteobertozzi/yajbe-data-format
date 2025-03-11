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

function assertEncode(input: unknown, expectedHex: string) {
  const enc = YAJBE.encode(input);
  assertEquals(encodeHex(enc), expectedHex);
}

Deno.test("testSimple", () => {
  assertEncodeDecode(null, '00');
  assertEncode(undefined, '00');
});

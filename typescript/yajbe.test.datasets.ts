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
import { Buffer } from "jsr:@std/io/buffer";
import * as YAJBE from './yajbe.ts';

function hex(digest: ArrayBuffer): string {
  const bytes = new Uint8Array(digest);
  const items: string[] = [];
  for (let i = 0; i < bytes.byteLength; ++i) {
    items.push(bytes[i].toString(16).padStart(2, '0'));
  }
  return items.join('');
}

async function testEncodeDecodeFile(path: string, entry: any, encOpts: any) {
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
    return;
  }

  let startTime = performance.now();
  const obj = JSON.parse(rawJson);
  const json = JSON.stringify(obj);
  let elapsed = performance.now() - startTime;
  assertEquals(obj, JSON.parse(json));
  console.log(entry.name, 'json decode/encode took', elapsed, 'encSize', json.length);

  const fullEncOpts = {bufSize: rawJson.length, ...encOpts};
  startTime = performance.now();
  const enc = YAJBE.encode(obj, fullEncOpts);
  const dec = YAJBE.decode(enc);
  elapsed = performance.now() - startTime;
  assertEquals(obj, dec);

  const digest = await crypto.subtle.digest('sha-256', enc);
  console.log(entry.name, 'yajbe encode/decode took', elapsed, 'encSize', enc.length, hex(digest));
}

Deno.test('data-sets.encodeDecode', async () => {
  const path = '../test-data/';
  const testOpts = [
    {},
    { enumConfig: { type: 'ANY' } },
    { enumConfig: { type: 'ANY', specs: { maxLength: 128 } } },
    { enumConfig: { type: 'LRU', specs: { lruSize: 256, minFreq: 1 } } },
  ];

  for (const encOpts of testOpts) {
    console.log('TEST', encOpts);
    for await (const entry of Deno.readDir(path)) {
      await testEncodeDecodeFile(path, entry, encOpts);
    }
  }
});

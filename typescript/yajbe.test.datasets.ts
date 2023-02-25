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
import { Buffer } from "https://deno.land/std@0.178.0/io/buffer.ts";
import * as YAJBE from './yajbe.ts';

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

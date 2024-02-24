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
import 'package:test/test.dart';

import 'yajbe_test.dart';

void main() {
  group('YAJBE Map Tests', () {
    test('Test Simple', () {
      assertDecode("30", {});
      assertDecode("3f01", {});
      assertDecode("3f81614001", {"a": 1});
      assertDecode("3f8161c2764101", {"a": "vA"});
      assertDecode("3f81612340414201", {"a": [1, 2, 3]});
      assertDecode("3f81613f816c234041420101", {"a": {"l": [1, 2, 3]}});
      assertDecode("3f81613f816c3f817840010101", {"a": {"l": {"x": 1}}});

      assertDecode("3f816140836f626a0001", {'a': 1, 'obj': null});
      assertDecode("3f816140836f626a3fa041a1000101", {'a': 1, 'obj': {'a': 2, 'obj': null}});
      assertDecode("3f816140836f626a3fa041a13fa042a100010101", {'a': 1, 'obj': {'a': 2, 'obj': {'a': 3, 'obj': null}}});

      assertEncodeDecode({"a": 1, "b": 2}, "32816140816241");
      assertEncodeDecode({"a": 1, "b": 2, "c": 3}, "33816140816241816342");
      assertEncodeDecode({"a": 1, "b": 2, "c": 3, "d": 4}, "34816140816241816342816443");
      assertEncodeDecode({"a": [1, 2, 3]}, "31816123404142");
      assertEncodeDecode({"a": {"l": [1, 2, 3]}}, "31816131816c23404142");
      assertEncodeDecode({"a": {"l": {"x": 1}}}, "31816131816c31817840");
    });

    test('test types', () {
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
  });
}

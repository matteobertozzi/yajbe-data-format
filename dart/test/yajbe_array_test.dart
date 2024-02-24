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
import 'dart:math';
import 'dart:typed_data';

import 'package:test/test.dart';
import 'package:yajbe/yajbe.dart';

import 'yajbe_test.dart';

void main() {
  group('YAJBE Array Tests', () {
    test('Test Simple', () {
      assertDecode("20", []);
      assertDecode("2f01", []);
      assertDecode("2f6001", [0]);
      assertDecode("2f606001", [0, 0]);
      assertDecode("2f4001", [1]);
      assertDecode("2f414101", [2, 2]);

      assertEncodeDecode([1], "2140");
      assertEncodeDecode([2, 2], "224141");
      assertEncodeDecode(List.filled(10, 0), "2a60606060606060606060");
      assertEncodeDecode(List.filled(11, 0), "2b016060606060606060606060");
      assertEncodeDecode(List.filled(0xff, 0), "2bf5${"60" * 0xff}");
      assertEncodeDecode(List.filled(265, 0), "2bff${"60" * 265}");
      assertEncodeDecode(List.filled(0xffff, 0), "2cf5ff${"60" * 0xffff}");
      assertEncodeDecode(List.filled(0xffffff, 0), "2df5ffff${"60" * 0xffffff}");

      assertEncodeDecode(["a"], "21c161");
      assertEncodeDecode(List.filled(0xff, null), "2bf5${"00" * 0xff}");
      assertEncodeDecode(List.filled(265, null), "2bff${"00" * 265}");
    });

    test('Test Small Length', () {
      for (int i = 0; i < 10; ++i) {
        var input = List.filled(i, i & 7);
        var enc = yajbeEncode(input);
        expect(1 + i, enc.length);
        expect(input, yajbeDecodeUint8Array(enc));
      }

      for (int i = 11; i <= 265; ++i) {
        var input = List.filled(i, i & 7);
        var enc = yajbeEncode(input);
        expect(2 + i, enc.length);
        expect(input, yajbeDecodeUint8Array(enc));
      }

      for (int i = 266; i <= 0xfff; ++i) {
        var input = List.filled(i, i & 7);
        var enc = yajbeEncode(input);
        expect(3 + i, enc.length);
        expect(input, yajbeDecodeUint8Array(enc));
      }
    });

    test('Test Rand Encode/Decode', () {
      Random rand = Random();
      for (int i = 0; i < 100; ++i) {
        var input = [];
        int length = rand.nextInt(1 << 10);
        for (int i = 0; i < length; ++i) {
          input.add({
            'boolValue': rand.nextDouble() > 0.5,
            'intValue': rand.nextDouble() * 2147483648,
            'floatValue': rand.nextDouble()
          });
        }
        Uint8List enc = yajbeEncode(input);
        expect(input, yajbeDecodeUint8Array(enc));
      }
    });

    test('Test Rand Set', () {
      Random rand = Random();
      Set<String> input = <String>{};
      for (int i = 0; i < 16; ++i) {
        input.add('k${rand.nextInt(2147483648)}');

        var enc = yajbeEncode(input, toEncodable: (Object? x) => x is Set ? List.from(x) : null);
        expect(List.from(input), yajbeDecodeUint8Array(enc));
      }
    });
  });
}

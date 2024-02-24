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
  group('YAJBE Bytes Tests', () {
    test('Test Simple', () {
      assertEncodeDecode(Uint8List(0), "80");
      assertEncodeDecode(Uint8List(1), "8100");
      assertEncodeDecode(Uint8List(3), "83000000");
      assertEncodeDecode(Uint8List(59), "bb${"00" * 59}");
      assertEncodeDecode(Uint8List(60), "bc01${"00" * 60}");
      assertEncodeDecode(Uint8List(127), "bc44${"00" * 127}");
      assertEncodeDecode(Uint8List(0xff), "bcc4${"00" * 255}");
      assertEncodeDecode(Uint8List(256), "bcc5${"00" * 256}");
      assertEncodeDecode(Uint8List(314), "bcff${"00" * 314}");
      assertEncodeDecode(Uint8List(315), "bd0001${"00" * 315}");
      assertEncodeDecode(Uint8List(0xffff), "bdc4ff${"00" * 0xffff}");
      assertEncodeDecode(Uint8List(0xfffff), "bec4ff0f${"00" * 0xfffff}");

      assertEncode(ByteData(0), "80");
      assertEncode(ByteData(1), "8100");
      assertEncode(ByteData(3), "83000000");
      assertEncode(ByteData(59), "bb${"00" * 59}");
      assertEncode(ByteData(60), "bc01${"00" * 60}");
      assertEncode(ByteData(127), "bc44${"00" * 127}");
    });

    test('Test Small Length', () {
      for (int i = 0; i < 60; ++i) {
        Uint8List input = Uint8List(i);
        Uint8List enc = yajbeEncode(input);
        expect(1 + i, enc.length);
        expect(input, yajbeDecodeUint8Array(enc));
      }

      for (int i = 60; i <= 314; ++i) {
        Uint8List input = Uint8List(i);
        Uint8List enc = yajbeEncode(input);
        expect(2 + i, enc.length);
        expect(input, yajbeDecodeUint8Array(enc));
      }

      for (int i = 315; i <= 0xfff; ++i) {
        Uint8List input = Uint8List(i);
        Uint8List enc = yajbeEncode(input);
        expect(3 + i, enc.length);
        expect(input, yajbeDecodeUint8Array(enc));
      }
    });

    test('Test Rand Encode/Decode', () {
      Random rand = Random();
      for (int i = 0; i < 100; ++i) {
        int length = rand.nextInt(1 << 16);
        Uint8List input = Uint8List(length);
        for (int k = 0; k < length; ++k) {
          input[k] = rand.nextInt(0xff);
        }
        Uint8List enc = yajbeEncode(input);
        expect(input, yajbeDecodeUint8Array(enc));
      }
    });
  });
}

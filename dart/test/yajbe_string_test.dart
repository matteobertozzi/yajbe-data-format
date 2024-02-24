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

const textChars = 'abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ';
String randText(int length) {
  Random rand = Random();
  String sw = '';
  for (int i = 0; i < length; ++i) {
    int wordLength = 4 + rand.nextInt(8);
    for (int w = 0; w < wordLength; ++w) {
      sw += textChars[rand.nextInt(textChars.length)];
    }
    sw += ' ';
  }
  return sw;
}

void main() {
  group('YAJBE String Tests', () {
    test('String Test Simple', () {
      assertEncodeDecode('', 'c0');
      assertEncodeDecode('a', 'c161');
      assertEncodeDecode('abc', 'c3616263');
      assertEncodeDecode('x' * 59, 'fb${'78' * 59}');
      assertEncodeDecode('y' * 60, 'fc01${'79' * 60}');
      assertEncodeDecode('y' * 127, 'fc44${'79' * 127}');
      assertEncodeDecode('y' * 255, 'fcc4${'79' * 255}');
      assertEncodeDecode('z' * 0x100, 'fcc5${'7a' * 256}');
      assertEncodeDecode('z' * 314, 'fcff${'7a' * 314}');
      assertEncodeDecode('z' * 315, 'fd0001${'7a' * 315}');
      assertEncodeDecode('z' * 0xffff, 'fdc4ff${'7a' * 0xffff}');
      assertEncodeDecode('k' * 0xfffff, 'fec4ff0f${'6b' * 0xfffff}');
      assertEncodeDecode('k' * 0xffffff, 'fec4ffff${'6b' * 0xffffff}');
      assertEncodeDecode('k' * 0x1000000, 'fec5ffff${'6b' * 0x1000000}');
      assertEncodeDecode('k' * 0x1000123, 'ffe8000001${'6b' * 0x1000123}');
    });

    test('Test Small String', () {
      for (int i = 0; i < 60; ++i) {
        String input = 'x' * i;
        Uint8List enc = yajbeEncode(input);
        expect(1 + i, enc.length);
        expect(input, yajbeDecodeUint8Array(enc));
      }

      for (int i = 60; i <= 314; ++i) {
        String input = 'x' * i;
        Uint8List enc = yajbeEncode(input);
        expect(2 + i, enc.length);
        expect(input, yajbeDecodeUint8Array(enc));
      }

      for (int i = 315; i <= 0x1fff; ++i) {
        String input = 'x' * i;
        Uint8List enc = yajbeEncode(input);
        expect(3 + i, enc.length);
        expect(input, yajbeDecodeUint8Array(enc));
      }
    });
  });

  test('Test Rand String Encode/Decode', () {
    Random rand = Random();
    for (int i = 0; i < 10; ++i) {
      int length = rand.nextInt(1 << 14);
      String input = randText(length);
      Uint8List enc = yajbeEncode(input);
      expect(input, yajbeDecodeUint8Array(enc));
    }
  });
}

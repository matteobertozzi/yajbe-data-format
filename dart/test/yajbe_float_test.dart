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
import 'package:convert/convert.dart';
import 'package:test/test.dart';
import 'package:yajbe/yajbe.dart';

void expectAlmostEquals(a, b) {
  expect((a - b).abs() < 0.000001, true);
}

void assertEncodeDecode(double input, String expectedHex) {
  Uint8List enc = yajbeEncode(input);
  expect(hex.encoder.convert(enc), expectedHex);
  double r = yajbeDecodeUint8Array(enc);
  expectAlmostEquals(r, input);
}

void assertDecode(String expectedHex, double input) {
  var enc = hex.decoder.convert(expectedHex);
  double r = yajbeDecodeUint8Array(Uint8List.fromList(enc));
  expectAlmostEquals(r, input);
}

void main() {
  group('YAJBE Float Tests', () {
    test('Test Simple', () {
      assertDecode("0500000000", 0.0);
      assertDecode("050000803f", 1.0);
      assertDecode("05cdcc8c3f", 1.1);
      assertDecode("050a1101c2", -32.26664);

      assertDecode("060000000000000080", -0.0);
      assertDecode("0600000000000010c0", -4.0);
      assertDecode("060000000000fcef40", 65504.0);
      assertDecode("0600000000006af840", 100000.0);

      assertEncodeDecode(-4.1, "0666666666666610c0");
      assertEncodeDecode(1.5, "06000000000000f83f");
      assertEncodeDecode(5.960464477539063e-8, "06000000000000703e");
      assertEncodeDecode(0.00006103515625, "06000000000000103f");
      assertEncodeDecode(-5.960464477539063e-8, "0600000000000070be");
      assertEncodeDecode(3.4028234663852886e+38, "06000000e0ffffef47");
      assertEncodeDecode(9007199254740994.0, "060100000000004043");
      assertEncodeDecode(-9007199254740994.0, "0601000000000040c3");
      assertEncodeDecode(1.0e+300, "069c7500883ce4377e");
      assertEncodeDecode(-40.049149, "06c8d0b1834a0644c0");
    });

    test("Test Rand Float Encode/Decode", () {
      Random rand = Random();
      for (int i = 0; i < 100; ++i) {
        double input = rand.nextDouble() * (1 << 16);
        Uint8List enc = yajbeEncode(input);
        expectAlmostEquals(input, yajbeDecodeUint8Array(enc));
      }
    });

    test("Test Rand FloatArray Encode/Decode", () {
      Random rand = Random();
      for (int i = 0; i < 100; ++i) {
        int length = rand.nextInt(1 << 14);
        var input = [];
        for (int i = 0; i < length; ++i) {
          input.add(rand.nextDouble() * (1 << 16));
        }
        Uint8List enc = yajbeEncode(input);
        var dec = yajbeDecodeUint8Array(enc);
        expect(dec.length, input.length);
        for (int i = 0; i < length; ++i) {
          expectAlmostEquals(dec[i], input[i]);
        }
      }
    });
  });
}

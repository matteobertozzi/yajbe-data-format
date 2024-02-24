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
  group('YAJBE Bool Tests', () {
    test('Test Simple', () {
      assertEncodeDecode(false, '02');
      assertEncodeDecode(true, '03');
    });


    test("Test Array", () {
      List<bool> allTrueSmall = List.filled(7, true);
      assertEncodeDecode(allTrueSmall, "2703030303030303");

      List<bool> allTrueLarge = List.filled(310, true);
      assertEncodeDecode(allTrueLarge, "2c2c0103030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303");

      List<bool> allFalseSmall = List.filled(4, false);
      assertEncodeDecode(allFalseSmall, "2402020202");

      List<bool> allFalseLarge = List.filled(128, false);
      assertEncodeDecode(allFalseLarge, "2b760202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202");

      List<bool> mixSmall = List.filled(10, false);
      for (int i = 0; i < mixSmall.length; ++i) {
        mixSmall[i] = (i & 2) == 0;
      }
      assertEncodeDecode(mixSmall, "2a03030202030302020303");

      List<bool> mixLarge = List.filled(128, true);
      for (int i = 0; i < mixLarge.length; ++i) {
        mixLarge[i] = (i & 3) == 0;
      }
      assertEncodeDecode(mixLarge, "2b760302020203020202030202020302020203020202030202020302020203020202030202020302020203020202030202020302020203020202030202020302020203020202030202020302020203020202030202020302020203020202030202020302020203020202030202020302020203020202030202020302020203020202");
    });

    test("Test Rand Array", () {
      Random rand = Random();
      for (int k = 0; k < 32; ++k) {
        int length = rand.nextInt(1 << 16);
        List<bool> items = List.filled(length, false);
        for (int i = 0; i < length; ++i) {
          items[i] = rand.nextDouble() > 0.5;
        }
        Uint8List enc = yajbeEncode(items);
        expect(items, yajbeDecodeUint8Array(enc));
      }
    });
  });
}

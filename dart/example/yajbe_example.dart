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
import 'dart:typed_data';

import 'package:yajbe/yajbe.dart';

void main() {

  Uint8List encA = yajbeEncode(0);
  print(yajbeDecodeUint8Array(encA));

  Uint8List encB = yajbeEncode({'a': 0});
  print(yajbeDecodeUint8Array(encB));

  Uint8List encC = yajbeEncode([1, 2, 3]);
  print(yajbeDecodeUint8Array(encC));

  Uint8List encD = yajbeEncode({'a': "hello", 'b': [1, 2, 3]});
  Map decD = yajbeDecodeUint8Array(encD); // {a: "hello", b: [1, 2, 3]}
  print(encD);
  print(decD);
}

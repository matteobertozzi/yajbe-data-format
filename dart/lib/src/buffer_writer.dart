/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

import 'dart:typed_data';

const pow2_8Shifts = [
  1,
  256,
  65536,
  16777216,
  4294967296,
  1099511627776,
  281474976710656,
  72057594037927936
];

class BufferWriter {
  final BytesBuilder _buf;

  BufferWriter() : _buf = BytesBuilder();

  Uint8List takeBytes() {
    return _buf.takeBytes();
  }

  void addByte(int byte) {
    _buf.addByte(byte);
  }

  void add(Uint8List bytes) {
    _buf.add(bytes);
  }

  void addUint(int value, int width) {
    for (int i = 0; i < width; ++i) {
      //_buf.addByte((value >> (i << 3)) & 0xff); // javascript/dart2js does not work well with shifts
      _buf.addByte(value ~/ pow2_8Shifts[i]);
    }
  }
}

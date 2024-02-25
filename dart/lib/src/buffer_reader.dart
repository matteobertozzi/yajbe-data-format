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

const pow2_8Shifts = [1, 256, 65536, 16777216, 4294967296, 1099511627776, 281474976710656, 72057594037927936];

class BufferReader {
  final ByteData data;
  int _position = 0;

  BufferReader(this.data);

  static BufferReader fromUint8Array(Uint8List buf) {
    ByteData data = ByteData(buf.length);
    for (int i = 0; i < buf.length; ++i) {
      data.setUint8(i, buf[i]);
    }
    return BufferReader(data);
  }

  bool get hasRemaining => _position < data.lengthInBytes;

  int peekUint8() {
    return _position < data.lengthInBytes ? data.getUint8(_position) : -1;
  }

  int readUint8() {
    return data.getUint8(_position++);
  }

  Uint8List readUint8Array(int length) {
    Uint8List r = data.buffer.asUint8List(_position, length);
    _position += length;
    return r;
  }

  int readUint(int width) {
    int result = 0;
    for (int i = 0; i < width; ++i) {
      //result |= data.getUint8(_position++) << (i << 3); // javascript/dart2js does not work well with shifts
      result += data.getUint8(_position++) * pow2_8Shifts[i];
    }
    return result;
  }

  double readFloat16() {
    throw UnsupportedError('Not implemented encode float16/vle-float');
  }

  double readFloat32() {
    double v = data.getFloat32(_position, Endian.little);
    _position += 4;
    return v;
  }

  double readFloat64() {
    double v = data.getFloat64(_position, Endian.little);
    _position += 8;
    return v;
  }
}
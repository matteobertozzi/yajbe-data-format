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

import 'dart:convert';
import 'dart:math';
import 'dart:typed_data';

import 'buffer_writer.dart';
import 'field_name_writer.dart';

class YajbeEncoder {
  final BufferWriter _buf;
  late final FieldNameWriter _fieldNameWriter;
  final Object? Function(Object? nonEncodable)? _toEncodable;

  YajbeEncoder(Object? Function(Object? nonEncodable)? toEncodable)
      : _buf = BufferWriter(),
        _toEncodable = toEncodable {
    _fieldNameWriter = FieldNameWriter(_buf);
  }

  Uint8List takeBytes() {
    return _buf.takeBytes();
  }

  void encodeItem(Object? object) {
    if (object is int) {
      encodeInt(object);
    } else if (object is double) {
      encodeDouble(object);
    } else if (object == null) {
      encodeNull();
    } else if (identical(object, true)) {
      encodeTrue();
    } else if (identical(object, false)) {
      encodeFalse();
    } else if (object is String) {
      encodeString(object);
    } else if (object is Uint8List) {
      encodeBytes(object);
    } else if (object is ByteData) {
      encodeBytes(Uint8List.view(object.buffer, object.offsetInBytes));
    } else if (object is List) {
      encodeList(object);
    } else if (object is Map) {
      encodeMap(object);
    } else if (_toEncodable != null) {
      encodeItem(_toEncodable(object));
    } else {
      throw UnimplementedError('unsupported type ${object.runtimeType}');
    }
  }

  void encodeNull() {
    _buf.addByte(0);
  }

  void encodeTrue() {
    _buf.addByte(3);
  }

  void encodeFalse() {
    _buf.addByte(2);
  }

  void encodeInt(int value) {
    if (value > 0) {
      encodePositiveInt(value);
    } else {
      encodeNegativeInt(value);
    }
  }

  void encodePositiveInt(int value) {
    if (value <= 24) {
      _buf.addByte(0x40 | (value - 1));
    } else {
      value -= 25;
      int bytes = _intBytesWidth(value);
      _buf.addByte(0x40 | (23 + bytes));
      _buf.addUint(value, bytes);
    }
  }

  void encodeNegativeInt(int value) {
    value = -value;
    if (value <= 23) {
      _buf.addByte(0x60 | value);
    } else {
      value -= 24;
      int bytes = _intBytesWidth(value);
      _buf.addByte(0x60 | (23 + bytes));
      _buf.addUint(value, bytes);
    }
  }

  void encodeDouble(double value) {
    ByteData fbuf = ByteData(8);
    fbuf.setFloat64(0, value, Endian.little);
    _buf.addByte(6);
    _buf.add(fbuf.buffer.asUint8List());
  }

  void encodeString(String value) {
    Uint8List utf8data = utf8.encode(value);
    _encodeLength(0xc0, 59, utf8data.length);
    _buf.add(utf8data);
  }

  void encodeBytes(Uint8List value) {
    _encodeLength(0x80, 59, value.length);
    _buf.add(value);
  }

  void encodeList(List value) {
    _encodeLength(0x20, 10, value.length);
    for (int i = 0; i < value.length; ++i) {
      encodeItem(value[i]);
    }
  }

  void encodeMap(Map value) {
    List keys = List.from(value.keys);
    keys.sort();

    _encodeLength(0x30, 10, value.length);
    for (int i = 0; i < keys.length; ++i) {
      var key = keys[i];
      if (key is String) {
        _fieldNameWriter.encodeString(key);
      } else {
        _fieldNameWriter.encodeString(key.toString());
      }
      encodeItem(value[key]);
    }
  }

  void _encodeLength(int head, int inlineMax, int length) {
    if (length <= inlineMax) {
      _buf.addByte(head | length);
    } else {
      int deltaLength = length - inlineMax;
      int bytes = _intBytesWidth(deltaLength);
      _buf.addByte(head | (inlineMax + bytes));
      _buf.addUint(deltaLength, bytes);
    }
  }
}

int _intBytesWidth(int value) {
  return max(1, (value.bitLength + 7) >> 3);
}

/// Converts [object] to a YAJBE binary format.
///
/// If value contains objects that are not directly encodable to YAJBE
/// (a value that is not a number, boolean, string, null, list or a map with string keys),
/// the [toEncodable] function is used to convert it to an object that must be directly encodable.
///
/// ```dart
/// const data = {'a': 10, 'b': ['hello', 'world']};
/// Uint8List enc = yajbeEncode(data);
/// ```
Uint8List yajbeEncode(Object? object, {Object? Function(Object? nonEncodable)? toEncodable}) {
  YajbeEncoder encoder = YajbeEncoder(toEncodable);
  encoder.encodeItem(object);
  return encoder.takeBytes();
}

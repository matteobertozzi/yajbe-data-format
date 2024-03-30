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
import 'dart:typed_data';

import 'package:yajbe/src/buffer_reader.dart';

class FieldNameReader {
  final List<Uint8List> _indexedNames;
  final BufferReader _buf;
  Uint8List _lastKey;

  FieldNameReader(BufferReader buf)
      : _buf = buf,
        _lastKey = Uint8List(0),
        _indexedNames = [];

  String decodeString() {
    int head = _buf.readUint8();
    switch ((head >> 5) & 7) {
      case 4:
        return _readFullFieldName(head);
      case 5:
        return _readIndexedFieldName(head);
      case 6:
        return _readPrefix(head);
      case 7:
        return _readPrefixSuffix(head);
      default:
        throw UnsupportedError('unexpected head: $head');
    }
  }

  int _readLength(int head) {
    int length = (head & 0x1f);
    if (length < 30) return length;
    if (length == 30) return _buf.readUint8() + 29;

    int b1 = _buf.readUint8();
    int b2 = _buf.readUint8();
    return 284 + 256 * b1 + b2;
  }

  String _addToIndex(Uint8List utf8data) {
    _indexedNames.add(utf8data);
    _lastKey = utf8data;
    return utf8.decode(utf8data);
  }

  String _readFullFieldName(int head) {
    int length = _readLength(head);
    Uint8List utf8data = _buf.readUint8Array(length);
    return _addToIndex(utf8data);
  }

  String _readIndexedFieldName(int head) {
    int fieldIndex = _readLength(head);
    Uint8List utf8data = _indexedNames[fieldIndex];
    _lastKey = utf8data;
    return utf8.decode(utf8data);
  }

  String _readPrefix(int head) {
    int length = _readLength(head);
    int prefix = _buf.readUint8();
    Uint8List kpart = _buf.readUint8Array(length);
    Uint8List utf8data = Uint8List(prefix + length);
    utf8data.setAll(0, _lastKey.sublist(0, prefix));
    utf8data.setAll(prefix, kpart);
    return _addToIndex(utf8data);
  }

  String _readPrefixSuffix(int head) {
    int length = _readLength(head);
    int prefix = _buf.readUint8();
    int suffix = _buf.readUint8();
    Uint8List kpart = _buf.readUint8Array(length);
    Uint8List utf8data = Uint8List(prefix + length + suffix);
    utf8data.setAll(0, _lastKey.sublist(0, prefix));
    utf8data.setAll(prefix, kpart);
    utf8data.setAll(prefix + length, _lastKey.sublist(_lastKey.length - suffix));
    return _addToIndex(utf8data);
  }
}

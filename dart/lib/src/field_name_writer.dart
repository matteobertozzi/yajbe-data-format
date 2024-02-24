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

class FieldNameWriter {
  final Map<String, int> _indexedMap;
  final BufferWriter _buf;
  Uint8List _lastKey;

  FieldNameWriter(BufferWriter buf)
    : _buf = buf,
      _indexedMap = {},
      _lastKey = Uint8List(0);

  void encodeString(String key) {
    Uint8List utf8data = utf8.encode(key);

    int? index = _indexedMap[key];
    if (index != null) {
      _writeIndexedFieldName(index);
      _lastKey = utf8data;
      return;
    }

    if (_lastKey.isNotEmpty && utf8data.length > 4) {
      int prefix = min(0xff, _prefix(utf8data));
      int suffix = _suffix(utf8data, prefix);

      if (suffix > 2) {
        _writePrefixSuffix(utf8data, prefix, min(0xff, suffix));
      } else if (prefix > 2) {
        _writePrefix(utf8data, prefix);
      } else {
        _writeFullFieldName(utf8data);
      }
    } else {
      _writeFullFieldName(utf8data);
    }

    if (_indexedMap.length < 65819) {
      _indexedMap[key] = _indexedMap.length;
    }
    _lastKey = utf8data;
  }

  void _writeFullFieldName(Uint8List fieldName) {
    // 100----- Full Field Name (0-29 length - 1, 30 1b-len, 31 2b-len)
    _writeLength(0x80, fieldName.length);
    _buf.add(fieldName);
  }

  void _writeIndexedFieldName(int fieldIndex) {
    // 101----- Field Offset (0-29 field, 30 1b-len, 31 2b-len)
    _writeLength(0xa0, fieldIndex);
  }

  void _writePrefix(Uint8List fieldName, int prefix) {
    // 110----- Prefix (1byte prefix, 0-29 length - 1, 30 1b-len, 31 2b-len)
    int length = fieldName.length - prefix;
    _writeLength(0xc0, length);
    _buf.addByte(prefix);
    _buf.add(fieldName.sublist(prefix));
  }

  void _writePrefixSuffix(Uint8List fieldName, int prefix, int suffix) {
    // 111----- Prefix/Suffix (1byte prefix, 1byte suffix, 0-29 length - 1, 30 1b-len, 31 2b-len)
    int length = fieldName.length - prefix - suffix;
    _writeLength(0xe0, length);
    _buf.addByte(prefix);
    _buf.addByte(suffix);
    _buf.add(fieldName.sublist(prefix, fieldName.length - suffix));
  }

  void _writeLength(int head, int length) {
    if (length < 30) {
      _buf.addByte(head | length);
    } else if (length <= 284) {
      _buf.addByte(head | 0x1e);
      _buf.addByte((length - 29) & 0xff);
    } else if (length <= 65819) {
      _buf.addByte(head | 0x1f);
      _buf.addByte((length - 284) ~/ 256);
      _buf.addByte((length - 284) & 255);
    } else {
      throw UnsupportedError("unexpected too many field names: $length");
    }
  }

  int _prefix(Uint8List key) {
    Uint8List a = _lastKey;
    Uint8List b = key;
    int len = min(a.length, b.length);
    for (int i = 0; i < len; ++i) {
      if (a[i] != b[i]) {
        return i;
      }
    }
    return len;
  }

  int _suffix(Uint8List key, int kPrefix) {
    Uint8List a = _lastKey;
    Uint8List b = key;
    int bLen = b.length - kPrefix;
    int len = min(a.length, bLen);
    for (int i = 1; i <= len; ++i) {
      if ((a[a.length - i] & 0xff) != (b[kPrefix + (bLen - i)] & 0xff)) {
        return i - 1;
      }
    }
    return len;
  }
}

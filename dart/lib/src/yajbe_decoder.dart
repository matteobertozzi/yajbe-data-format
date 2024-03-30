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

import 'field_name_reader.dart';
import 'buffer_reader.dart';

class YajbeDecoder {
  final FieldNameReader _fieldNameReader;
  final BufferReader _buffer;

  YajbeDecoder(BufferReader buffer)
      : _fieldNameReader = FieldNameReader(buffer),
        _buffer = buffer;

  dynamic decodeItem() {
    while (true) {
      int head = _buffer.readUint8();
      if ((head & 0xc0) == 0xc0) {
        return decodeString(head);
      } else if ((head & 0x80) == 0x80) {
        return decodeBytes(head);
      } else if ((head & 0x40) == 0x40) {
        return decodeInt(head);
      } else if ((head & 0x30) == 0x30) {
        return decodeObject(head);
      } else if ((head & 0x20) == 0x20) {
        return decodeArray(head);
      } else if ((head & 0x8) == 0x8) {
        throw UnsupportedError('unsupported enum item head: $head');
      } else if ((head & 0x4) == 0x4) {
        return decodeFloat(head);
      } else {
        switch (head) {
          // null
          case 0:
            return null;
          // boolean
          case 2:
            return false;
          case 3:
            return true;
          default:
            throw UnsupportedError('unsupported item head: $head');
        }
      }
    }
  }

  int decodeInt(int head) {
    bool signed = (head & 0x60) == 0x60;

    int w = head & 0x1f;
    if (w < 24) {
      return signed ? ((w != 0) ? -w : 0) : (1 + w);
    }

    int value = _buffer.readUint(w - 23);
    return signed ? -(value + 24) : (value + 25);
  }

  double decodeFloat(int head) {
    switch (head & 3) {
      case 0:
        return _buffer.readFloat16();
      case 1:
        return _buffer.readFloat32();
      case 2:
        return _buffer.readFloat64();
      case 3:
        throw UnsupportedError('decode bigdecimal');
    }
    return 0;
  }

  int readBytesLength(int head) {
    int w = head & 0x3f;
    if (w <= 59) return w;
    return 59 + _buffer.readUint(w - 59);
  }

  Uint8List decodeBytes(int head) {
    int length = readBytesLength(head);
    return _buffer.readUint8Array(length);
  }

  String decodeString(int head) {
    Uint8List utf8data = decodeBytes(head);
    String text = utf8.decode(utf8data);
    //this.enumMapping?.add(text);
    return text;
  }

  bool readHasMore() {
    if (_buffer.peekUint8() != 1) {
      return true;
    }
    _buffer.readUint8();
    return false;
  }

  int readItemCount(int w) {
    if (w <= 10) return w;
    return 10 + _buffer.readUint(w - 10);
  }

  List decodeArray(int head) {
    int w = head & 0xf;
    if (w == 0xf) {
      List retArray = [];
      while (readHasMore()) {
        retArray.add(decodeItem());
      }
      return retArray;
    }

    int length = readItemCount(w);
    List retArray = [];
    for (int i = 0; i < length; ++i) {
      retArray.add(decodeItem());
    }
    return retArray;
  }

  Map<String, dynamic> decodeObject(int head) {
    int w = head & 0xf;
    if (w == 0xf) {
      Map<String, dynamic> retObject = {};
      while (readHasMore()) {
        String key = _fieldNameReader.decodeString();
        retObject[key] = decodeItem();
      }
      return retObject;
    }

    int length = readItemCount(w);
    Map<String, dynamic> retObject = {};
    for (int i = 0; i < length; ++i) {
      String key = _fieldNameReader.decodeString();
      retObject[key] = decodeItem();
    }
    return retObject;
  }
}

/// Parses the YAJBE encoded data and returns the resulting "Json" object.
dynamic yajbeDecode(BufferReader buffer) {
  YajbeDecoder decoder = YajbeDecoder(buffer);
  return decoder.decodeItem();
}

/// Parses the YAJBE Uint8List encoded data and returns the resulting "Json" object.
///
/// ```dart
/// const data = {'a': 10, 'b': ['hello', 'world']};
/// final Uint8List enc = yajbeEncode(data);
/// final obj = yajbeDecodeUint8Array(enc);
/// print(obj); // {'a': 10, 'b': ['hello', 'world']}
/// ```
dynamic yajbeDecodeUint8Array(Uint8List data) {
  return yajbeDecode(BufferReader.fromUint8Array(data));
}

/// Parses the YAJBE encoded data and returns the resulting "Json" object.
dynamic yajbeDecodeByteData(ByteData data) {
  return yajbeDecode(BufferReader(data));
}

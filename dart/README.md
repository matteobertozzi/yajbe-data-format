# YAJBE for Dart

YAJBE is a compact binary data format built to be a drop-in replacement for JSON (JavaScript Object Notation).

## Usage & Examples
```dart
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
```

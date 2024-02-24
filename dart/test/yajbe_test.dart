import 'dart:typed_data';

import 'package:convert/convert.dart';
import 'package:yajbe/yajbe.dart';
import 'package:test/test.dart';

void assertEncode(Object? input, String expectedHex) {
  Uint8List enc = yajbeEncode(input);
  expect(hex.encoder.convert(enc), expectedHex);
}

void assertDecode(String expectedHex, Object? input) {
  var enc = hex.decoder.convert(expectedHex);
  expect(yajbeDecodeUint8Array(Uint8List.fromList(enc)), input);
}

void assertEncodeDecode(Object? input, String expectedHex) {
  Uint8List enc = yajbeEncode(input);
  expect(hex.encoder.convert(enc), expectedHex);
  expect(yajbeDecodeUint8Array(enc), input);
}

void main() {
  // no-op
}
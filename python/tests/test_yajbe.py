#
# Licensed to the Apache Software Foundation (ASF) under one or more
# contributor license agreements.  See the NOTICE file distributed with
# this work for additional information regarding copyright ownership.
# The ASF licenses this file to You under the Apache License, Version 2.0
# (the "License") you may not use this file except in compliance with
# the License.  You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

import unittest

from yajbe import encode_as_bytes, decode_bytes


class TestYajbe(unittest.TestCase):
    def test_bool_simple(self):
        self.assertEncodeDecode(False, "02")
        self.assertEncodeDecode(True, "03")

    def test_null_simple(self):
        self.assertEncodeDecode(None, "00")
        self.assertEncodeDecode([None], "2100")
        self.assertEncodeDecode([None, None], "220000")

    def test_int_simple(self):
        # positive ints
        self.assertEncodeDecode(1, "40")
        self.assertEncodeDecode(7, "46")
        self.assertEncodeDecode(24, "57")
        self.assertEncodeDecode(25, "5800")
        self.assertEncodeDecode(0xff, "58e6")
        self.assertEncodeDecode(0xffff, "59e6ff")
        self.assertEncodeDecode(0xffffff, "5ae6ffff")
        self.assertEncodeDecode(0xffffffff, "5be6ffffff")
        self.assertEncodeDecode(0xffffffffff, "5ce6ffffffff")
        self.assertEncodeDecode(0xffffffffffff, "5de6ffffffffff")
        self.assertEncodeDecode(0x1fffffffffffff, "5ee6ffffffffff1f")
        self.assertEncodeDecode(0xffffffffffffff, "5ee6ffffffffffff")
        self.assertEncodeDecode(0xfffffffffffffff, "5fe6ffffffffffff0f")
        self.assertEncodeDecode(0x7fffffffffffffff, "5fe6ffffffffffff7f")

        self.assertEncodeDecode(100, "584b")
        self.assertEncodeDecode(1000, "59cf03")
        self.assertEncodeDecode(1000000, "5a27420f")
        self.assertEncodeDecode(1000000000000, "5ce70fa5d4e8")
        self.assertEncodeDecode(100000000000000, "5de73f7a10f35a")

        # negative ints
        self.assertEncodeDecode(0, "60")
        self.assertEncodeDecode(-1, "61")
        self.assertEncodeDecode(-7, "67")
        self.assertEncodeDecode(-23, "77")
        self.assertEncodeDecode(-24, "7800")
        self.assertEncodeDecode(-25, "7801")
        self.assertEncodeDecode(-0xff, "78e7")
        self.assertEncodeDecode(-0xffff, "79e7ff")
        self.assertEncodeDecode(-0xffffff, "7ae7ffff")
        self.assertEncodeDecode(-0xffffffff, "7be7ffffff")
        self.assertEncodeDecode(-0xffffffffff, "7ce7ffffffff")
        self.assertEncodeDecode(-0xffffffffffff, "7de7ffffffffff")
        self.assertEncodeDecode(-0x1fffffffffffff, "7ee7ffffffffff1f")
        self.assertEncodeDecode(-0xffffffffffffff, "7ee7ffffffffffff")
        self.assertEncodeDecode(-0xfffffffffffffff, "7fe7ffffffffffff0f")
        self.assertEncodeDecode(-0x7fffffffffffffff, "7fe7ffffffffffff7f")

        self.assertEncodeDecode(-100, "784c")
        self.assertEncodeDecode(-1000, "79d003")
        self.assertEncodeDecode(-1000000, "7a28420f")
        self.assertEncodeDecode(-1000000000000, "7ce80fa5d4e8")
        self.assertEncodeDecode(-100000000000000, "7de83f7a10f35a")

    def test_float_simple(self):
        self.assertDecodeFloat("0500000000", 0.0)
        self.assertDecodeFloat("050000803f", 1.0)
        self.assertDecodeFloat("05cdcc8c3f", 1.1)
        self.assertDecodeFloat("050a1101c2", -32.26664)
        # self.assertEncodeDecode(float('inf'), "050000807f")
        # self.assertEncodeDecode(float('NaN'), "050000c07f")
        # self.assertEncodeDecode(-float('inf'), "05000080ff")

        self.assertDecodeFloat("060000000000000080", -0.0)
        self.assertDecodeFloat("0600000000000010c0", -4.0)
        self.assertEncodeDecodeFloat(-4.1, "0666666666666610c0")
        self.assertEncodeDecode(1.5, "06000000000000f83f")
        #self.assertEncodeDecode(65504.0, "59e0ff")
        self.assertDecodeFloat("060000000000fcef40", 65504.0)
        #self.assertEncodeDecode(100000.0, "5aa08601")
        self.assertDecodeFloat("0600000000006af840", 100000.0)
        self.assertEncodeDecode(5.960464477539063e-8, "06000000000000703e")
        self.assertEncodeDecode(0.00006103515625, "06000000000000103f")
        self.assertEncodeDecode(-5.960464477539063e-8, "0600000000000070be")
        self.assertEncodeDecode(3.4028234663852886e+38, "06000000e0ffffef47")
        self.assertEncodeDecode(9007199254740994.0, "060100000000004043")
        self.assertEncodeDecode(-9007199254740994.0, "0601000000000040c3")
        self.assertEncodeDecode(1.0e+300, "069c7500883ce4377e")
        self.assertEncodeDecode(-40.049149, "06c8d0b1834a0644c0")

    def test_string_simple(self):
        self.assertEncodeDecode("", "c0")
        self.assertEncodeDecode("a", "c161")
        self.assertEncodeDecode("abc", "c3616263")
        self.assertEncodeDecode("x" * 59, "fb" + "78" * 59)
        self.assertEncodeDecode("y" * 60, "fc01" + "79" * 60)
        self.assertEncodeDecode("y" * 127, "fc44" + "79" * 127)
        self.assertEncodeDecode("y" * 0xff, "fcc4" + "79" * 255)
        self.assertEncodeDecode("z" * 0x100, "fcc5" + "7a" * 256)
        self.assertEncodeDecode("z" * 314, "fcff" + "7a" * 314)
        self.assertEncodeDecode("z" * 315, "fd0001" + "7a" * 315)
        self.assertEncodeDecode("z" * 0xffff, "fdc4ff" + "7a" * 0xffff)
        self.assertEncodeDecode("k" * 0xfffff, "fec4ff0f" + "6b" * 0xfffff)
        self.assertEncodeDecode("k" * 0xffffff, "fec4ffff" + "6b" * 0xffffff)
        self.assertEncodeDecode("k" * 0x1000000, "fec5ffff" + "6b" * 0x1000000)

    def test_array_simple(self):
        self.assertDecode("2f01", [])
        self.assertEncodeDecode([], "20")
        self.assertEncodeDecode([1], "2140")
        self.assertEncodeDecode([0] * 10, "2a60606060606060606060")
        self.assertEncodeDecode([0] * 11, "2b016060606060606060606060")
        self.assertEncodeDecode([0] * 0xff, "2bf5" + "60" * 0xff)
        self.assertEncodeDecode([0] * 265, "2bff" + "60" * 265)
        self.assertEncodeDecode([0] * 0xffff, "2cf5ff" + "60" * 0xffff)
        self.assertEncodeDecode([0] * 0xffffff, "2df5ffff" + "60" * 0xffffff)

    def test_bytes_simple(self):
        self.assertEncodeDecode(bytearray(0), "80")
        self.assertEncodeDecode(bytearray(1), "8100")
        self.assertEncodeDecode(bytearray(3), "83000000")
        self.assertEncodeDecode(bytearray(59), "bb" + "00" * 59)
        self.assertEncodeDecode(bytearray(60), "bc01" + "00" * 60)
        self.assertEncodeDecode(bytearray(127), "bc44" + "00" * 127)
        self.assertEncodeDecode(bytearray(0xff), "bcc4" + "00" * 255)
        self.assertEncodeDecode(bytearray(0x100), "bcc5" + "00" * 256)
        self.assertEncodeDecode(bytearray(314), "bcff" + "00" * 314)
        self.assertEncodeDecode(bytearray(315), "bd0001" + "00" * 315)
        self.assertEncodeDecode(bytearray(0xffff), "bdc4ff" + "00" * 0xffff)
        self.assertEncodeDecode(bytearray(0xfffff), "bec4ff0f" + "00" * 0xfffff)
        self.assertEncodeDecode(bytearray(0xffffff), "bec4ffff" + "00" * 0xffffff)
        # self.assertEncodeDecode(bytearray(0x1000000), "bec5ffff" + "00" * (0x1000000))

    def test_map_simple(self):
        self.assertEncodeDecode({}, "30")
        self.assertEncode({}, "30")
        self.assertDecode("3f01", {})

        self.assertEncodeDecode({"a": 1}, "31816140")
        self.assertEncodeDecode({"a": "vA"}, "318161c27641")
        self.assertEncodeDecode({"a": [1, 2, 3]}, "31816123404142")
        self.assertEncodeDecode({"a": {"l": [1, 2, 3]}}, "31816131816c23404142")
        self.assertEncodeDecode({"a": {"l": {"x": 1}}}, "31816131816c31817840")

        self.assertDecode("3f81614001", {"a": 1})
        self.assertDecode("3f8161c2764101", {"a": "vA"})
        self.assertDecode("3f81612340414201", {"a": [1, 2, 3]})
        self.assertDecode("3f81613f816c234041420101", {"a": {"l": [1, 2, 3]}})
        self.assertDecode("3f81613f816c3f817840010101", {"a": {"l": {"x": 1}}})

        self.assertDecode("3f816140836f626a0001", {'a': 1, 'obj': None})
        self.assertDecode("3f816140836f626a3fa041a1000101", {'a': 1, 'obj': {'a': 2, 'obj': None}})
        self.assertDecode("3f816140836f626a3fa041a13fa042a100010101",
                          {'a': 1, 'obj': {'a': 2, 'obj': {'a': 3, 'obj': None}}})

    def test_map_provided_fields(self):
        INITIAL_FIELDS = ["hello", "world"]

        input = {'world': 2, 'hello': 1}

        # encode/decode with fields already present in the map. the names will not be in the encoded data
        enc = encode_as_bytes(input, INITIAL_FIELDS)
        self.assertEqual(bytes.fromhex("32a141a040"), enc)
        dec = decode_bytes(enc, INITIAL_FIELDS)
        self.assertEqual(input, dec)
        decx = decode_bytes(bytes.fromhex("3fa141a04001"), INITIAL_FIELDS)
        self.assertEqual(input, decx)

        # encode/decode adding a fields not in the base list
        input['something new'] = 3
        enc2 = encode_as_bytes(input, INITIAL_FIELDS)
        self.assertEqual(bytes.fromhex("33a141a0408d736f6d657468696e67206e657742"), enc2)
        dec2 = decode_bytes(enc2, INITIAL_FIELDS)
        self.assertEqual(input, dec2)
        dec2x = decode_bytes(bytes.fromhex("3fa141a0408d736f6d657468696e67206e65774201"), INITIAL_FIELDS)
        self.assertEqual(input, dec2x)

    def test_data_set_encode_decode(self):
        import hashlib
        import json
        import gzip
        import os
        for root, _, files in os.walk('../../test-data/'):
            for name in files:
                path = os.path.join(root, name)
                if name.endswith('.json'):
                    with open(path) as fd:
                        obj = json.load(fd)
                elif name.endswith('.json.gz'):
                    with gzip.open(path) as fd:
                        obj = json.load(fd)
                else:
                    continue

                enc = encode_as_bytes(obj)
                dec = decode_bytes(enc)
                self.assertEqual(obj, dec)
                print('encode/decode', path, len(enc), hashlib.sha256(enc).hexdigest())

    def assertEncode(self, input_obj, expected_hex: str):
        enc = encode_as_bytes(input_obj)
        self.assertEqual(enc, bytes.fromhex(expected_hex))
        return enc

    def assertDecode(self, hex_data: str, expected_obj):
        dec = decode_bytes(bytes.fromhex(hex_data))
        self.assertEqual(dec, expected_obj)

    def assertEncodeDecode(self, input_obj, expected_hex: str):
        enc = self.assertEncode(input_obj, expected_hex)
        dec = decode_bytes(enc)
        self.assertEqual(dec, input_obj)

    def assertDecodeFloat(self, hex_data: str, expected_obj: float):
        dec = decode_bytes(bytes.fromhex(hex_data))
        self.assertAlmostEqual(dec, expected_obj, 4)

    def assertEncodeDecodeFloat(self, input_obj: float, expected_hex: str):
        enc = encode_as_bytes(input_obj)
        self.assertEqual(enc, bytes.fromhex(expected_hex))
        dec = decode_bytes(enc)
        self.assertAlmostEqual(dec, input_obj, 4)


if __name__ == '__main__':
    unittest.main()

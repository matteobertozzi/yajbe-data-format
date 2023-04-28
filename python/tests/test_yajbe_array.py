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
import random

from yajbe import encode_as_bytes, decode_bytes


class TestYajbeArray(unittest.TestCase):
    def test_simple(self):
        self.assertDecode("20", [])
        self.assertDecode("2f01", [0] * 0)
        self.assertDecode("2f6001", [0] * 1)
        self.assertDecode("2f606001", [0] * 2)
        self.assertDecode("2f4001", [1])
        self.assertDecode("2f414101", [2, 2])

        self.assertEncodeDecode([0] * 0, "20")
        self.assertEncodeDecode([1], "2140")
        self.assertEncodeDecode([2, 2], "224141")
        self.assertEncodeDecode([0] * 10, "2a60606060606060606060")
        self.assertEncodeDecode([0] * 11, "2b016060606060606060606060")
        self.assertEncodeDecode([0] * 0xff, "2bf5" + "60" * 0xff)
        self.assertEncodeDecode([0] * 265, "2bff" + "60" * 265)
        self.assertEncodeDecode([0] * 0xffff, "2cf5ff" + "60" * 0xffff)
        #self.assertEncodeDecode([0] * 0xffffff, "2df5ffff" + "60" * 0xffffff) # too slow
        #self.assertEncodeDecode([0] * 0x1fffffff, "2e1fffff" + "60" * 0x1fffffff) # too slow

        self.assertEncodeDecode([None] * 0, "20")
        self.assertEncodeDecode([None] * 1, "2100")
        self.assertEncodeDecode([""], "21c0")
        self.assertEncodeDecode(["a"], "21c161")
        self.assertEncodeDecode([None] * 0xff, "2bf5" + "00" * 0xff)
        self.assertEncodeDecode([None] * 265, "2bff" + "00" * 265)
        self.assertEncodeDecode([None] * 0xffff, "2cf5ff" + "00" * 0xffff)
        #self.assertEncodeDecode([None] * 0xffffff, "2df5ffff" + "00" * 0xffffff) # too slow

    def test_small_array_length(self):
        for i in range(10):
            input = [0] * i
            enc = encode_as_bytes(input)
            self.assertEqual(1 + i, len(enc))
            self.assertEqual(input, decode_bytes(enc))

        for i in range(11, 266):
            input = [0] * i
            enc = encode_as_bytes(input)
            self.assertEqual(2 + i, len(enc))
            self.assertEqual(input, decode_bytes(enc))

        for i in range(266, 0x3ff):
            input = [0] * i
            enc = encode_as_bytes(input)
            self.assertEqual(3 + i, len(enc))
            self.assertEqual(input, decode_bytes(enc))

    def test_rand_string_set(self):
        set_data = set()
        for i in range(16):
            set_data.add('k-%s' % random.randint(0, 2 ** 63))

            encoded = encode_as_bytes(set_data)
            decoded = decode_bytes(encoded)
            self.assertEqual(set(decoded), set_data)

            self.assertEqual(encoded, encode_as_bytes(decoded))

    def test_rand_object_array(self):
        data = []
        for k in range(32):
            data.append({
              'boolValue': random.random() > 0.5,
              'intValue': random.randint(-((2 ** 31) - 1), (2 ** 31) - 1),
              'longValue': random.randint(-((2 ** 63) - 1), (2 ** 63) - 1),
              'floatValue': random.random(),
            })

            encoded = encode_as_bytes(data)
            decoded = decode_bytes(encoded)
            self.assertEqual(decoded, data)


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

if __name__ == '__main__':
    unittest.main()

#
# Licensed to the Apache Software Foundation (ASF) under one or more
# contributor license agreements.  See the NOTICE file distributed with
# this work for additional information regarding copyright ownership.
# The ASF licenses this file to You under the Apache License, Version 2.0
# (the "License"); you may not use this file except in compliance with
# the License.  You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

import struct
import io


class FieldNameReader:
    def __init__(self, decoder, initialFieldNames=None) -> None:
        self._decoder = decoder
        self._indexed_names = []
        self._last_key = b''
        if initialFieldNames:
            for name in initialFieldNames[:0xffff]:
                self._indexed_names.append(name.encode('utf-8'))

    def decode_string(self) -> str:
        head = self._decoder._read_byte()
        match (head >> 5) & 0b111:
            case 0b100:
                return self._read_full_field_name(head)
            case 0b101:
                return self._read_indexed_field_name(head)
            case 0b110:
                return self._read_prefix(head)
            case 0b111:
                return self._read_prefix_suffix(head)

    def _read_full_field_name(self, head: int) -> str:
        length = self._read_length(head)
        utf8 = self._decoder._read_bytes(length)
        return self._add_to_index(utf8)

    def _read_indexed_field_name(self, head: int) -> str:
        field_index = self._read_length(head)
        utf8 = self._indexed_names[field_index]
        self._last_key = utf8
        return utf8.decode('utf-8')

    def _read_prefix(self, head: int) -> str:
        length = self._read_length(head)
        prefix = self._decoder._read_byte()
        kpart = self._decoder._read_bytes(length)
        utf8 = self._last_key[0:prefix] + kpart
        return self._add_to_index(utf8)

    def _read_prefix_suffix(self, head: int) -> str:
        length = self._read_length(head)
        prefix = self._decoder._read_byte()
        suffix = self._decoder._read_byte()
        kpart = self._decoder._read_bytes(length)
        utf8 = self._last_key[0:prefix] + kpart + self._last_key[len(self._last_key) - suffix:]
        return self._add_to_index(utf8)

    def _read_length(self, head: int) -> int:
        length = head & 0b000_11111
        match length:
            case 30:
                return self._decoder._read_byte()
            case 31:
                return self._decoder._read_uint(2)
        return length

    def _add_to_index(self, utf8: bytes) -> str:
        self._indexed_names.append(utf8)
        self._last_key = utf8
        return utf8.decode('utf-8')


class YajbeDecoder:
    def __init__(self, stream: io.BufferedReader, initialFieldNames=None) -> None:
        self._stream = stream
        self._field_name_reader = FieldNameReader(self, initialFieldNames)

    def decode_item(self):
        head = self._read_byte()
        if (head & 0b11_000000) == 0b11_000000:
            return self._decode_string(head)
        if (head & 0b10_000000) == 0b10_000000:
            return self._decode_bytes(head)
        if (head & 0b010_00000) == 0b010_00000:
            return self._decode_int(head)
        if (head & 0b0011_0000) == 0b0011_0000:
            return self._decode_object(head)
        if (head & 0b0010_0000) == 0b0010_0000:
            return self._decode_array(head)
        if (head & 0b000001_00) == 0b000001_00:
            return self._decode_float(head)
        match head:
            case 0b00000000:
                return None
            case 0b00000001:
                return None
            case 0b00000010:
                return False
            case 0b00000011:
                return True
            case other:
                raise TypeError("unsupported head " + bin(other))

    def _decode_bytes(self, head: int) -> bytes:
        w = head & 0b111111
        if w < 60:
            return self._read_bytes(w)

        length = self._read_uint(w - 59)
        return self._read_bytes(length)

    def _decode_string(self, head: int) -> str:
        utf8 = self._decode_bytes(head)
        return utf8.decode('utf-8')

    def _decode_int(self, head: int) -> int:
        signed = (head & 0b011_00000) == 0b011_00000

        w = head & 0b11111
        if w < 24:
            return -w if signed else (1 + w)

        value = self._read_uint(w - 23)
        return -value if signed else value

    def _decode_float(self, head: int) -> float:
        match head & 0b11:
            case 0b00:
                raise ValueError('unsupported decode float16/var-float')
            case 0b01:
                return struct.unpack('<f', self._read_bytes(4))[0]
            case 0b10:
                return struct.unpack('<d', self._read_bytes(8))[0]
            case 0b11:
                raise ValueError('unsupported decode bigdecimal')

    def _decode_array(self, head: int) -> list:
        w = head & 0b1111
        if w == 0b1111:
            result = []
            while self._read_has_more():
                result.append(self.decode_item())
            return result

        length = self._read_length(w, 10)
        result = []
        for _ in range(length):
            result.append(self.decode_item())
        return result

    def _decode_object(self, head: int) -> dict:
        w = head & 0b1111
        if w == 0b1111:
            result = {}
            while self._read_has_more():
                key = self._field_name_reader.decode_string()
                result[key] = self.decode_item()
            return result

        length = self._read_length(w, 10)
        result = {}
        for _ in range(length):
            key = self._field_name_reader.decode_string()
            result[key] = self.decode_item()
        return result

    def _read_has_more(self) -> bool:
        v = self._stream.peek(1)
        if len(v) < 1:
            raise EOFError('peek %s' % v)
        if v[0] != 0b00000001:
            return True
        self._read_byte()
        return False

    def _read_length(self, w: int, inline_max: int) -> int:
        if w <= inline_max:
            return w
        return self._read_uint(w - inline_max)

    def _read_byte(self) -> int:
        v = self._stream.read(1)
        if len(v) != 1:
            raise EOFError()
        return v[0] & 0xff

    def _read_bytes(self, length: int) -> bytes:
        data = self._stream.read(length)
        if len(data) != length:
            raise EOFError()
        return data

    def _read_uint(self, width: int) -> int:
        buf = self._read_bytes(width)
        value = 0
        for i in range(width):
            value += (buf[i] & 0xff) << (i << 3)
        return value


def decode_stream(stream: io.BufferedReader, initialFieldNames=None):
    if not isinstance(stream, io.BufferedReader):
        raise Exception('expected a buffered stream')

    decoder = YajbeDecoder(stream, initialFieldNames)
    return decoder.decode_item()


def decode_bytes(data: bytes, initialFieldNames=None):
    with io.BufferedReader(io.BytesIO(data)) as stream:
        return decode_stream(stream, initialFieldNames)

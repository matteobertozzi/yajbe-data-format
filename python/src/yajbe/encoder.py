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

import io
import struct
import types

from .freq import EnumLruMapping, YajbeEncoderEnumConfig

CALLABLES_TYPES = (types.FunctionType, types.MethodType)

def is_object(obj) -> bool:
    # TODO: we must find a better way to do this!
    return not isinstance(obj, CALLABLES_TYPES) and hasattr(obj, '__dict__')

def object_as_dict(obj: object) -> dict[str, any]:
    return {k: v for k, v in obj.__dict__.items() if k[0] != '_' and not isinstance(v, CALLABLES_TYPES)}

def int_bytes_width(v: int) -> int:
    return (v.bit_length() + 7) // 8 if v != 0 else 1


class FieldNameWriter:
    def __init__(self, encoder, initial_field_names: list[str] = None) -> None:
        self._encoder = encoder
        self._indexed_map = {}
        self._last_key = b''
        if initial_field_names:
            for i, name in enumerate(initial_field_names[:65819]):
                self._indexed_map[name] = i

    def encode_string(self, key: str) -> None:
        utf8 = key.encode('utf-8')

        index = self._indexed_map.get(key)
        if index is not None:
            self._write_indexed_field_name(index)
            self._last_key = utf8
            return

        if self._last_key and len(utf8) > 4:
            prefix = min(0xff, self._prefix(utf8))
            suffix = self._suffix(utf8, prefix)

            if suffix > 2:
                self._write_prefix_suffix(utf8, prefix, min(0xff, suffix))
            elif prefix > 2:
                self._write_prefix(utf8, prefix)
            else:
                self._write_full_field_name(utf8)
        else:
            self._write_full_field_name(utf8)

        if len(self._indexed_map) < 65819:
            self._indexed_map[key] = len(self._indexed_map)
        self._last_key = utf8

    def _write_full_field_name(self, utf8: bytes):
        # 100----- Full Field Name (0-29 length - 1, 30 1b-len, 31 2b-len)
        self._write_length(0b100_00000, len(utf8))
        self._encoder._write_bytes(utf8)

    def _write_indexed_field_name(self, index: int):
        # 101----- Field Offset (0-29 field, 30 1b-len, 31 2b-len)
        self._write_length(0b101_00000, index)

    def _write_prefix(self, utf8: bytes, prefix: int):
        # 110----- Prefix (1byte prefix, 0-29 length - 1, 30 1b-len, 31 2b-len)
        length = len(utf8) - prefix
        self._write_length(0b110_00000, length)
        self._encoder._write_byte(prefix)
        self._encoder._write_bytes(utf8[prefix:])

    def _write_prefix_suffix(self, utf8: bytes, prefix: int, suffix: int):
        # 111----- Prefix/Suffix (1byte prefix, 1byte suffix, 0-29 length - 1, 30 1b-len, 31 2b-len)
        length = len(utf8) - prefix - suffix
        self._write_length(0b111_00000, length)
        self._encoder._write_byte(prefix)
        self._encoder._write_byte(suffix)
        self._encoder._write_bytes(utf8[prefix:prefix + length])

    def _write_length(self, head: int, length: int):
        if length < 30:
            self._encoder._write_byte(head | length)
        elif length <= 284:
            # 30 + 1byte = 284
            self._encoder._write_byte(head | 0b11110)
            self._encoder._write_byte((length - 29) & 0xff)
        elif length <= 65819:
            # 31 + 2byte = 65819
            self._encoder._write_byte(head | 0b11111)
            self._encoder._write_byte((length - 284) // 256)
            self._encoder._write_byte((length - 284) & 255)
        else:
            raise Exception("unexpected too many field names: %s" % length)

    def _prefix(self, key: bytes) -> int:
        last_key = self._last_key
        min_len = min(len(last_key), len(key))
        for i in range(min_len):
            if last_key[i] != key[i]:
                return i
        return min_len

    def _suffix(self, key: bytes, key_prefix: int) -> int:
        last_key = self._last_key
        last_key_len = len(last_key)
        key_len = len(key) - key_prefix
        min_len = min(len(last_key), key_len)
        for i in range(1, min_len + 1):
            if (last_key[last_key_len - i] & 0xff) != (key[key_prefix + (key_len - i)] & 0xff):
                return i - 1
        return min_len

from dataclasses import is_dataclass, asdict as dataclass_as_dict
class YajbeEncoder:
    def __init__(self, stream: io.BufferedIOBase, initial_field_names: list[str] = None, enum_config: YajbeEncoderEnumConfig = None) -> None:
        self._stream = stream
        self._field_name_writer = FieldNameWriter(self, initial_field_names)
        self._enum_config = enum_config
        self._enum_mapping = None
        self._types_map = {
            bool: self.encode_bool,
            int: self.encode_int,
            float: self.encode_float,
            bytes: self.encode_bytes,
            bytearray: self.encode_bytes,
            str: self.encode_string,
            list: self.encode_array,
            tuple: self.encode_array,
            set: self.encode_array,
            dict: self.encode_object,
        }

    def encode_item(self, item):
        if item is None:
            return self.encode_null()

        item_type = type(item)
        encoder = self._types_map.get(item_type)
        if encoder is None:
            if is_dataclass(item):
                item = dataclass_as_dict(item)
                return self.encode_object(item)
            elif is_object(item):
                item = object_as_dict(item)
                return self.encode_object(item)
            else:
                raise Exception('unsupported type %s: %s' % (item_type, item))
        return encoder(item)

    def encode_null(self):
        self._write_byte(0)

    def encode_bool(self, value: bool):
        self._write_byte(3 if value else 2)

    def _encode_positive_int(self, value: int) -> None:
        if value <= 24:
            self._write_byte(0b010_00000 | (value - 1))
        else:
            value -= 25
            nbytes = int_bytes_width(value)
            self._write_byte(0b010_00000 | (23 + nbytes))
            self._write_uint(value, nbytes)

    def _encode_negative_int(self, value: int) -> None:
        value = -value
        if value <= 23:
            self._write_byte(0b011_00000 | value)
        else:
            value -= 24
            nbytes = int_bytes_width(value)
            self._write_byte(0b011_00000 | (23 + nbytes))
            self._write_uint(value, nbytes)

    def encode_int(self, value: int) -> None:
        if value > 0:
            self._encode_positive_int(value)
        else:
            self._encode_negative_int(value)

    def encode_float(self, value: float) -> None:
        data = struct.pack('<d', value)
        self._write_byte(0b00000_110)
        self._stream.write(data)

    def encode_string(self, value: str) -> None:
        if self._enum_config and self.encode_enum(value):
            return

        utf8 = value.encode('utf-8')
        self._write_length(0b11_000000, 59, len(utf8))
        self._stream.write(utf8)

    def encode_enum(self, text: str) -> bool:
        if self._enum_mapping is None:
            self.new_enum_mapping()

        index = self._enum_mapping.add(text)
        if index < 0: return False

        if index <= 0xff:
            self._write_byte(0b00001001)
            self._write_byte(index)
        elif index <= 0xffff:
            self._write_byte(0b00001010)
            self._write_uint(index, 2)
        else:
            raise Exception("enum index too large " + index)
        return True

    def new_enum_mapping(self) -> None:
        config = self._enum_config
        match config.TYPE:
            case 'LRU':
                self._enum_mapping = EnumLruMapping(config.lru_size, config.min_freq)

                self._write_byte(0b00001000)
                self._write_byte(config.lru_size.bit_length() - 6)
                self._write_byte(config.min_freq - 1)
                return

    def encode_bytes(self, value: bytes) -> None:
        self._write_length(0b10_000000, 59, len(value))
        self._stream.write(value)

    def encode_object(self, obj: dict) -> None:
        keys = obj.keys()
        self._write_length(0b0011_0000, 10, len(keys))
        for key in keys:
            self._field_name_writer.encode_string(key)
            self.encode_item(obj[key])

    def encode_array(self, array: list) -> None:
        self._write_length(0b0010_0000, 10, len(array))
        for v in array:
            self.encode_item(v)

    def _write_length(self, head: int, inline_max: int, length: int) -> None:
        if length <= inline_max:
            self._write_byte(head | length)
        else:
            delta_length = length - inline_max
            nbytes = int_bytes_width(delta_length)
            self._write_byte(head | (inline_max + nbytes))
            self._write_uint(delta_length, nbytes)

    def _write_byte(self, v: int) -> None:
        self._stream.write(bytes([v & 0xff]))

    def _write_bytes(self, value: bytes) -> None:
        self._stream.write(value)

    def _write_uint(self, value: int, width: int) -> None:
        buf = bytearray(width)
        for i in range(width):
            buf[i] = (value >> (i << 3)) & 0xff
        self._stream.write(buf)


def encode_to_stream(stream: io.BufferedIOBase, value, initial_field_names: list[str] = None, enum_config: YajbeEncoderEnumConfig = None) -> None:
    decoder = YajbeEncoder(stream, initial_field_names, enum_config)
    decoder.encode_item(value)


def encode_as_bytes(value, initial_field_names: list[str] = None, enum_config: YajbeEncoderEnumConfig = None) -> bytes:
    with io.BytesIO() as stream:
        encode_to_stream(stream, value, initial_field_names, enum_config)
        return stream.getvalue()

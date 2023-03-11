#!/usr/bin/env python3

from yajbe import YajbeEnumLruConfig
import yajbe

# simple usage: encode and decode from bytes
enc = yajbe.encode_as_bytes({'a': 10, 'b': ['hello', 10]})
dec = yajbe.decode_bytes(enc)
print(dec)

# options: identify common strings and avoid writing them every time
enc = yajbe.encode_as_bytes([{'a': "foooo"}, {'a': "foooo"}, {'a': "foooo"}, {'a': "foooo"}], enum_config=YajbeEnumLruConfig(256, 4))
dec = yajbe.decode_bytes(enc)
print(dec)

# encode directly to a stream
with open('test.yajbe', 'wb') as fd:
    yajbe.encode_to_stream(fd, {'a': 10, 'b': ['hello', 10]})

# decode directly from a stream
with open('test.yajbe', 'rb') as fd:
    print(yajbe.decode_stream(fd))  # {'a': 10, 'b': ['hello', 10]}


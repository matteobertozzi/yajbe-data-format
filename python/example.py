#!/usr/bin/env python3

from dataclasses import dataclass

from yajbe import YajbeEnumLruConfig
import yajbe

def demo_simple():
    # simple usage: encode and decode from bytes
    enc = yajbe.encode_as_bytes({'a': 10, 'b': ['hello', 10]})
    dec = yajbe.decode_bytes(enc)
    print(dec)

    # options: identify common strings and avoid writing them every time
    enc = yajbe.encode_as_bytes([{'a': "foooo"}, {'a': "foooo"}, {'a': "foooo"}, {'a': "foooo"}], enum_config=YajbeEnumLruConfig(256, 4))
    dec = yajbe.decode_bytes(enc)
    print(dec)

def demo_stream():
    # encode directly to a stream
    with open('test.yajbe', 'wb') as fd:
        yajbe.encode_to_stream(fd, {'a': 10, 'b': ['hello', 10]})

    # decode directly from a stream
    with open('test.yajbe', 'rb') as fd:
        print(yajbe.decode_stream(fd))  # {'a': 10, 'b': ['hello', 10]}

@dataclass
class DcBar:
    x: int
    y: str

@dataclass
class DcFoo:
    a: int
    b: DcBar

def demo_dataclass():
    b = DcBar(5, 'foo')
    r = yajbe.encode_as_bytes(b)
    print(r, yajbe.decode_bytes(r))

    f = DcFoo(10, b)
    r = yajbe.encode_as_bytes(f)
    print(r, yajbe.decode_bytes(r))

class ObjBar:
    def __init__(self, x: int, y: str) -> None:
        self.x = x
        self.y = y
        self._z = x * 123

    def xyz(self):
        pass

class ObjFoo:
    def __init__(self, a: int, b: ObjBar) -> None:
        self.a = a
        self.b = b
        self._z = a * 456

    def xyz(self):
        pass

def demo_class_obj():
    b = ObjBar(5, 'foo')
    r = yajbe.encode_as_bytes(b)
    print(r, yajbe.decode_bytes(r))

    f = ObjFoo(10, b)
    r = yajbe.encode_as_bytes(f)
    print(r, yajbe.decode_bytes(r))

if __name__ == '__main__':
    demo_simple()
    demo_stream()
    demo_dataclass()
    demo_class_obj()
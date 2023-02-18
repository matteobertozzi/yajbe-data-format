# YAJBE for Python

NOTE: the python implementation is still a work in progress

```python
import yajbe

# encode and decode from bytes
enc = yajbe.encode_as_bytes({'a': 10, 'b': ['hello', 10]})
dec = yajbe.decode_bytes(enc)
print(dec)

# encode directly to a stream
with open('test.yajbe', 'wb') as fd:
  yajbe.encode_to_stream(fd, {'a': 10, 'b': ['hello', 10]})

# decode directly from a stream
with open('test.yajbe', 'rb') as fd:
  print(yajbe.decode_stream(fd)) # {'a': 10, 'b': ['hello', 10]}
```

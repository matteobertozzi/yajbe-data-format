# YAJBE for Python

YAJBE is a compact binary data format built to be a drop-in replacement for JSON (JavaScript Object Notation).


## Motivation for a new format
We have a lot of services exchanging or storing data using JSON, and most of them don't want to switch to a data format that requires a schema.

We wanted to remove the overhead of the JSON format (especially field names), but keeping the same data model flexibility (numbers, strings, arrays, maps/objects, and a few values such as false, true, and null).

See more at https://github.com/matteobertozzi/yajbe-data-format

### Install the package
You can find the package at https://pypi.org/project/yajbe. \
Python >=3.10 is required. To install or upgrade you can use:
```bash
$ pip install --upgrade yajbe
```

## Usage

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

# YAJBE npm package for Javascript/Typescript

YAJBE is a compact binary data format built to be a drop-in replacement for JSON (JavaScript Object Notation).

## Install
```bash
$ npm install yajbe-dataformat
```

```
$ cat package.json
...
"dependencies": {
  "yajbe-dataformat": "^1.0.0",
  ...
},
```

## Usage & Examples
A simple example using deno remote import is below. but you can use local import as usual.
```typescript
import * as YAJBE from 'yajbe-dataformat';

const enc: Uint8Array = YAJBE.encode({a: "hello", b: [1, 2, 3]});
const dec = YAJBE.decode(enc); // {a: "hello", b: [1, 2, 3]}
```

## Supported Types
Aside from the basic types supported by JSON.stringify(), YAJBE.encode() also support **Map**, **Set**, **Uint8Array** and the others ArrayBufferView.

### TODO
Some things are not supported yet by the implementation.
 * Decode variable length float
 * Encode/Decode BigInt
 * Encode/Decode BigDecimal
 * Handle number > MAX_SAFE as BigInt?

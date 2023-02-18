# YAJBE for Javascript/Typescript

The single file source contains only **TextEncoder**/**TextDecoder** as "dependency".

If you use [Deno](https://deno.land/) or [Bun](https://bun.sh/) or a browser you'll have no problem to run the code.

A simple example using deno remote import is below. but you can use local import as usual.
```typescript
import * as YAJBE from 'https://raw.githubusercontent.com/matteobertozzi/yajbe-data-format/main/typescript/yajbe.ts';

const enc: Uint8Array = YAJBE.encode({a: "hello", b: [1, 2, 3]});
const dec = YAJBE.decode(enc); // {a: "hello", b: [1, 2, 3]}
```

_We use some Deno code on the server side and Angular for web consoles in our code bases._ So tests are written using Deno, and you can run them using `deno test -A yajbe.test.ts`

## Supported Types
Aside from the basic types supported by JSON.stringify(), YAJBE.encode() also support **Map**, **Set**, **Uint8Array** and the others ArrayBufferView.

### TODO
Some things are not supported yet by the implementation.
 * Decode variable length float
 * Encode/Decode BigInt
 * Encode/Decode BigDecimal
 * Handle number > MAX_SAFE as BigInt?

import * as YAJBE from './yajbe.ts';

const enc: Uint8Array = YAJBE.encode({a: "hello", b: [1, 2, 3]});
const dec = YAJBE.decode(enc); // {a: "hello", b: [1, 2, 3]}
console.log(enc);
console.log(dec);

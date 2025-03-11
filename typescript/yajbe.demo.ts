import * as YAJBE from './yajbe.ts';

const r = YAJBE.encode({a: 0});
console.log(YAJBE.decode(r));

const r2 = YAJBE.encode(0);
console.log(YAJBE.decode(r2));

// Simple usage
const enc: Uint8Array = YAJBE.encode({a: "hello", b: [1, 2, 3]});
const dec = YAJBE.decode(enc); // {a: "hello", b: [1, 2, 3]}
console.log(enc);
console.log(dec);

// options: known field names
const opts2 = {fieldNames: ['a', 'b']}
const enc2: Uint8Array = YAJBE.encode({a: "hello", b: [1, 2, 3]}, opts2);
const dec2 = YAJBE.decode(enc2, opts2); // {a: "hello", b: [1, 2, 3]}
console.log(enc2);
console.log(dec2);


// options: identify common strings and avoid writing them every time
const opts3: YAJBE.YajbeEncoderOptions = {enumConfig: { type: 'LRU', specs: { lruSize: 128, minFreq: 1 }  }};
const enc3: Uint8Array = YAJBE.encode([{a: "foooo"}, {a: "foooo"}, {a: "foooo"}, {a: "foooo"}], opts3);
const dec3 = YAJBE.decode(enc3); // [{a: "foooo"}, {a: "foooo"}, {a: "foooo"}, {a: "foooo"}]
console.log(enc3);
console.log(dec3);

// options: identify common strings and avoid writing them every time
const opts4: YAJBE.YajbeEncoderOptions = {enumConfig: { type: 'ANY' }};
const enc4: Uint8Array = YAJBE.encode([{a: "aaaaa"}, {a: "bbbbb"}, {a: "aaaaa"}, {a: "bbbbb"}], opts4);
const dec4 = YAJBE.decode(enc3); // [{a: "aaaaa"}, {a: "bbbbb"}, {a: "aaaaa"}, {a: "bbbbb"}]
console.log(enc4);
console.log(dec4);


const data = ['aa', 'aa', 'aa', 'aa', 'bb', 'bb', 'bb'];
console.log(YAJBE.encode(data));
const enc5 = YAJBE.encode(data, { enumConfig: { type: 'ANY' }});
console.log(enc5);
console.log(YAJBE.decode(enc5));
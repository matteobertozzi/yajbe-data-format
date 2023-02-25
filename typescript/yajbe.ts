/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

export function encode(value: unknown, options?: { bufSize?: number, fieldNames?: string[] }): Uint8Array {
  const writer = new InMemoryBytesWriter(options?.bufSize);
  const encoder = new YajbeEncoder(writer, options?.fieldNames);
  encoder.encodeItem(value);
  encoder.flush();
  return writer.slice();
}

export function decode<T>(data: Uint8Array, options?: { fieldNames?: string[] }): T {
  const reader = new InMemoryBytesReader(data);
  const decoder = new YajbeDecoder(reader, options?.fieldNames);
  return decoder.decodeItem() as T;
}

// ==============================================================================================================
class DataEncoder {
  encodeItem(value: unknown): void {
    if (value === false) {
      this.encodeFalse();
    } else if (value === true) {
      this.encodeTrue();
    } else if (value === null) {
      this.encodeNull();
    } else if (value === undefined) {
      this.encodeUndefined();
    } else switch (typeof value) {
      case 'string':
        this.encodeString(value as string);
        break;
      case 'number':
        this.encodeNumber(value);
        break;
      case 'bigint':
        this.encodeBigInt(value);
        break;
      default:
        if (Array.isArray(value)) {
          this.encodeArray(value);
        } else if (ArrayBuffer.isView(value)) {
          this.encodeArrayBuffer(value);
        } else if (value instanceof Date) {
          this.encodeDate(value);
        } else if (value instanceof Map) {
          this.encodeMap(value);
        } else if (value instanceof Set) {
          this.encodeSet(value);
        } else {
          this.encodeObject(value as {[key: string]: unknown});
        }
        break;
    }
  }

  protected encodeNumber(value: number): void {
    if (Number.isSafeInteger(value)) {
      this.encodeInteger(value);
    } else if (Math.floor(value) === value) {
      const POW_2_53 = 9007199254740992;
      if (-POW_2_53 <= value && value <= POW_2_53) {
        this.encodeInteger(value);
      } else {
        this.encodeFloat(value);
      }
    } else {
      this.encodeFloat(value);
    }
  }

  protected encodeArrayBuffer(value: ArrayBufferView): void {
    if (value instanceof Uint8Array) {
      this.encodeUint8Array(value);
    } else {
      this.encodeUint8Array(new Uint8Array(value.buffer, value.byteOffset, value.byteLength));
    }
  }

  // null
  protected encodeNull(): void { throw new Error("Not implemented"); }
  protected encodeUndefined(): void { throw new Error("Not implemented"); }

  // boolean
  protected encodeTrue(): void { throw new Error("Not implemented"); }
  protected encodeFalse(): void { throw new Error("Not implemented"); }

  // float
  protected encodeFloat(_: number): void { throw new Error("Not implemented"); }

  // integer
  protected encodeInteger(value: number): void { throw new Error("Not implemented"); }
  protected encodeBigInt(_: BigInt): void { throw new Error("Not implemented"); }

  protected encodeDate(_: Date): void { throw new Error("Not implemented"); }

  // string/bytes
  protected encodeString(_: string): void { throw new Error("Not implemented"); }
  protected encodeUint8Array(_: Uint8Array): void { throw new Error("Not implemented"); }

  // array/set
  protected encodeArray(_: ArrayLike<unknown>): void { throw new Error("Not implemented"); }
  protected encodeSet(v: Set<unknown>): void { this.encodeArray(Array.from(v)); }

  // object/map
  protected encodeObject(_: {[key: string]: unknown}): void { throw new Error("Not implemented"); }
  protected encodeMap(v: Map<unknown, unknown>): void { this.encodeObject(Object.fromEntries(v)); }
}

interface BytesReader {
  reset(): void;
  hasMore(): boolean;
  peekUint8(): number;
  readUint8(): number;
  readUint8Array(nbytes: number): Uint8Array;

  readUint(width: number): number;
  readUint16(bigEndian?: boolean): number;
  readUint24(bigEndian?: boolean): number;
  readUint32(bigEndian?: boolean): number;
  readUint40(bigEndian?: boolean): number;
  readUint48(bigEndian?: boolean): number;
  readUint56(bigEndian?: boolean): number;
  readUint64(bigEndian?: boolean): number;

  readFloat16(bigEndian?: boolean): number;
  readFloat32(bigEndian?: boolean): number;
  readFloat64(bigEndian?: boolean): number;
}

interface BytesWriter {
  flush(): void;
  writeUint8(value: number): void;
  writeUint8Array(value: Uint8Array | ArrayLike<number> | number[]): void;

  writeUint(value: number, width: number, bigEndian?: boolean): void;
  writeUint16(value: number, bigEndian?: boolean): void;
  writeUint24(value: number, bigEndian?: boolean): void;
  writeUint32(value: number, bigEndian?: boolean): void;
  writeUint40(value: number, bigEndian?: boolean): void;
  writeUint48(value: number, bigEndian?: boolean): void;
  writeUint56(value: number, bigEndian?: boolean): void;
  writeUint64(value: number, bigEndian?: boolean): void;

  writeFloat32(value: number, bigEndian?: boolean): void;
  writeFloat64(value: number, bigEndian?: boolean): void;
}

// ==============================================================================================================
function intBytesWidth(value: number) {
  if (value <= 0xff) return 1;
  if (value <= 0xffff) return 2;
  if (value <= 0xffffff) return 3;
  if (value <= 0xffffffff) return 4;
  if (value <= 0xffffffffff) return 5;
  if (value <= 0xffffffffffff) return 6;
  if (value <= 0xffffffffffffff) return 7;
  return 8;
}

const POW2_8SHIFTS = [1, 256, 65536, 16777216, 4294967296, 1099511627776, 281474976710656, 72057594037927940];

function decodeInt(buffer: Uint8Array, offset: number, width: number, bigEndian?: boolean): number {
  let value = 0;
  if (bigEndian) {
    for (let i = 0; i < width; ++i) {
      value += (buffer[offset + i] & 0xff) * POW2_8SHIFTS[(width - 1) - i];
    }
  } else {
    for (let i = 0; i < width; ++i) {
      value += (buffer[offset + i] & 0xff) * POW2_8SHIFTS[i];
    }
  }
  return value;
}

function encodeInt(buffer: Uint8Array, offset: number, value: number, width: number, bigEndian?: boolean): void {
  if (bigEndian) {
    for (let i = (width - 1); i >= 0; --i) {
      buffer[offset++] = Math.floor(value / POW2_8SHIFTS[i]);
    }
  } else {
    for (let i = 0; i < width; ++i) {
      buffer[offset + i] = Math.floor(value / POW2_8SHIFTS[i]);
    }
  }
}


// ================================================================================================
//  Float related
// ================================================================================================
function encodeFloat(buffer: Uint8Array, offset: number, value: number, width: number, bigEndian?: boolean): void {
  switch (width) {
    case 4: { // float32
      new DataView(buffer.buffer, offset).setFloat32(0, value, !bigEndian);
      return;
    }
    case 8: { // float64
      new DataView(buffer.buffer, offset).setFloat64(0, value, !bigEndian);
      return;
    }
    case 2: { // float16/vle-float
      throw new Error("Not implemented encode float16/vle-float");
    }
  }
  throw new Error("Not implemented width " + width);
}

function decodeFloat(buffer: Uint8Array, offset: number, width: number, bigEndian?: boolean): number {
  switch (width) {
    case 4: { // float32
      return new DataView(buffer.buffer).getFloat32(0, !bigEndian);
    }
    case 8: { // float64
      return new DataView(buffer.buffer).getFloat64(0, !bigEndian);
    }
    case 2: { // float16/vle-float
      throw new Error("Not implemented decode float16/vle-float");
    }
  }
  throw new Error("Not implemented decode float width " + width);
}

// ==============================================================================================================
class AbstractBytesReader implements BytesReader {
  reset(): void { throw new Error("Not implemented"); }
  hasMore(): boolean { throw new Error("Not implemented"); }
  peekUint8(): number { throw new Error("Not implemented"); }
  readUint8(): number { throw new Error("Not implemented"); }
  readUint8Array(_: number): Uint8Array { throw new Error("not implemented"); }

  constructor() {
    // no-op
  }

  readFloat16(bigEndian?: boolean): number {
    const buf = this.readUint8Array(2);
    return decodeFloat(buf, 0, 2, bigEndian);
  }

  readFloat32(bigEndian?: boolean): number {
    const buf = this.readUint8Array(4);
    return decodeFloat(buf, 0, 4, bigEndian);
  }

  readFloat64(bigEndian?: boolean): number {
    const buf = this.readUint8Array(8);
    return decodeFloat(buf, 0, 8, bigEndian);
  }

  readUint(width: number, bigEndian?: boolean): number {
    const buf = this.readUint8Array(width);
    return decodeInt(buf, 0, width, bigEndian);
  }

  readUint16(bigEndian?: boolean): number { return this.readUint(2, bigEndian); }
  readUint24(bigEndian?: boolean): number { return this.readUint(3, bigEndian); }
  readUint32(bigEndian?: boolean): number { return this.readUint(4, bigEndian); }
  readUint40(bigEndian?: boolean): number { return this.readUint(5, bigEndian); }
  readUint48(bigEndian?: boolean): number { return this.readUint(6, bigEndian); }
  readUint56(bigEndian?: boolean): number { return this.readUint(7, bigEndian); }
  readUint64(bigEndian?: boolean): number { return this.readUint(8, bigEndian); }
}

export class InMemoryBytesReader extends AbstractBytesReader {
  private readonly buffer: Uint8Array;
  private offset: number;

  constructor(buffer: Uint8Array) {
    super();
    this.buffer = buffer;
    this.offset = 0;
  }

  reset(): void {
    this.offset = 0;
  }

  hasMore(): boolean {
    return this.offset < this.buffer.length;
  }

  peekUint8(): number {
    return this.buffer[this.offset] & 0xff;
  }

  readUint8(): number {
    return this.buffer[this.offset++] & 0xff;
  }

  readUint8Array(nbytes: number): Uint8Array {
    const data = this.buffer.slice(this.offset, this.offset + nbytes);
    this.offset += nbytes;
    return data;
  }
}

// ==============================================================================================================
class AbstractBytesWriter implements BytesWriter {
  flush(): void { throw new Error("Not implemented"); }
  writeUint8(_: number): void { throw new Error("Not implemented"); }
  writeUint8Array(_: Uint8Array | ArrayLike<number> | number[]): void { throw new Error("not implemented"); }

  constructor() {
    // no-op
  }

  writeFloat32(value: number, bigEndian?: boolean): void {
    const buf = new Uint8Array(4);
    encodeFloat(buf, 0, value, 4, bigEndian);
    this.writeUint8Array(buf);
  }

  writeFloat64(value: number, bigEndian?: boolean): void {
    const buf = new Uint8Array(8);
    encodeFloat(buf, 0, value, 8, bigEndian);
    this.writeUint8Array(buf);
  }

  writeUint(value: number, width: number, bigEndian?: boolean): void {
    const buf = new Uint8Array(width);
    encodeInt(buf, 0, value, width, bigEndian);
    this.writeUint8Array(buf);
  }

  writeUint16(value: number, bigEndian?: boolean): void { return this.writeUint(value, 2, bigEndian); }
  writeUint24(value: number, bigEndian?: boolean): void { return this.writeUint(value, 3, bigEndian); }
  writeUint32(value: number, bigEndian?: boolean): void { return this.writeUint(value, 4, bigEndian); }
  writeUint40(value: number, bigEndian?: boolean): void { return this.writeUint(value, 5, bigEndian); }
  writeUint48(value: number, bigEndian?: boolean): void { return this.writeUint(value, 6, bigEndian); }
  writeUint56(value: number, bigEndian?: boolean): void { return this.writeUint(value, 7, bigEndian); }
  writeUint64(value: number, bigEndian?: boolean): void { return this.writeUint(value, 8, bigEndian); }
}

export class InMemoryBytesWriter extends AbstractBytesWriter {
  private buffer: Uint8Array;
  private offset: number;

  constructor(bufSize: number = 8192) {
    super();
    this.buffer = new Uint8Array(bufSize);
    this.offset = 0;
  }

  flush(): void {
    // no-op
  }

  reset(): void {
    this.offset = 0;
  }

  slice(): Uint8Array {
    return this.buffer.slice(0, this.offset);
  }

  size(): number {
    return this.offset;
  }

  writeUint8(value: number): void {
    const offset = this.offset++;
    if (offset >= this.buffer.length) {
      this.grow(offset + 64);
    }
    this.buffer[offset] = value;
  }

  writeUint8Array(value: Uint8Array | ArrayLike<number> | number[]): void {
    this.ensureSpace(value.length);
    this.buffer.set(value, this.offset);
    this.offset += value.length;
  }

  ensureSpace(size: number): void {
    const requiredLength = this.offset + size;
    if (requiredLength <= this.buffer.length) return;

    // resize with 64bytes alignment
    this.grow((requiredLength + (64 - 1)) & -64);
  }

  grow(newLength: number): void {
    const newBuffer = new Uint8Array(newLength);
    newBuffer.set(this.buffer);
    this.buffer = newBuffer;
  }
}

// ==============================================================================================================
export class YajbeEncoder extends DataEncoder {
  private readonly fieldNameWriter: FieldNameWriter;
  private readonly textEncoder: TextEncoder;
  private readonly writer: BytesWriter;

  constructor(writer: BytesWriter, initialFieldNames?: string[]) {
    super();
    this.textEncoder = new TextEncoder();
    this.fieldNameWriter = new FieldNameWriter(writer, this.textEncoder, initialFieldNames);
    this.writer = writer;
  }

  flush(): void {
    this.writer.flush();
  }

  protected encodeFloat(value: number): void {
    // assume float64
    this.writer.writeUint8(0b00000_110);
    this.writer.writeFloat64(value);
  }

  protected encodeInteger(value: number): void {
    if (value > 0) {
      this.encodePositiveInt(value);
    } else {
      this.encodeNegativeInt(value);
    }
  }

  protected encodePositiveInt(value: number): void {
    if (value <= 24) {
      this.writer.writeUint8(0b010_00000 | (value - 1));
    } else {
      value -= 25;
      const bytes = intBytesWidth(value);
      this.writer.writeUint8(0b010_00000 | (23 + bytes));
      this.writer.writeUint(value, bytes);
    }
  }

  protected encodeNegativeInt(value: number): void {
    value = -value;
    if (value <= 23) {
      this.writer.writeUint8(0b011_00000 | value);
    } else {
      value -= 24;
      const bytes = intBytesWidth(value);
      this.writer.writeUint8(0b011_00000 | (23 + bytes));
      this.writer.writeUint(value, bytes);
    }
  }

  protected encodeObject(dict: {[key: string]: unknown}): void {
    const keys = Object.keys(dict);
    //keys.sort();
    this.writeLength(0b0011_0000, 10, keys.length);
    for (let i = 0; i < keys.length; ++i) {
      const key = keys[i];
      this.fieldNameWriter.encodeString(key);
      this.encodeItem(dict[key]);
    }
  }

  protected encodeArray(value: ArrayLike<unknown>): void {
    this.writeLength(0b0010_0000, 10, value.length);
    for (let i = 0; i < value.length; ++i) {
      this.encodeItem(value[i]);
    }
  }

  protected encodeUint8Array(value: Uint8Array): void {
    this.writeLength(0b10_000000, 59, value.length);
    this.writer.writeUint8Array(value);
  }

  protected encodeString(value: string): void {
    const utf8data = this.textEncoder.encode(value);
    this.writeLength(0b11_000000, 59, utf8data.length);
    this.writer.writeUint8Array(utf8data);
  }

  private writeLength(head: number, inlineMax: number, length: number): void {
    if (length <= inlineMax) {
      this.writer.writeUint8(head | length);
    } else {
      const deltaLength = length - inlineMax;
      const bytes = intBytesWidth(deltaLength);
      this.writer.writeUint8(head | (inlineMax + bytes));
      this.writer.writeUint(deltaLength, bytes);
    }
  }

  protected encodeNull(): void { this.writer.writeUint8(0); }
  protected encodeUndefined(): void { this.writer.writeUint8(0); }
  protected encodeTrue(): void { this.writer.writeUint8(0b11); }
  protected encodeFalse(): void { this.writer.writeUint8(0b10); }
}

export class YajbeDecoder {
  private readonly fieldNameReader: FieldNameReader;
  private readonly textDecoder: TextDecoder;
  private readonly buffer: BytesReader;

  constructor(buffer: BytesReader, initialFieldNames?: string[]) {
    this.textDecoder = new TextDecoder();
    this.fieldNameReader = new FieldNameReader(buffer, this.textDecoder, initialFieldNames);
    this.buffer = buffer;
  }

  decodeItem(): unknown {
    const head = this.buffer.readUint8();
    if ((head & 0b11_000000) == 0b11_000000) {
      return this.decodeString(head);
    } else if ((head & 0b10_000000) == 0b10_000000) {
      return this.decodeBytes(head);
    } else if ((head & 0b010_00000) == 0b010_00000) {
      return this.decodeInt(head);
    } else if ((head & 0b0011_0000) == 0b0011_0000) {
      return this.decodeObject(head);
    } else if ((head & 0b0010_0000) == 0b0010_0000) {
      return this.decodeArray(head);
    } else if ((head & 0b000001_00) == 0b000001_00) {
      return this.decodeFloat(head);
    } else switch (head) {
      // null
      case 0b00000000: return null;
      // boolean
      case 0b00000010: return false;
      case 0b00000011: return true;
    }
    throw new Error('unsupported head ' + head);
  }

  private decodeInt(head: number): number {
    const signed = (head & 0b011_00000) == 0b011_00000;

    const w = head & 0b11111;
    if (w < 24) return signed ? -w : (1 + w);

    const value = this.buffer.readUint(w - 23);
    return signed ? -(value + 24) : (value + 25);
  }

  private decodeFloat(head: number): number {
    switch (head & 0b11) {
      case 0b00: return this.buffer.readFloat16();
      case 0b01: return this.buffer.readFloat32();
      case 0b10: return this.buffer.readFloat64();
      case 0b11: throw new Error('decode bigdecimal');
    }
    return 0;
  }

  private readBytesLength(head: number): number {
    const w = head & 0b111111;
    if (w <= 59) return w;
    return 59 + this.buffer.readUint(w - 59);
  }

  private decodeBytes(head: number): Uint8Array {
    const length = this.readBytesLength(head);
    return this.buffer.readUint8Array(length);
  }

  private decodeString(head: number): string {
    const buffer = this.decodeBytes(head);
    return this.textDecoder.decode(buffer);
  }

  private readHasMore(): boolean {
    if (this.buffer.peekUint8() !== 0b00000001) {
      return true;
    }
    this.buffer.readUint8();
    return false;
  }

  private readItemCount(w: number): number {
    if (w <= 10) return w;
    return 10 + this.buffer.readUint(w - 10);
  }

  private decodeArray(head: number): unknown[] | Array<unknown> {
    const w = head & 0b1111;
    if (w == 0b1111) {
      const retArray: unknown[] = [];
      while (this.readHasMore()) {
        retArray.push(this.decodeItem());
      }
      return retArray;
    }

    const length = this.readItemCount(w);
    const retArray = new Array(length);
    for (let i = 0; i < length; ++i) {
      retArray[i] = this.decodeItem();
    }
    return retArray;
  }

  private decodeObject(head: number): {[key: string]: unknown} {
    const w = head & 0b1111;
    if (w == 0b1111) {
      const retObject: {[key: string]: unknown} = {};
      while (this.readHasMore()) {
        const key = this.fieldNameReader.decodeString();
        retObject[key] = this.decodeItem();
      }
      return retObject;
    }

    const length = this.readItemCount(w);
    const retObject: {[key: string]: unknown} = {};
    for (let i = 0; i < length; ++i) {
      const key = this.fieldNameReader.decodeString();
      retObject[key] = this.decodeItem();
    }
    return retObject;
  }
}

export class FieldNameWriter {
  private readonly indexedMap = new Map<string, number>();
  private readonly textEncoder: TextEncoder;
  private readonly writer: BytesWriter;
  private lastKey?: Uint8Array;

  constructor(writer: BytesWriter, textEncoder: TextEncoder, initialFieldNames?: string[]) {
    this.writer = writer;
    this.textEncoder = textEncoder;

    if (initialFieldNames) {
      for (let i = 0; i < initialFieldNames.length && i < 65819; ++i) {
        this.indexedMap.set(initialFieldNames[i], i);
      }
    }
  }

  encodeString(key: string): void {
    const utf8 = this.textEncoder.encode(key);

    const index = this.indexedMap.get(key);
    if (index != null) {
      this.writeIndexedFieldName(index);
      this.lastKey = utf8;
      return;
    }

    if (this.lastKey && utf8.length > 4) {
      const prefix = Math.min(0xff, this.prefix(utf8));
      const suffix = this.suffix(utf8, prefix);

      if (suffix > 2) {
        this.writePrefixSuffix(utf8, prefix, Math.min(0xff, suffix));
      } else if (prefix > 2) {
        this.writePrefix(utf8, prefix);
      } else {
        this.writeFullFieldName(utf8);
      }
    } else {
      this.writeFullFieldName(utf8);
    }

    if (this.indexedMap.size < 65819) {
      this.indexedMap.set(key, this.indexedMap.size);
    }
    this.lastKey = utf8;
  }

  private writeFullFieldName(fieldName: Uint8Array | number[]) {
    // 100----- Full Field Name (0-29 length - 1, 30 1b-len, 31 2b-len)
    this.writeLength(0b100_00000, fieldName.length);
    this.writer.writeUint8Array(fieldName);
  }

  private writeIndexedFieldName(fieldIndex: number) {
    // 101----- Field Offset (0-29 field, 30 1b-len, 31 2b-len)
    this.writeLength(0b101_00000, fieldIndex);
  }

  private writePrefix(fieldName: Uint8Array, prefix: number) {
    // 110----- Prefix (1byte prefix, 0-29 length - 1, 30 1b-len, 31 2b-len)
    const writer = this.writer;
    const length = fieldName.length - prefix;
    this.writeLength(0b110_00000, length);
    writer.writeUint8(prefix);
    writer.writeUint8Array(fieldName.slice(prefix));
  }

  private writePrefixSuffix(fieldName: Uint8Array, prefix: number, suffix: number) {
    // 111----- Prefix/Suffix (1byte prefix, 1byte suffix, 0-29 length - 1, 30 1b-len, 31 2b-len)
    const writer = this.writer;
    const length = fieldName.length - prefix - suffix;
    this.writeLength(0b111_00000, length);
    writer.writeUint8(prefix);
    writer.writeUint8(suffix);
    writer.writeUint8Array(fieldName.slice(prefix, fieldName.length - suffix));
  }

  private writeLength(head: number, length: number) {
    const writer = this.writer;
    if (length < 30) {
      writer.writeUint8(head | length);
    } else if (length <= 284) {
      writer.writeUint8(head | 0b11110);
      writer.writeUint8((length - 29) & 0xff);
    } else if (length <= 65819) {
      writer.writeUint8(head | 0b11111);
      writer.writeUint8(Math.floor((length - 284) / 256));
      writer.writeUint8((length - 284) & 255);
    } else {
      throw new Error("unexpected too many field names: " + length);
    }
  }

  private prefix(key: Uint8Array): number {
    const a = this.lastKey!;
    const b = key;
    const len = Math.min(a.length, b.length);
    for (let i = 0; i < len; ++i) {
      if (a[i] != b[i]) {
        return i;
      }
    }
    return len;
  }


  private suffix(key: Uint8Array, kPrefix: number): number {
    const a = this.lastKey!;
    const b = key;
    const bLen = b.length - kPrefix;
    const len = Math.min(a.length, bLen);
    for (let i = 1; i <= len; ++i) {
      if ((a[a.length - i] & 0xff) != (b[kPrefix + (bLen - i)] & 0xff)) {
        return i - 1;
      }
    }
    return len;
  }
}

export class FieldNameReader {
  private readonly indexedNames: Uint8Array[] = [];
  private readonly textDecoder: TextDecoder;
  private readonly reader: BytesReader;

  private lastKey: Uint8Array = new Uint8Array(0);

  constructor(reader: BytesReader, textDecoder: TextDecoder, initialFieldNames?: string[]) {
    this.reader = reader;
    this.textDecoder = textDecoder;

    if (initialFieldNames) {
      const textEncoder = new TextEncoder();
      for (let i = 0; i < initialFieldNames.length && i < 65819; ++i) {
        this.indexedNames.push(textEncoder.encode(initialFieldNames[i]));
      }
    }
  }

  decodeString(): string {
    const head = this.reader.readUint8();
    switch ((head >> 5) & 0b111) {
      case 0b100: return this.readFullFieldName(head);
      case 0b101: return this.readIndexedFieldName(head);
      case 0b110: return this.readPrefix(head);
      case 0b111: return this.readPrefixSuffix(head);
      default: throw new Error("unexpected head: " + head.toString(2));
    }
  }

  private readLength(head: number) {
    const length = (head & 0b000_11111);
    if (length < 30) return length;
    if (length == 30) return this.reader.readUint8() + 29;

    const b1 = this.reader.readUint8();
    const b2 = this.reader.readUint8();
    return 284 + 256 * b1 + b2;
  }

  private addToIndex(utf8: Uint8Array): string {
    this.indexedNames.push(utf8);
    this.lastKey = utf8;
    return this.textDecoder.decode(utf8);
  }

  private readFullFieldName(head: number): string {
    const length = this.readLength(head);
    const utf8 = this.reader.readUint8Array(length);
    return this.addToIndex(utf8);
  }

  private readIndexedFieldName(head: number): string {
    const fieldIndex = this.readLength(head);
    const utf8 = this.indexedNames[fieldIndex];
    this.lastKey = utf8;
    return this.textDecoder.decode(utf8);
  }

  private readPrefix(head: number): string {
    const length = this.readLength(head);
    const prefix = this.reader.readUint8();
    const kpart = this.reader.readUint8Array(length);
    const utf8 = new Uint8Array(prefix + length);
    utf8.set(this.lastKey.slice(0, prefix));
    utf8.set(kpart, prefix);
    return this.addToIndex(utf8);
  }

  private readPrefixSuffix(head: number): string {
    const length = this.readLength(head);
    const prefix = this.reader.readUint8();
    const suffix = this.reader.readUint8();
    const kpart = this.reader.readUint8Array(length);
    const utf8 = new Uint8Array(prefix + length + suffix);
    utf8.set(this.lastKey.slice(0, prefix));
    utf8.set(kpart, prefix);
    utf8.set(this.lastKey.slice(this.lastKey.length - suffix), prefix + length);
    return this.addToIndex(utf8);
  }
}
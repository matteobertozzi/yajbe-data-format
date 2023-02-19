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

package io.github.matteobertozzi.yajbe;

import java.io.IOException;
import java.nio.charset.StandardCharsets;
import java.util.Arrays;

final class YajbeFieldNameWriter {
  private final IndexedHashSet indexedMap = new IndexedHashSet(128);
  private final YajbeWriter stream;
  private byte[] lastKey;

  public YajbeFieldNameWriter(final YajbeWriter stream) {
    this.stream = stream;
  }

  void setInitialFieldNames(final String[] names) {
    if (indexedMap.size != 0) {
      throw new UnsupportedOperationException("field names already added");
    }

    for (int i = 0; i < names.length && i < 0xffff; ++i) {
      indexedMap.add(names[i]);
    }
  }

  public void write(final String key) throws IOException {
    final byte[] utf8 = key.getBytes(StandardCharsets.UTF_8);

    final int index = this.indexedMap.get(key);
    if (index >= 0) {
      this.writeIndexedFieldName(index);
      this.lastKey = utf8;
      return;
    }

    if (this.lastKey != null && utf8.length > 4) {
      final int prefix = Math.min(0xff, this.prefix(utf8));
      final int suffix = this.suffix(utf8, prefix);

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

    if (indexedMap.size() < 0xffff) {
      indexedMap.add(key);
    }
    this.lastKey = utf8;
  }

  public void writeFullFieldName(final byte[] fieldName) throws IOException {
    // 100----- Full Field Name (0-29 length - 1, 30 1b-len, 31 2b-len)
    //System.out.println(" -> WRITE FULL " + new String(fieldName));
    writeLength(0b100_00000, fieldName.length);
    this.stream.write(fieldName, 0, fieldName.length);
  }

  public void writeIndexedFieldName(final int fieldIndex) throws IOException {
    // 101----- Field Offset (0-29 field, 30 1b-len, 31 2b-len)
    this.writeLength(0b101_00000, fieldIndex);
  }

  public void writePrefix(final byte[] fieldName, final int prefix) throws IOException {
    // 110----- Prefix (1byte prefix, 0-29 length - 1, 30 1b-len, 31 2b-len)
    final int length = fieldName.length - prefix;
    this.writeLength(0b110_00000, length);
    stream.write(prefix);
    stream.write(fieldName, prefix, length);
  }

  public void writePrefixSuffix(final byte[] fieldName, final int prefix, final int suffix) throws IOException {
    // 111----- Prefix/Suffix (1byte prefix, 1byte suffix, 0-29 length - 1, 30 1b-len, 31 2b-len)
    final int length = fieldName.length - prefix - suffix;
    this.writeLength(0b111_00000, length);
    stream.write(prefix);
    stream.write(suffix);
    stream.write(fieldName, prefix, length);
  }

  private void writeLength(final int head, final int length) throws IOException {
    if (length < 30) {
      stream.write(head | length);
    } else if (length <= 0xff) {
      final byte[] buf = stream.rawBuffer();
      final int bufOff = stream.rawBufferOffset(2);
      buf[bufOff] = (byte) (head | 0b11110);
      buf[bufOff + 1] = (byte) (length & 0xff);
    } else if (length <= 0xffff) {
      final byte[] buf = stream.rawBuffer();
      final int bufOff = stream.rawBufferOffset(3);
      buf[bufOff] = (byte) (head | 0b11111);
      YajbeWriter.writeFixed(buf, bufOff + 1, length, 2);
    } else {
      throw new Error("unexpected too many field names: " + length);
    }
  }

  private int prefix(final byte[] key) {
    final byte[] a = this.lastKey;
    final int len = Math.min(a.length, key.length);
    for (int i = 0; i < len; ++i) {
      if (a[i] != key[i]) {
        return i;
      }
    }
    return len;
  }

  private int suffix(final byte[] key, final int kPrefix) {
    final byte[] a = this.lastKey;
    final int bLen = key.length - kPrefix;
    final int len = Math.min(a.length, bLen);
    for (int i = 1; i <= len; ++i) {
      if ((a[a.length - i] & 0xff) != (key[kPrefix + (bLen - i)] & 0xff)) {
        return i - 1;
      }
    }
    return len;
  }

  private static final class IndexedHashSet {
    private String[] values;
    private int[] table; // hash/next
    private int[] buckets;
    private int size;

    public IndexedHashSet(final int estimateSize) {
      this.values = new String[estimateSize];
      this.table = new int[estimateSize * 2];
      this.buckets = new int[tableSizeForItems(estimateSize)];
      this.size = 0;
      Arrays.fill(buckets, -1);
    }

    public int size() {
      return size;
    }

    public void add(final String key) {
      if (size == values.length) {
        resize();
      }

      final int keyIndex = size++;
      final int keyHash = hash(key);
      final int targetBucket = keyHash & (buckets.length - 1);
      final int itemIndex = keyIndex << 1;
      values[keyIndex] = key;
      table[itemIndex] = keyHash;
      table[itemIndex + 1] = buckets[targetBucket];
      buckets[targetBucket] = keyIndex;
    }

    public int get(final String key) {
      final int hash = hash(key);
      int index = buckets[hash & (buckets.length - 1)];
      while (index >= 0) {
        final int itemIndex = (index << 1);
        if (hash == table[itemIndex] && key.equals(values[index])) {
          return index;
        }
        index = table[itemIndex + 1];
      }
      return -1;
    }

    private void resize() {
      this.values = Arrays.copyOf(this.values, this.values.length << 1);
      this.table = Arrays.copyOf(this.table, this.table.length << 1);

      final int newBucketsCount = tableSizeForItems(values.length);
      //System.out.println("table " + newBucketsCount + "/" + values.length);
      if (newBucketsCount == buckets.length) return;

      final int[] newBuckets = new int[newBucketsCount];
      Arrays.fill(newBuckets, -1);
      final int mask = newBucketsCount - 1;
      for (int i = 0, itemIndex = 0; i < size; ++i, itemIndex += 2) {
        final int targetBucket = table[itemIndex] & mask;
        table[itemIndex + 1] = newBuckets[targetBucket];
        newBuckets[targetBucket] = i;
      }
      this.buckets = newBuckets;
    }

    private static int tableSizeForItems(final int expectedItems) {
      return 1 << (Integer.SIZE - Integer.numberOfLeadingZeros((expectedItems * 2) - 1));
    }

    private static int hash(final String key) {
      final int h = key.hashCode();
      return (h ^ (h >>> 16)) & 0x7fffffff;
    }
  }
}

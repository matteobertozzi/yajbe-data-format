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

import io.github.matteobertozzi.yajbe.YajbeReader.ByteArraySlice;

final class YajbeFieldNameReader {
  private final YajbeReader reader;

  private Object[] indexedNames = new Object[32];
  private int indexedNameCount = 0;
  private ByteArraySlice lastKey;

  public YajbeFieldNameReader(final YajbeReader reader) {
    this.reader = reader;
  }

  void setInitialFieldNames(final String[] names) {
    if (indexedNameCount != 0) {
      throw new UnsupportedOperationException("field names already added");
    }

    if (indexedNames.length <= (names.length * 2)) {
      indexedNames = Arrays.copyOf(names, names.length * 2);
    }

    for (int i = 0; i < names.length; ++i) {
      indexedNames[indexedNameCount++] = new ByteArraySlice(names[i].getBytes(StandardCharsets.UTF_8));
      indexedNames[indexedNameCount++] = names[i];
    }
  }

  public String read() throws IOException {
    final int head = this.reader.read();
    return switch ((head >> 5) & 0b111) {
      case 0b100 -> this.readFullFieldName(head);
      case 0b101 -> this.readIndexedFieldName(head);
      case 0b110 -> this.readPrefix(head);
      case 0b111 -> this.readPrefixSuffix(head);
      default -> throw new Error("unexpected head: " + Integer.toBinaryString(head));
    };
  }

  private int readLength(final int head) throws IOException {
    final int length = (head & 0b000_11111);
    return switch (length) {
      case 30 -> reader.read();
      case 31 -> reader.readFixedInt(2);
      default -> length;
    };
  }

  private String addToIndex(final ByteArraySlice utf8) {
    if (indexedNameCount == indexedNames.length) {
      indexedNames = Arrays.copyOf(indexedNames, indexedNameCount << 1);
    }

    final String str = utf8.toString(StandardCharsets.UTF_8);
    indexedNames[indexedNameCount++] = utf8;
    indexedNames[indexedNameCount++] = str;

    this.lastKey = utf8;
    return str;
  }

  private String readFullFieldName(final int head) throws IOException {
    final int length = this.readLength(head);
    final ByteArraySlice utf8 = this.reader.readNBytes(length);
    return addToIndex(utf8);
  }

  private String readIndexedFieldName(final int head) throws IOException {
    final int fieldIndex = this.readLength(head) << 1;
    this.lastKey = (ByteArraySlice) this.indexedNames[fieldIndex];
    return (String) this.indexedNames[fieldIndex + 1];
  }

  private String readPrefix(final int head) throws IOException {
    final int length = this.readLength(head);
    final int prefix = this.reader.read();

    final byte[] utf8 = new byte[prefix + length];
    System.arraycopy(this.lastKey.buf(), this.lastKey.off(), utf8, 0, prefix);
    this.reader.readNBytes(utf8, prefix, length);
    return addToIndex(new ByteArraySlice(utf8));
  }

  private String readPrefixSuffix(final int head) throws IOException {
    final int length = this.readLength(head);
    final int prefix = this.reader.read();
    final int suffix = this.reader.read();

    final byte[] utf8 = new byte[prefix + length + suffix];
    System.arraycopy(this.lastKey.buf(), this.lastKey.off(), utf8, 0, prefix);
    System.arraycopy(lastKey.buf(), lastKey.off() + lastKey.len() - suffix, utf8, prefix + length, suffix);
    this.reader.readNBytes(utf8, prefix, length);

    return addToIndex(new ByteArraySlice(utf8));
  }
}

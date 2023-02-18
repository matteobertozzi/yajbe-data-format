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

import java.nio.charset.StandardCharsets;

final class YajbeReaderByteArray extends YajbeReader {
  private final byte[] data;
  private final int length;
  private int offset;

  public YajbeReaderByteArray(final byte[] data, final int offset, final int len) {
    this.data = data;
    this.offset = offset;
    this.length = len;
  }

  protected int peek() {
    return (offset < length) ? (data[offset] & 0xff) : -1;
  }

  protected int read() {
    return (offset < length) ? (data[offset++] & 0xff) : -1;
  }

  protected ByteArraySlice readNBytes(final int n) {
    final ByteArraySlice slice = new ByteArraySlice(data, offset, n);
    offset += n;
    return slice;
  }

  protected void readNBytes(final byte[] buf, final int off, final int len) {
    System.arraycopy(data, offset, buf, off, len);
    offset += len;
  }

  protected String readString(final int n) {
    final String r = new String(data, offset, n, StandardCharsets.UTF_8);
    offset += n;
    return r;
  }

  protected long readFixed(final int width) {
    final int off = this.offset;
    this.offset += width;
    return readFixed(data, off, width);
  }

  protected int readFixedInt(final int width) {
    final int off = this.offset;
    this.offset += width;
    return readFixedInt(data, off, width);
  }
}

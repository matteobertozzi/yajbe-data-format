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

import java.io.BufferedInputStream;
import java.io.ByteArrayInputStream;
import java.io.IOException;
import java.io.InputStream;
import java.nio.charset.StandardCharsets;

final class YajbeReaderStream extends YajbeReader {
  private static final ByteArraySlice EMPTY_BYTES = new ByteArraySlice(null, 0, 0);
  private static final String EMPTY_STRING = "";

  private final byte[] buf8 = new byte[8];
  private final InputStream stream;

  public YajbeReaderStream(final InputStream in) {
    if (in instanceof final ByteArrayInputStream bytesIn) {
      this.stream = bytesIn;
    } else if (in instanceof final BufferedInputStream bufIn) {
      this.stream = bufIn;
    } else {
      this.stream = new BufferedInputStream(in);
    }
  }

  @Override
  protected int peek() throws IOException {
    stream.mark(1);
    final int b = stream.read();
    stream.reset();
    return b;
  }

  @Override
  protected int read() throws IOException {
    return stream.read();
  }

  @Override
  protected String readString(final int n) throws IOException {
    if (n == 0) return EMPTY_STRING;
    final byte[] buf = stream.readNBytes(n);
    return new String(buf, StandardCharsets.UTF_8);
  }

  @Override
  protected ByteArraySlice readNBytes(final int n) throws IOException {
    if (n == 0) return EMPTY_BYTES;
    final byte[] buf = stream.readNBytes(n);
    return new ByteArraySlice(buf);
  }

  @Override
  protected void readNBytes(final byte[] buf, final int off, final int len) throws IOException {
    if (stream.readNBytes(buf, off, len) != len) {
      throw new IOException("unable to read " + len + " bytes from the stream");
    }
  }

  @Override
  protected long readFixed(final int width) throws IOException {
    readNBytes(buf8, 0, width);
    return readFixed(buf8, 0, width);
  }

  @Override
  protected int readFixedInt(final int width) throws IOException {
    readNBytes(buf8, 0, width);
    return readFixedInt(buf8, 0, width);
  }
}

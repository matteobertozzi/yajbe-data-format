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
import java.io.OutputStream;

final class YajbeWriterStream extends YajbeWriter {
  private final OutputStream stream;
  private final byte[] wbuf;
  private int wbufOff;

  public YajbeWriterStream(final OutputStream stream, final byte[] buffer) {
    this.stream = stream;
    this.wbuf = buffer;
    this.wbufOff = 0;
  }

  @Override
  public void flush() throws IOException {
    rawBufferFlush();
    stream.flush();
  }

  @Override
  protected void write(final int v) throws IOException {
    if (wbufOff == wbuf.length) {
      rawBufferFlush();
    }

    wbuf[wbufOff++] = (byte)v;
  }

  @Override
  protected void write(final byte[] buf, final int off, final int len) throws IOException {
    if (len >= wbuf.length) {
      rawBufferFlush();
      stream.write(buf, off, len);
      return;
    }

    if (len > (wbuf.length - wbufOff)) {
      rawBufferFlush();
    }
    System.arraycopy(buf, off, wbuf, wbufOff, len);
    wbufOff += len;
  }

  @Override
  protected final byte[] rawBuffer() {
    return wbuf;
  }

  @Override
  protected final int rawBufferOffset() {
    return wbufOff;
  }

  @Override
  protected final int rawBufferOffset(final int size) throws IOException {
    if ((wbufOff + size) >= wbuf.length) {
      stream.write(wbuf, 0, wbufOff);
      wbufOff = size;
      return 0;
    }

    final int offset = wbufOff;
    wbufOff += size;
    return offset;
  }

  @Override
  protected final void rawBufferFlush(final int length, final int availSizeRequired) throws IOException {
    if ((wbuf.length - length) < availSizeRequired) {
      stream.write(wbuf, 0, length);
      wbufOff = 0;
    } else {
      wbufOff = length;
    }
  }

  private void rawBufferFlush() throws IOException {
    if (wbufOff != 0) {
      stream.write(wbuf, 0, wbufOff);
      wbufOff = 0;
    }
  }
}

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
import java.math.BigDecimal;
import java.math.BigInteger;
import java.util.Arrays;

import com.fasterxml.jackson.core.Base64Variant;
import com.fasterxml.jackson.core.JsonGenerator;
import com.fasterxml.jackson.core.ObjectCodec;
import com.fasterxml.jackson.core.base.GeneratorBase;
import com.fasterxml.jackson.core.io.IOContext;

/**
 * {@link JsonGenerator} implementation that writes YAJBE encoded content.
 */
final class YajbeGenerator extends GeneratorBase {
  private final YajbeFieldNameWriter fileNameWriter;
  private final YajbeWriter stream;
  private final IOContext ctxt;
  private final byte[] wbuffer;

  YajbeGenerator(final IOContext ctxt, final int features, final ObjectCodec codec, final OutputStream stream) {
    super(features, codec);
    this.ctxt = ctxt;

    this.wbuffer = ctxt.allocWriteEncodingBuffer(9);
    this.stream = YajbeWriter.forBufferedStream(stream, wbuffer);
    this.fileNameWriter = new YajbeFieldNameWriter(this.stream);
  }

  void setInitialFieldNames(final String[] names) {
    fileNameWriter.setInitialFieldNames(names);
  }

  @Override
  public void close() throws IOException {
    flush();
    _releaseBuffers();
    super.close();
  }

  @Override
  public void flush() throws IOException {
    stream.flush();
  }

  @Override
  protected void _releaseBuffers() {
    ctxt.releaseWriteEncodingBuffer(wbuffer);
  }

  @Override
  protected void _verifyValueWrite(final String typeMsg) {
    // TODO Auto-generated method stub
  }

  private boolean[] stackBlocks = new boolean[32]; // it can be a bitset (eof required true/false)
  private int stackSize = 0;

  private void openBlock(final boolean eofRequired) {
    if (stackSize == stackBlocks.length) {
      stackBlocks = Arrays.copyOf(stackBlocks, stackSize + 16);
    }
    stackBlocks[stackSize++] = eofRequired;
  }

  private void closeBlock() throws IOException {
    if (stackBlocks[--stackSize]) {
      stream.writeEof();
    }
  }

  @Override
  public void writeStartArray() throws IOException {
    openBlock(stream.newArray());
  }

  @Override
  public void writeStartArray(final Object forValue, final int size) throws IOException {
    setCurrentValue(forValue);
    openBlock(stream.newArray(size));
  }

  @Override
  public void writeEndArray() throws IOException {
    closeBlock();
  }

  @Override
  public void writeArray(final int[] array, final int offset, final int length) throws IOException {
    stream.writeArray(array, offset, length);
  }

  @Override
  public void writeArray(final long[] array, final int offset, final int length) throws IOException {
    stream.writeArray(array, offset, length);
  }

  @Override
  public void writeStartObject() throws IOException {
    openBlock(stream.newObject());
  }

  @Override
  public void writeStartObject(final Object forValue, final int size) throws IOException {
    setCurrentValue(forValue);
    openBlock(stream.newObject(size));
  }

  @Override
  public void writeEndObject() throws IOException {
    closeBlock();
  }

  @Override
  public void writeFieldName(final String name) throws IOException {
    fileNameWriter.write(name);
  }

  @Override
  public void writeString(final String text) throws IOException {
    if (text == null || text.isEmpty()) {
      stream.writeEmptyString();
      return;
    }

    stream.writeString(text);
  }

  @Override
  public void writeString(final char[] buffer, final int offset, final int len) throws IOException {
    if (len != 0) {
      final String text = new String(buffer, offset, len);
      stream.writeString(text);
    } else {
      stream.writeEmptyString();
    }
  }

  @Override
  public void writeRawUTF8String(final byte[] buffer, final int offset, final int len) {
    throw new UnsupportedOperationException();
  }

  @Override
  public void writeUTF8String(final byte[] buffer, final int offset, final int len) throws IOException {
    if (len != 0) {
      stream.writeUtf8(buffer, offset, len);
    } else {
      stream.writeEmptyString();
    }
  }

  @Override
  public void writeRaw(final String text) {
    throw new UnsupportedOperationException();
  }

  @Override
  public void writeRaw(final String text, final int offset, final int len) {
    throw new UnsupportedOperationException();
  }

  @Override
  public void writeRaw(final char[] text, final int offset, final int len) {
    throw new UnsupportedOperationException();
  }

  @Override
  public void writeRaw(final char c) {
    throw new UnsupportedOperationException();
  }

  @Override
  public void writeBinary(final Base64Variant bv, final byte[] data, final int offset, final int len) throws IOException {
    stream.writeBytes(data, offset, len);
  }

  @Override
  public void writeNumber(final int v) throws IOException {
    stream.writeInt(v);
  }

  @Override
  public void writeNumber(final long v) throws IOException {
    stream.writeInt(v);
  }

  @Override
  public void writeNumber(final BigInteger v) throws IOException {
    if (v != null) {
      stream.writeBigInteger(v);
    } else {
      stream.writeNull();
    }
  }

  @Override
  public void writeNumber(final float v) throws IOException {
    stream.writeFloat32(v);
  }

  @Override
  public void writeNumber(final double v) throws IOException {
    stream.writeFloat64(v);
  }

  @Override
  public void writeNumber(final BigDecimal v) throws IOException {
    if (v != null) {
      stream.writeBigDecimal(v);
    } else {
      stream.writeNull();
    }
  }

  @Override
  public void writeNumber(final String encodedValue) {
    throw new UnsupportedOperationException();
  }

  @Override
  public void writeBoolean(final boolean state) throws IOException {
    stream.writeBool(state);
  }

  @Override
  public void writeNull() throws IOException {
    stream.writeNull();
  }
}

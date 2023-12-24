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
import java.nio.charset.StandardCharsets;

import io.github.matteobertozzi.yajbe.YajbeEnumMapping.YajbeEnumLruMappingConfig;
import io.github.matteobertozzi.yajbe.YajbeEnumMapping.YajbeEnumMappingConfig;

abstract class YajbeWriter {
  @FunctionalInterface
  public interface RawBufferWriter {
    int writeItem(byte[] buf, int off, int itemIndex);
  }

  protected abstract void flush() throws IOException;
  protected abstract void write(int v) throws IOException;
  protected abstract void write(byte[] buf, int off, int len) throws IOException;

  protected abstract byte[] rawBuffer();
  protected abstract int rawBufferOffset();
  protected abstract int rawBufferOffset(int size) throws IOException;
  protected abstract void rawBufferFlush(int length, final int availSizeRequired) throws IOException;

  public static YajbeWriter forBufferedStream(final OutputStream stream, final byte[] buffer) {
    return new YajbeWriterStream(stream, buffer);
  }

  // =========================================================================================================
  @SuppressWarnings("fallthrough")
  public static void writeFixed(final byte[] buf, final int off, final long v, final int width) {
    switch (width) {
      case 8: buf[off + 7] = ((byte)((v >>> 56) & 0xff));
      case 7: buf[off + 6] = ((byte)((v >>> 48) & 0xff));
      case 6: buf[off + 5] = ((byte)((v >>> 40) & 0xff));
      case 5: buf[off + 4] = ((byte)((v >>> 32) & 0xff));
      case 4: buf[off + 3] = ((byte)((v >>> 24) & 0xff));
      case 3: buf[off + 2] = ((byte)((v >>> 16) & 0xff));
      case 2: buf[off + 1] = ((byte)((v >>> 8) & 0xff));
      case 1: buf[off] = (byte)(v & 0xff);
    }
  }

  @SuppressWarnings("fallthrough")
  public static void writeFixed(final byte[] buf, final int off, final int v, final int width) {
    switch (width) {
      case 4: buf[off + 3] = ((byte)((v >>> 24) & 0xff));
      case 3: buf[off + 2] = ((byte)((v >>> 16) & 0xff));
      case 2: buf[off + 1] = ((byte)((v >>> 8) & 0xff));
      case 1: buf[off] = (byte)(v & 0xff);
    }
  }
  // =========================================================================================================

  public final void writeNull() throws IOException {
    write(0);
  }

  public final void writeEof() throws IOException {
    write(1);
  }

  public final void writeBool(final boolean value) throws IOException {
    write((value ? 0b11 : 0b10));
  }

  // ====================================================================================================
  //  Float related
  // ====================================================================================================
  public final void writeFloat32(final float v) throws IOException {
    final byte[] buf = rawBuffer();
    final int bufOff = rawBufferOffset(5);
    buf[bufOff] = 0b00000_101;
    writeFixed(buf, bufOff + 1, Float.floatToIntBits(v), 4);
  }

  public final void writeFloat64(final double v) throws IOException {
    final byte[] buf = rawBuffer();
    final int bufOff = rawBufferOffset(9);
    buf[bufOff] = 0b00000_110;
    writeFixed(buf, bufOff + 1, Double.doubleToLongBits(v), 8);
  }

  public final void writeBigDecimal(final BigDecimal v) throws IOException {
    writeBigDecimal(v.scale(), v.precision(), v.unscaledValue());
  }

  public final void writeBigInteger(final BigInteger v) throws IOException {
    writeBigDecimal(0, 0, v);
  }

  private void writeBigDecimal(int scale, final int precision, BigInteger unscaledValue) throws IOException {
    final boolean signedValue = (unscaledValue.signum() < 0);
    if (signedValue) unscaledValue = unscaledValue.negate();

    final boolean signedScale = (scale < 0);
    if (signedScale) scale = -scale;

    final byte[] vData = unscaledValue.toByteArray();
    final int vDataBytes = (vData.length == 0) ? 1 : ((32 - Integer.numberOfLeadingZeros(vData.length)) + 7) >> 3;
    final int scaleBytes = (scale == 0) ? 1 : ((32 - Integer.numberOfLeadingZeros(scale)) + 7) >> 3;
    final int precisionBytes = (precision == 0) ? 1 : ((32 - Integer.numberOfLeadingZeros(precision)) + 7) >> 3;

    final byte[] buf = rawBuffer();
    int bufOff = rawBufferOffset(2 + scaleBytes + precisionBytes + vDataBytes);

    buf[bufOff++] = 0b00000_111;
    buf[bufOff++] = (byte) ((signedScale ? 0x80 : 0)
                  | ((scaleBytes - 1) << 5)
                  | ((precisionBytes - 1) << 3)
                  | (signedValue ? 4 : 0)
                  | (vDataBytes - 1));

    writeFixed(buf, bufOff, scale, scaleBytes);         bufOff += scaleBytes;
    writeFixed(buf, bufOff, precision, precisionBytes); bufOff += precisionBytes;
    writeFixed(buf, bufOff, vData.length, vDataBytes);
    write(vData, 0, vData.length);
  }

  // ====================================================================================================
  //  Int related
  // ====================================================================================================
  public final void writeInt(final long v) throws IOException {
    if (v >= -23 && v <= 24) {
      writeSmallInt((int)v);
    } else if (v > 0) {
      writeExternalInt(0b010_00000, v - 25);
    } else {
      writeExternalInt(0b011_00000, (-v) - 24);
    }
  }

  private void writeSmallInt(final int v) throws IOException {
    write((v > 0) ? (0b010_00000 | (v - 1)) : (0b011_00000 | (-v)));
  }

  private void writeExternalInt(final int head, final long v) throws IOException {
    final int w = (v != 0) ? ((64 - Long.numberOfLeadingZeros(v)) + 7) >> 3 : 1;
    final byte[] buf = rawBuffer();
    final int bufOff = rawBufferOffset(1 + w);
    buf[bufOff] = (byte) (head | (23 + w));
    writeFixed(buf, bufOff + 1, v, w);
  }

  private static int writeRawInt(final byte[] buf, final int off, long v) {
    final long inlineValue;
    final int head;
    if (v > 0) {
      inlineValue = v - 1;
      v = v - 25;
      head = 0b010_00000;
    } else {
      inlineValue = -v;
      v = (-v) - 24;
      head = 0b011_00000;
    }

    if (inlineValue < 24) {
      buf[off] = (byte) (head | inlineValue);
      return 1;
    }

    final int w = (v != 0) ? ((64 - Long.numberOfLeadingZeros(v)) + 7) >> 3 : 1;
    buf[off] = (byte) (head | (23 + w));
    writeFixed(buf, off + 1, v, w);
    return 1 + w;
  }

  // ====================================================================================================
  //  Items/Length related
  // ====================================================================================================
  private void writeLength(final int head, final int inlineMax, final int length) throws IOException {
    if (length <= inlineMax) {
      write(head | length);
      return;
    }

    final int deltaLength = length - inlineMax;
    final int bytes = ((32 - Integer.numberOfLeadingZeros(deltaLength)) + 7) >> 3;
    final byte[] buf = rawBuffer();
    final int bufOff = rawBufferOffset(1 + bytes);
    buf[bufOff] = (byte) (head | (inlineMax + bytes));
    writeFixed(buf, bufOff + 1, deltaLength, bytes);
  }

  // ====================================================================================================
  //  Bytes related
  // ====================================================================================================
  public final void writeBytes(final byte[] buf, final int off, final int len) throws IOException {
    writeLength(0b10_000000, 59, len);
    write(buf, off, len);
  }

  // ====================================================================================================
  //  String related
  // ====================================================================================================
  public final void writeEmptyString() throws IOException {
    write(0b11_000000);
  }

  public final void writeString(final String text) throws IOException {
    final byte[] utf8 = text.getBytes(StandardCharsets.UTF_8);
    writeUtf8(utf8, 0, utf8.length);
  }

  public final void writeUtf8(final byte[] utf8, final int off, final int len) throws IOException {
    writeLength(0b11_000000, 59, len);
    write(utf8, off, len);
  }

  private YajbeEnumMapping enumMapping;
  public final void writeStringOrEnum(final YajbeEnumMappingConfig enumConfig, final String text) throws IOException {
    if (enumMapping == null) newEnumMapping(enumConfig);

    final int index = enumMapping.add(text);
    if (index < 0) {
      writeString(text);
      return;
    }

    //System.out.println("WRITE ENUM " + index + " -> " + text);
    final byte[] buf = rawBuffer();
    if (index <= 0xff) {
      final int bufOff = rawBufferOffset(2);
      buf[bufOff] = (byte)0b00001001;
      buf[bufOff + 1] = (byte)index;
    } else if (index <= 0xffff) {
      final int bufOff = rawBufferOffset(3);
      buf[bufOff] = (byte)0b00001010;
      writeFixed(buf, bufOff + 1, index, 2);
    } else {
      throw new IllegalArgumentException("enum index too large " + index);
    }
  }

  private void newEnumMapping(final YajbeEnumMappingConfig config) throws IOException {
    this.enumMapping = YajbeEnumMapping.fromConfig(config);

    final byte[] buf = rawBuffer();
    final int bufOff = rawBufferOffset(3);
    buf[bufOff] = (byte) 0b00001000;
    if (config instanceof final YajbeEnumLruMappingConfig lruConfig) {
      buf[bufOff + 1] = (byte) (26 - Integer.numberOfLeadingZeros(lruConfig.lruSize()));
      buf[bufOff + 2] = (byte) (lruConfig.minFreq() - 1);
      return;
    }
  }

  // ====================================================================================================
  //  Array related
  // ====================================================================================================
  public final boolean newArray() throws IOException {
    write(0b0010_1111);
    return true;
  }

  public final boolean newArray(final int size) throws IOException {
    writeLength(0b0010_0000, 10, size);
    return false;
  }

  public final void writeArray(final int[] array, final int offset, final int length) throws IOException {
    writeLength(0b0010_0000, 10, length);

    final byte[] buf = rawBuffer();
    int bufOff = rawBufferOffset();
    int itemIndex = 0;

    while (itemIndex < length) {
      int bufAvail = buf.length - bufOff;
      while (bufAvail >= 5 && itemIndex < length) {
        final int n = writeRawInt(buf, bufOff, array[offset + itemIndex++]);
        bufOff += n;
        bufAvail -= n;
      }

      rawBufferFlush(bufOff, 5);
      bufOff = 0;
    }
  }

  public final void writeArray(final long[] array, final int offset, final int length) throws IOException {
    writeLength(0b0010_0000, 10, length);

    final byte[] buf = rawBuffer();
    int bufOff = rawBufferOffset();
    int itemIndex = 0;

    while (itemIndex < length) {
      int bufAvail = buf.length - bufOff;
      while (bufAvail >= 9 && itemIndex < length) {
        final int n = writeRawInt(buf, bufOff, array[offset + itemIndex++]);
        bufOff += n;
        bufAvail -= n;
      }

      rawBufferFlush(bufOff, 9);
      bufOff = 0;
    }
  }

  // ====================================================================================================
  //  Object related
  // ====================================================================================================
  public final boolean newObject() throws IOException {
    write(0b0011_1111);
    return true;
  }

  public final boolean newObject(final int size) throws IOException {
    writeLength(0b0011_0000, 10, size);
    return false;
  }
}

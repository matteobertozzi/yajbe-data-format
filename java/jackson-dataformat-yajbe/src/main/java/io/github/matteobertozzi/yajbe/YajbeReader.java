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
import java.io.InputStream;
import java.math.BigDecimal;
import java.math.BigInteger;
import java.math.MathContext;
import java.nio.charset.Charset;

import com.fasterxml.jackson.core.JsonParser.NumberType;

abstract class YajbeReader {
  protected abstract int peek() throws IOException;
  protected abstract int read() throws IOException;
  protected abstract String readString(final int n) throws IOException;
  protected abstract ByteArraySlice readNBytes(final int n) throws IOException;
  protected abstract void readNBytes(final byte[] buf, final int off, final int len) throws IOException;
  protected abstract long readFixed(final int width) throws IOException;
  protected abstract int readFixedInt(final int width) throws IOException;

  // =========================================================================================================
  @SuppressWarnings("fallthrough")
  public static long readFixed(final byte[] buf, final int off, final int width) {
    long result = 0;
    switch (width) {
      case 8: result |= (buf[off + 7] & 0xFFL) << 56;
      case 7: result |= (buf[off + 6] & 0xFFL) << 48;
      case 6: result |= (buf[off + 5] & 0xFFL) << 40;
      case 5: result |= (buf[off + 4] & 0xFFL) << 32;
      case 4: result |= (buf[off + 3] & 0xFFL) << 24;
      case 3: result |= (buf[off + 2] & 0xFFL) << 16;
      case 2: result |= (buf[off + 1] & 0xFFL) << 8;
      case 1: result |= buf[off] & 0xFFL;
    }
    return result;
  }

  @SuppressWarnings("fallthrough")
  public static int readFixedInt(final byte[] buf, final int off, final int width) {
    long result = 0;
    switch (width) {
      case 4: result |= (buf[off + 3] & 0xFFL) << 24;
      case 3: result |= (buf[off + 2] & 0xFFL) << 16;
      case 2: result |= (buf[off + 1] & 0xFFL) << 8;
      case 1: result |= buf[off] & 0xFFL;
    }
    return (int)result;
  }
  // =========================================================================================================
  public static YajbeReader fromBytes(final byte[] buf) {
    return fromBytes(buf, 0, buf.length);
  }

  public static YajbeReader fromBytes(final byte[] buf, final int off, final int len) {
    return new YajbeReaderByteArray(buf, off, len);
  }

  public static YajbeReader fromStream(final InputStream in) {
    return new YajbeReaderStream(in);
  }

  // =========================================================================================================
  private NumberType numberType;
  private int intValue;
  private long longValue;
  private float floatValue;
  private double doubleValue;
  private BigInteger bigInteger;
  private BigDecimal bigDecimal;
  private String strValue;
  private ByteArraySlice bytesValue;

  public NumberType numberType() { return numberType; }
  public int intValue() { return intValue; }
  public long longValue() { return longValue; }
  public float floatValue() { return floatValue; }
  public double doubleValue() { return doubleValue; }
  public BigInteger bigInteger() { return bigInteger; }
  public BigDecimal bigDecimal() { return bigDecimal; }
  public ByteArraySlice bytesValue() { return bytesValue; }
  public String stringValue() { return strValue; }

  // ====================================================================================================
  //  String related
  // ====================================================================================================
  public final void decodeSmallString(final int head) throws IOException {
    strValue = readString(head & 0b111111);
    if (enumMapping != null) enumMapping.add(strValue);
  }

  public final void decodeString(final int head) throws IOException {
    final int length = 59 + readFixedInt((head & 0b111111) - 59);
    strValue = readString(length);
    if (enumMapping != null) enumMapping.add(strValue);
  }

  // ====================================================================================================
  //  Enum/String related
  // ====================================================================================================
  private YajbeEnumMapping enumMapping;

  public final void decodeEnumConfig(final int head) throws IOException {
    final int h1 = read();
    switch ((h1 >>> 4) & 0b1111) {
      case 0: // LRU
        final int freq = read();
        enumMapping = new YajbeEnumLruMapping(1 << (5 + (h1 & 0b1111)), 1 + freq);
        break;
    }
  }

  public final void decodeEnumString(final int head) throws IOException {
    switch (head) {
      case 0b00001001: {
        final int index = read();
        strValue = enumMapping.get(index);
        return;
      }
      case 0b00001010: {
        final int index = readFixedInt(2);
        strValue = enumMapping.get(index);
        return;
      }
    }
  }

  // ====================================================================================================
  //  Bytes related
  // ====================================================================================================
  public final void decodeSmallBytes(final int head) throws IOException {
    bytesValue = readNBytes(head & 0b111111);
    strValue = null;
  }

  public final void decodeBytes(final int head) throws IOException {
    final int length = 59 + readFixedInt((head & 0b111111) - 59);
    bytesValue = readNBytes(length);
    strValue = null;
  }

  // ====================================================================================================
  //  Int related
  // ====================================================================================================
  public final void decodeSmallInt(final int head) {
    final boolean signed = (head & 0b011_00000) == 0b011_00000;
    final int w = head & 0b11111;
    intValue = signed ? -w : (1 + w);
    numberType = NumberType.INT;
  }

  public final void decodeIntPositive(final int head) throws IOException {
    final int w = head & 0b11111;
    final long v = 25L + readFixed(w - 23);
    if (v <= Integer.MAX_VALUE) {
      intValue = (int)v;
      numberType = NumberType.INT;
    } else {
      longValue = v;
      numberType = NumberType.LONG;
    }
  }

  public final void decodeIntNegative(final int head) throws IOException {
    final int w = head & 0b11111;
    final long v = readFixed(w - 23);
    final long signedValue = -(v + 24L);
    if (signedValue >= Integer.MIN_VALUE) {
      intValue = (int)signedValue;
      numberType = NumberType.INT;
    } else {
      longValue = signedValue;
      numberType = NumberType.LONG;
    }
  }

  // ====================================================================================================
  //  Float related
  // ====================================================================================================
  public final void decodeFloatVle() {
    throw new UnsupportedOperationException("Not implemented decode float16/vle-float");
  }

  public final void decodeFloat32() throws IOException {
    final int i32 = readFixedInt(4);
    this.floatValue = Float.intBitsToFloat(i32);
    this.numberType = NumberType.FLOAT;
  }

  public final void decodeFloat64() throws IOException {
    final long i64 = readFixed(8);
    this.doubleValue = Double.longBitsToDouble(i64);
    this.numberType = NumberType.DOUBLE;
  }

  public final void decodeBigDecimal() throws IOException {
    final int head = read();
    final boolean signedScale = (head & 0x80) == 0x80;
    final int scaleBytes = 1 + ((head >> 5) & 3);
    final int precisionBytes = 1 + ((head >> 3) & 3);
    final boolean signedValue = (head & 4) == 4;
    final int vDataBytes = 1 + (head & 3);

    int scale = readFixedInt(scaleBytes);
    final int precision = readFixedInt(precisionBytes);
    final int vDataLength = readFixedInt(vDataBytes);
    final ByteArraySlice data = readNBytes(vDataLength);

    BigInteger unscaled = new BigInteger(data.buf(), data.off(), data.len());
    if (signedValue) unscaled = unscaled.negate();
    if (scale == 0 && precision == 0) {
      this.numberType = NumberType.BIG_INTEGER;
      this.bigInteger = unscaled;
      return;
    }

    if (signedScale) scale = -scale;
    this.numberType = NumberType.BIG_DECIMAL;
    this.bigDecimal = new BigDecimal(unscaled, scale, new MathContext(precision));
  }

  // ====================================================================================================
  //  Array/Object length related
  // ====================================================================================================
  public final int readItemCount(final int head) throws IOException {
    final int w = head & 0b1111;
    if (w <= 10) return w;
    return 10 + readFixedInt(w - 10);
  }

  // ====================================================================================================
  //  Utils
  // ====================================================================================================
  record ByteArraySlice (byte[] buf, int off, int len) {
    public ByteArraySlice(final byte[] buf) {
      this(buf, 0, buf.length);
    }

    public byte[] toByteArray() {
      final byte[] copy = new byte[len];
      System.arraycopy(buf, off, copy, 0, len);
      return copy;
    }

    public String toString(final Charset charsets) {
      return new String(buf, off, len, charsets);
    }
  }
}

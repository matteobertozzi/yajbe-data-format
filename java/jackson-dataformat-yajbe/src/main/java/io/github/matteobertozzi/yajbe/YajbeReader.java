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

public abstract class YajbeReader {
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

  public void decodeString(final int head) throws IOException {
    final int w = head & 0b111111;
    if (w < 60) {
      strValue = readString(w);
    } else {
      final int length = readFixedInt(w - 59);
      strValue = readString(length);
    }
  }

  public void decodeBytes(final int head) throws IOException {
    final int w = head & 0b111111;
    if (w < 60) {
      bytesValue = readNBytes(w);
    } else {
      final int length = readFixedInt(w - 59);
      bytesValue = readNBytes(length);
    }
  }

  public void decodeInt(final int head) throws IOException {
    final boolean signed = (head & 0b011_00000) == 0b011_00000;

    final int w = head & 0b11111;
    if (w < 24) {
      numberType = NumberType.INT;
      intValue = signed ? -w : (1 + w);
      return;
    }

    final int width = w - 23;
    final long v = readFixed(width);
    if (v <= Integer.MAX_VALUE) {
      numberType = NumberType.INT;
      intValue = (int)(signed ? -v : v);
    } else {
      numberType = NumberType.LONG;
      longValue = signed ? -v : v;
    }
  }

  private static String printArray(final byte[] buf) {
    final StringBuilder builder = new StringBuilder();
    builder.append("[");
    for (int i = 0; i < buf.length; ++i) {
      if (i > 0) builder.append(", ");
      builder.append(buf[i] & 0xff);
    }
    builder.append("]");
    return builder.toString();
  }

  public void decodeFloat(final int head) throws IOException {
    switch (head & 0b11) {
      case 0b00: {
        throw new UnsupportedOperationException("Not implemented decode float16/vle-float");
      }
      case 0b01: {
        final int i32 = readFixedInt(4);
        this.floatValue = Float.intBitsToFloat(i32);
        this.numberType = NumberType.FLOAT;
        return;
      }
      case 0b10: {
        final long i64 = readFixed(8);
        this.doubleValue = Double.longBitsToDouble(i64);
        this.numberType = NumberType.DOUBLE;
        return;
      }
      case 0b11: {
        decodeBigDecimal();
        return;
      }
    }
  }

  private void decodeBigDecimal() throws IOException {
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

  public int readItemCount(final int head) throws IOException {
    final int w = head & 0b1111;
    if (w <= 10) return w;
    if (w == 0b1111) return Integer.MAX_VALUE;
    return readFixedInt(w - 10);
  }

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

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

package tech.dnaco.yajbe;

import static org.junit.jupiter.api.Assertions.assertArrayEquals;
import static org.junit.jupiter.api.Assertions.assertEquals;

import java.io.IOException;
import java.math.BigDecimal;
import java.math.BigInteger;

import org.junit.jupiter.api.Test;

public class TestYajbeFloats extends BaseYajbeTest {
  @Test
  public void testSimple() throws IOException {
    assertEncodeDecode(0.0f, "0500000000");
    assertEncodeDecode(1.0f, "050000803f");
    assertEncodeDecode(1.1f, "05cdcc8c3f");
    assertEncodeDecode(-32.26664f, "050a1101c2");
    assertEncodeDecode(Float.POSITIVE_INFINITY, "050000807f");
    assertEncodeDecode(Float.NaN, "050000c07f");
    assertEncodeDecode(Float.NEGATIVE_INFINITY, "05000080ff");

    assertEncodeDecode(-0.0, "060000000000000080");
    assertEncodeDecode(-4.0, "0600000000000010c0");
    assertEncodeDecode(-4.1, "0666666666666610c0");
    assertEncodeDecode(1.5, "06000000000000f83f");
    assertEncodeDecode(65504.0, "060000000000fcef40");
    assertEncodeDecode(100000.0, "0600000000006af840");
    assertEncodeDecode(5.960464477539063e-8, "06000000000000703e");
    assertEncodeDecode(0.00006103515625, "06000000000000103f");
    assertEncodeDecode(-5.960464477539063e-8, "0600000000000070be");
    assertEncodeDecode(3.4028234663852886e+38, "06000000e0ffffef47");
    assertEncodeDecode(9007199254740994.0, "060100000000004043");
    assertEncodeDecode(-9007199254740994.0, "0601000000000040c3");
    assertEncodeDecode(1.0e+300, "069c7500883ce4377e");
    assertEncodeDecode(-40.049149, "06c8d0b1834a0644c0");
    assertEncodeDecode(Double.POSITIVE_INFINITY, "06000000000000f07f");
    assertEncodeDecode(Double.NaN, "06000000000000f87f");
    assertEncodeDecode(Double.NEGATIVE_INFINITY, "06000000000000f0ff");

    assertEncodeDecode(new BigInteger("-340282366920938463463374607431768211455"), "070400001100ffffffffffffffffffffffffffffffff");
    assertEncodeDecode(new BigInteger("340282366920938463463374607431768211455"),  "070000001100ffffffffffffffffffffffffffffffff");
    assertEncodeDecode(new BigDecimal(new BigInteger("-1234567"), -12), "07840c070312d687");
    assertEncodeDecode(new BigDecimal(new BigInteger("-1234567"), 12), "07040c070312d687");
    assertEncodeDecode(new BigDecimal(new BigInteger("1234567"), -12), "07800c070312d687");
    assertEncodeDecode(new BigDecimal(new BigInteger("1234567"), 12), "07000c070312d687");
  }

  @Test
  public void testRandFloatEncodeDecode() throws IOException {
    for (int i = 0; i < 100; ++i) {
      final float input = RANDOM.nextFloat();
      final byte[] enc = YAJBE_MAPPER.writeValueAsBytes(input);
      assertEquals(input, YAJBE_MAPPER.readValue(enc, float.class), 0.00000000001f);
    }
  }

  @Test
  public void testRandDoubleEncodeDecode() throws IOException {
    for (int i = 0; i < 100; ++i) {
      final double input = RANDOM.nextDouble();
      final byte[] enc = YAJBE_MAPPER.writeValueAsBytes(input);
      assertEquals(input, YAJBE_MAPPER.readValue(enc, double.class), 0.00000000001);
    }
  }

  @Test
  public void testRandFloatArrayEncodeDecode() throws IOException {
    for (int k = 0; k < 32; ++k) {
      final int length = RANDOM.nextInt(1 << 20);
      final float[] input = new float[length];
      for (int i = 0; i < length; ++i) {
        input[i] = RANDOM.nextFloat();
      }
      final byte[] enc = YAJBE_MAPPER.writeValueAsBytes(input);
      assertArrayEquals(input, YAJBE_MAPPER.readValue(enc, float[].class), 0.00000001f);
    }
  }

  @Test
  public void testRandDoubleArrayEncodeDecode() throws IOException {
    for (int k = 0; k < 32; ++k) {
      final int length = RANDOM.nextInt(1 << 20);
      final double[] input = new double[length];
      for (int i = 0; i < length; ++i) {
        input[i] = RANDOM.nextDouble();
      }
      final byte[] enc = YAJBE_MAPPER.writeValueAsBytes(input);
      assertArrayEquals(input, YAJBE_MAPPER.readValue(enc, double[].class), 0.00000001);
    }
  }

  @Test
  public void testRandBigIntegerEncodeDecode() throws IOException {
    for (int i = 0; i < 100; ++i) {
      final BigInteger input = new BigInteger(RANDOM.nextInt(256), RANDOM);
      final byte[] enc = YAJBE_MAPPER.writeValueAsBytes(input);
      assertEquals(input, YAJBE_MAPPER.readValue(enc, BigInteger.class));
    }
  }

  @Test
  public void testRandBigDecimalEncodeDecode() throws IOException {
    for (int i = 0; i < 100; ++i) {
      final BigDecimal input = new BigDecimal(new BigInteger(RANDOM.nextInt(512), RANDOM), RANDOM.nextInt());
      final byte[] enc = YAJBE_MAPPER.writeValueAsBytes(input);
      assertEquals(input, YAJBE_MAPPER.readValue(enc, BigDecimal.class));
    }
  }

  private void assertEncodeDecode(final float input, final String expectedEnc) throws IOException {
    final byte[] enc = YAJBE_MAPPER.writeValueAsBytes(input);
    assertHexEquals(expectedEnc, enc);
    assertEquals(input, YAJBE_MAPPER.readValue(enc, float.class), 0.00000000001f);
  }

  private void assertEncodeDecode(final double input, final String expectedEnc) throws IOException {
    final byte[] enc = YAJBE_MAPPER.writeValueAsBytes(input);
    assertHexEquals(expectedEnc, enc);
    assertEquals(input, YAJBE_MAPPER.readValue(enc, double.class), 0.00000000001);
  }

  private void assertEncodeDecode(final BigInteger input, final String expectedEnc) throws IOException {
    final byte[] enc = YAJBE_MAPPER.writeValueAsBytes(input);
    assertHexEquals(expectedEnc, enc);
    assertEquals(input, YAJBE_MAPPER.readValue(enc, BigInteger.class));
  }

  private void assertEncodeDecode(final BigDecimal input, final String expectedEnc) throws IOException {
    final byte[] enc = YAJBE_MAPPER.writeValueAsBytes(input);
    assertHexEquals(expectedEnc, enc);
    assertEquals(input, YAJBE_MAPPER.readValue(enc, BigDecimal.class));
  }
}

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

import static org.junit.jupiter.api.Assertions.assertArrayEquals;
import static org.junit.jupiter.api.Assertions.assertEquals;

import java.io.IOException;
import java.util.HexFormat;
import java.util.Random;

import com.fasterxml.jackson.annotation.JsonAutoDetect.Visibility;
import com.fasterxml.jackson.annotation.JsonInclude;
import com.fasterxml.jackson.annotation.PropertyAccessor;
import com.fasterxml.jackson.databind.DeserializationFeature;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.databind.SerializationFeature;
import com.fasterxml.jackson.databind.json.JsonMapper;

public abstract class BaseYajbeTest {
  public static long RANDOM_SEED = Long.parseLong(System.getProperty("yajbe.test.random.seed", String.valueOf(Math.round(Math.random() * Long.MAX_VALUE))));

  public final Random RANDOM;
  public final ObjectMapper YAJBE_MAPPER = newObjectMapper(new YajbeMapper());
  public final ObjectMapper JSON_MAPPER = newObjectMapper(new JsonMapper());

  protected BaseYajbeTest() {
    System.out.println("Running test " + getClass().getName() + " with seed " + RANDOM_SEED);
    this.RANDOM = new Random(RANDOM_SEED);
  }

  public static String toBinaryString(final byte[] buf) {
    return toBinaryString(buf, 0, buf.length);
  }

  public static String toBinaryString(final byte[] buf, final int off, final int len) {
    final StringBuilder builder = new StringBuilder(len * 9);
    for (int i = 0; i < len; ++i) {
      if (i != 0) builder.append(", ");

      final String bin = Integer.toBinaryString(buf[off + i] & 0xff);
      if (bin.length() != 8) builder.append("0".repeat(8 - bin.length()));
      builder.append(bin);
    }
    return builder.toString();
  }

  public static void assertHexEquals(final String expected, final byte[] actual) {
    assertEquals(expected, HexFormat.of().formatHex(actual));
  }

  public <T> void assertEncodeDecode(final T input, final Class<T> classOfT, final String expectedEnc) throws IOException {
    final byte[] enc = YAJBE_MAPPER.writeValueAsBytes(input);
    assertHexEquals(expectedEnc, enc);
    assertEquals(input, YAJBE_MAPPER.readValue(enc, classOfT));
  }

  public <T> void assertArrayEncodeDecode(final T[] input, final Class<T[]> classOfT, final String expectedEnc) throws IOException {
    final byte[] enc = YAJBE_MAPPER.writeValueAsBytes(input);
    assertHexEquals(expectedEnc, enc);
    assertArrayEquals(input, YAJBE_MAPPER.readValue(enc, classOfT));
  }

  private static ObjectMapper newObjectMapper(final ObjectMapper mapper) {
    mapper.setVisibility(PropertyAccessor.FIELD, Visibility.ANY);
    mapper.setVisibility(PropertyAccessor.GETTER, Visibility.NONE);
    mapper.setVisibility(PropertyAccessor.IS_GETTER, Visibility.NONE);

    // --- Deserialization ---
    // Just ignore unknown fields, don't stop parsing
    mapper.configure(DeserializationFeature.FAIL_ON_UNKNOWN_PROPERTIES, false);
    // Trying to deserialize value into an enum, don't fail on unknown value, use null instead
    mapper.configure(DeserializationFeature.READ_UNKNOWN_ENUM_VALUES_AS_NULL, true);

    // --- Serialization ---
    // Don't include properties with null value in JSON output
    mapper.setSerializationInclusion(JsonInclude.Include.ALWAYS);
    // Use default pretty printer
    mapper.configure(SerializationFeature.INDENT_OUTPUT, false);
    mapper.configure(SerializationFeature.FAIL_ON_EMPTY_BEANS, false);

    //mapper.setAnnotationIntrospector(new ExtentedAnnotationIntrospector());
    return mapper;
  }

  // ===============================================================================================
  /**
   * Random int[] with balanced items between 1 and 4 bytes
   */
  public int[] randIntBlock(final int length) {
    final int[] block = new int[length];
    for (int i = 0; i < length; ++i) {
      final int w = 1 + RANDOM.nextInt(0, 4);
      final int v = RANDOM.nextInt(0, Math.toIntExact(Math.round(Math.pow(2, (w << 3) - 1) - 1)));
      block[i] = RANDOM.nextBoolean() ? v : -v;
    }
    return block;
  }

  /**
   * Random long[] with balanced items between 1 and 8 bytes
   */
  public long[] randLongBlock(final int length) {
    final long[] block = new long[length];
    for (int i = 0; i < length; ++i) {
      final int w = 1 + RANDOM.nextInt(0, 8);
      final long v = RANDOM.nextLong(0, Math.round(Math.pow(2, (w << 3) - 1) - 1));
      block[i] = RANDOM.nextBoolean() ? v : -v;
    }
    return block;
  }

  private static final char[] TEXT_CHARS = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ".toCharArray();
  public String randText(final int length) {
    final StringBuilder sw = new StringBuilder(length);
    for (int i = 0; i < length; ++i) {
      final int wordLength = 4 + RANDOM.nextInt(8);
      for (int w = 0; w < wordLength; ++w) {
        sw.append(TEXT_CHARS[RANDOM.nextInt(TEXT_CHARS.length)]);
      }
      sw.append(' ');
    }
    return sw.toString();
  }

  private static final char[] FIELD_NAME_CHARS = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ_-".toCharArray();
  public String generateFieldName(final int minLength, final int maxLength) {
    final int length = minLength + RANDOM.nextInt(maxLength - minLength);
    final StringBuilder builder = new StringBuilder(length);
    for (int i = 0; i < length; ++i) {
      builder.append(FIELD_NAME_CHARS[RANDOM.nextInt(FIELD_NAME_CHARS.length)]);
    }
    return builder.toString();
  }
}

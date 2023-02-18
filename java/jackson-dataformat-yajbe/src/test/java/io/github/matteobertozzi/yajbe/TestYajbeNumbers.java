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

import static org.junit.jupiter.api.Assertions.assertEquals;

import java.io.IOException;
import java.math.BigDecimal;
import java.math.BigInteger;
import java.util.LinkedHashMap;
import java.util.Map;

import org.junit.jupiter.api.Test;

public class TestYajbeNumbers extends BaseYajbeTest {
  record ObjectWithNumbers (int intValue, long longValue, float floatValue, double doubleValue, BigInteger bigInt, BigDecimal bigDecimal) {}

  @Test
  public void testSimple() throws IOException {
    testObjectNumbers(0, 0L, 0.0f, 0.0d, BigInteger.ZERO, BigDecimal.ZERO,
      "3f88696e7456616c756560e400056c6f6e6760e50005666c6f61740500000000e60005646f75626c6506000000000000000086626967496e74070000000100c703446563696d616c07000001010001");
    testObjectNumbers(1, 1L, 1.0f, 1.0d, BigInteger.ONE, BigDecimal.ONE,
      "3f88696e7456616c756540e400056c6f6e6740e50005666c6f6174050000803fe60005646f75626c6506000000000000f03f86626967496e74070000000101c703446563696d616c07000001010101");
    testObjectNumbers(123, 12345L, 123.45f, 12345.6789d, new BigInteger("1180591620717411303423"), new BigDecimal("1180591620717411303423.12345"),
      "3f88696e7456616c7565587be400056c6f6e67593930e50005666c6f61740566e6f642e60005646f75626c6506a1f831e6d61cc84086626967496e7407000000093fffffffffffffffffc703446563696d616c0700051b0b61a7fffffffffffffea99901");
    testObjectNumbers(-123, -12345L, -123.45f, -12345.6789d, new BigInteger("-1180591620717411303423"), new BigDecimal("-1180591620717411303423.12345"),
      "3f88696e7456616c7565787be400056c6f6e67793930e50005666c6f61740566e6f6c2e60005646f75626c6506a1f831e6d61cc8c086626967496e7407040000093fffffffffffffffffc703446563696d616c0704051b0b61a7fffffffffffffea99901");
  }

  @Test
  public void testRandom() throws IOException {
    for (int k = 0; k < 32; ++k) {
      testObjectNumbers(RANDOM.nextInt(), RANDOM.nextLong(),
        RANDOM.nextFloat(), RANDOM.nextDouble(),
        new BigInteger(RANDOM.nextInt(256), RANDOM),
        new BigDecimal(new BigInteger(RANDOM.nextInt(256), RANDOM), RANDOM.nextInt(0xff)),
        null);
    }
  }

  private void testObjectNumbers(final int intValue, final long longValue,
      final float floatValue, final double doubleValue,
      final BigInteger bigInt, final BigDecimal bigDecimal,
      final String encHex) throws IOException {
    final Map<String, Object> map = new LinkedHashMap<>();
    map.put("intValue", intValue);
    map.put("longValue", longValue);
    map.put("floatValue", floatValue);
    map.put("doubleValue", doubleValue);
    map.put("bigInt", bigInt);
    map.put("bigDecimal", bigDecimal);
    final byte[] encMap = YAJBE_MAPPER.writeValueAsBytes(map);
    if (encHex != null) assertHexEquals(encHex, encMap);

    final Map<String, Object> result = YAJBE_MAPPER.readerFor(Map.class).readValue(encMap);
    assertEquals(Integer.class, result.get("intValue").getClass());
    assertEquals(intValue, result.get("intValue"));
    if (longValue >= Integer.MIN_VALUE && longValue <= Integer.MAX_VALUE) {
      assertEquals(Integer.class, result.get("longValue").getClass());
      assertEquals((int)longValue, result.get("longValue"));
    } else {
      assertEquals(Long.class, result.get("longValue").getClass());
      assertEquals(longValue, result.get("longValue"));
    }
    assertEquals(Float.class, result.get("floatValue").getClass());
    assertEquals(floatValue, (float)result.get("floatValue"), 0.00000001);
    assertEquals(Double.class, result.get("doubleValue").getClass());
    assertEquals(doubleValue, (double)result.get("doubleValue"), 0.00000001);
    assertEquals(BigInteger.class, result.get("bigInt").getClass());
    assertEquals(bigInt, result.get("bigInt"));
    assertEquals(BigDecimal.class, result.get("bigDecimal").getClass());
    assertEquals(bigDecimal, result.get("bigDecimal"));

    final ObjectWithNumbers obj = YAJBE_MAPPER.readValue(encMap, ObjectWithNumbers.class);
    final byte[] encObj = YAJBE_MAPPER.writeValueAsBytes(obj);
    if (encHex != null) assertHexEquals(encHex, encObj);
    assertEquals(intValue, obj.intValue());
    assertEquals(longValue, obj.longValue());
    assertEquals(floatValue, obj.floatValue(), 0.00000001);
    assertEquals(doubleValue, obj.doubleValue(), 0.00000001);
    assertEquals(bigInt, obj.bigInt());
    assertEquals(bigDecimal, obj.bigDecimal());
  }
}

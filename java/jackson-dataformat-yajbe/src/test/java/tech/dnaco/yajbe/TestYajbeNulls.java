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

import java.io.IOException;
import java.util.LinkedHashMap;
import java.util.List;

import org.junit.jupiter.api.Test;

import static org.junit.jupiter.api.Assertions.*;

public class TestYajbeNulls  extends BaseYajbeTest {
  record ObjectWithNulls (Object a, Object b, Object[] c) {}

  @Test
  public void testSimple() throws IOException {
    final byte[] NULL_BYTES = new byte[] { 0x00 };

    assertArrayEquals(NULL_BYTES, YAJBE_MAPPER.writeValueAsBytes(null));
    assertNull(YAJBE_MAPPER.readValue(NULL_BYTES, Object.class));
  }

  @Test
  public void testArray() throws IOException {
    final Object[] array = new Object[] { null, null, new Object[] { null, null }, null };
    final byte[] enc = YAJBE_MAPPER.writeValueAsBytes(array);
    assertHexEquals("24000022000000", enc);

    final Object[] r = YAJBE_MAPPER.readValue(enc, Object[].class);
    assertEquals(4, r.length);
    assertNull(r[0]);
    assertNull(r[1]);
    @SuppressWarnings("unchecked")
    final List<Object> innerArray = (List<Object>) r[2];
    assertEquals(2, innerArray.size());
    assertNull(innerArray.get(0));
    assertNull(innerArray.get(1));
    assertNull(r[3]);
  }

  @Test
  public void testRecord() throws IOException {
    final ObjectWithNulls obj = new ObjectWithNulls(null, null, new Object[] { null, null });
    final byte[] enc = YAJBE_MAPPER.writeValueAsBytes(obj);
    assertHexEquals("3f816100816200816322000001", enc);
    assertNull(obj.a);
    assertNull(obj.b);
    assertEquals(2, obj.c.length);
    assertNull(obj.c[0]);
    assertNull(obj.c[1]);
  }

  @Test
  public void testMap() throws IOException {
    final LinkedHashMap<String, Object> map = new LinkedHashMap<>();
    map.put("a", null);
    map.put("b", null);
    final byte[] enc = YAJBE_MAPPER.writeValueAsBytes(map);
    assertHexEquals("3f81610081620001", enc);
    assertNull(map.getOrDefault("a", 1));
    assertNull(map.getOrDefault("b", 2));
  }
}

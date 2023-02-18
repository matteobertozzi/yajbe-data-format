/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package tech.dnaco.yajbe.examples;

import java.util.HexFormat;
import java.util.Map;

import com.fasterxml.jackson.databind.json.JsonMapper;

import tech.dnaco.yajbe.YajbeMapper;

public class SimpleEncodeDecode {
  public record TestObj (int a, float b, String c) {}

  public static void main(final String[] args) throws Exception {
    final JsonMapper json = new JsonMapper();
    final YajbeMapper yajbe = new YajbeMapper(); // the YAJBE mapper to be used for encode/decode

    // encode/decode using json mapper
    final String j1 = json.writeValueAsString(Map.of("a", 10, "b", 20));
    System.out.println(j1); // { "a": 10, "b": 20 }
    System.out.println(json.readValue(j1, Map.class)); // {a=10, b=20}

    // encode/decode using yajbe mapper
    final byte[] y1 = yajbe.writeValueAsBytes(Map.of("a", 10, "b", 20));
    System.out.println(HexFormat.of().formatHex(y1)); // 3f81614981625301
    System.out.println(yajbe.readValue(y1, Map.class)); // {a=10, b=20}

    // encode decode a java record
    final byte[] y2 = yajbe.writeValueAsBytes(new TestObj(1, 5.23f, "test"));
    System.out.println(HexFormat.of().formatHex(y2)); // 3f816140816205295ca7408163c47465737401
    System.out.println(yajbe.readValue(y2, TestObj.class)); // TestObj[a=1, b=5.23, c=test]
  }
}

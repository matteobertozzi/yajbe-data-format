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

import org.junit.jupiter.api.Test;

import com.fasterxml.jackson.core.JsonProcessingException;

public class TestYajbeInts extends BaseYajbeTest {
  @Test
  public void testSimple() throws IOException {
    // positive ints
    assertEncodeDecode(1, "40");
    assertEncodeDecode(7, "46");
    assertEncodeDecode(24, "57");
    assertEncodeDecode(25, "5819");
    assertEncodeDecode(0xff, "58ff");
    assertEncodeDecode(0xffff, "59ffff");
    assertEncodeDecode(0xffffff, "5affffff");
    assertEncodeDecode(0xffffffffL, "5bffffffff");
    assertEncodeDecode(0xffffffffffL, "5cffffffffff");
    assertEncodeDecode(0xffffffffffffL, "5dffffffffffff");
    assertEncodeDecode(0x1fffffffffffffL, "5effffffffffff1f");
    assertEncodeDecode(0xffffffffffffffL, "5effffffffffffff");
    assertEncodeDecode(0xfffffffffffffffL, "5fffffffffffffff0f");
    assertEncodeDecode(0x7fffffffffffffffL, "5fffffffffffffff7f");

    assertEncodeDecode(100, "5864");
    assertEncodeDecode(1000, "59e803");
    assertEncodeDecode(1000000L, "5a40420f");
    assertEncodeDecode(1000000000000L, "5c0010a5d4e8");
    assertEncodeDecode(100000000000000L, "5d00407a10f35a");

    // negative ints
    assertEncodeDecode(0, "60");
    assertEncodeDecode(-1, "61");
    assertEncodeDecode(-7, "67");
    assertEncodeDecode(-23, "77");
    assertEncodeDecode(-24, "7818");
    assertEncodeDecode(-25, "7819");
    assertEncodeDecode(-0xff, "78ff");
    assertEncodeDecode(-0xffff, "79ffff");
    assertEncodeDecode(-0xffffff, "7affffff");
    assertEncodeDecode(-0xffffffffL, "7bffffffff");
    assertEncodeDecode(-0xffffffffffL, "7cffffffffff");
    assertEncodeDecode(-0xffffffffffffL, "7dffffffffffff");
    assertEncodeDecode(-0x1fffffffffffffL, "7effffffffffff1f");
    assertEncodeDecode(-0xffffffffffffffL, "7effffffffffffff");
    assertEncodeDecode(-0xfffffffffffffffL, "7fffffffffffffff0f");
    assertEncodeDecode(-0x7fffffffffffffffL, "7fffffffffffffff7f");

    assertEncodeDecode(-100, "7864");
    assertEncodeDecode(-1000, "79e803");
    assertEncodeDecode(-1000000L, "7a40420f");
    assertEncodeDecode(-1000000000000L, "7c0010a5d4e8");
    assertEncodeDecode(-100000000000000L, "7d00407a10f35a");
  }

  @Test
  public void testSmallInlineInt() throws JsonProcessingException {
    final String[] expected = new String[] {
      "790001", "78FF", "78FE", "78FD", "78FC", "78FB", "78FA", "78F9", "78F8", "78F7", "78F6", "78F5", "78F4",
      "78F3", "78F2", "78F1", "78F0", "78EF", "78EE", "78ED", "78EC", "78EB", "78EA", "78E9", "78E8", "78E7", "78E6",
      "78E5", "78E4", "78E3", "78E2", "78E1", "78E0", "78DF", "78DE", "78DD", "78DC", "78DB", "78DA", "78D9", "78D8",
      "78D7", "78D6", "78D5", "78D4", "78D3", "78D2", "78D1", "78D0", "78CF", "78CE", "78CD", "78CC", "78CB", "78CA",
      "78C9", "78C8", "78C7", "78C6", "78C5", "78C4", "78C3", "78C2", "78C1", "78C0", "78BF", "78BE", "78BD", "78BC",
      "78BB", "78BA", "78B9", "78B8", "78B7", "78B6", "78B5", "78B4", "78B3", "78B2", "78B1", "78B0", "78AF", "78AE",
      "78AD", "78AC", "78AB", "78AA", "78A9", "78A8", "78A7", "78A6", "78A5", "78A4", "78A3", "78A2", "78A1", "78A0",
      "789F", "789E", "789D", "789C", "789B", "789A", "7899", "7898", "7897", "7896", "7895", "7894", "7893", "7892",
      "7891", "7890", "788F", "788E", "788D", "788C", "788B", "788A", "7889", "7888", "7887", "7886", "7885", "7884",
      "7883", "7882", "7881", "7880", "787F", "787E", "787D", "787C", "787B", "787A", "7879", "7878", "7877", "7876",
      "7875", "7874", "7873", "7872", "7871", "7870", "786F", "786E", "786D", "786C", "786B", "786A", "7869", "7868",
      "7867", "7866", "7865", "7864", "7863", "7862", "7861", "7860", "785F", "785E", "785D", "785C", "785B", "785A",
      "7859", "7858", "7857", "7856", "7855", "7854", "7853", "7852", "7851", "7850", "784F", "784E", "784D", "784C",
      "784B", "784A", "7849", "7848", "7847", "7846", "7845", "7844", "7843", "7842", "7841", "7840", "783F", "783E",
      "783D", "783C", "783B", "783A", "7839", "7838", "7837", "7836", "7835", "7834", "7833", "7832", "7831", "7830",
      "782F", "782E", "782D", "782C", "782B", "782A", "7829", "7828", "7827", "7826", "7825", "7824", "7823", "7822",
      "7821", "7820", "781F", "781E", "781D", "781C", "781B", "781A", "7819", "7818", "77", "76", "75", "74", "73",
      "72", "71", "70", "6F", "6E", "6D", "6C", "6B", "6A", "69", "68", "67", "66", "65", "64", "63", "62", "61",
      "60", "40", "41", "42", "43", "44", "45", "46", "47", "48", "49", "4A", "4B", "4C", "4D", "4E", "4F", "50",
      "51", "52", "53", "54", "55", "56", "57", "5819", "581A", "581B", "581C", "581D", "581E", "581F", "5820",
      "5821", "5822", "5823", "5824", "5825", "5826", "5827", "5828", "5829", "582A", "582B", "582C", "582D", "582E",
      "582F", "5830", "5831", "5832", "5833", "5834", "5835", "5836", "5837", "5838", "5839", "583A", "583B", "583C",
      "583D", "583E", "583F", "5840", "5841", "5842", "5843", "5844", "5845", "5846", "5847", "5848", "5849", "584A",
      "584B", "584C", "584D", "584E", "584F", "5850", "5851", "5852", "5853", "5854", "5855", "5856", "5857", "5858",
      "5859", "585A", "585B", "585C", "585D", "585E", "585F", "5860", "5861", "5862", "5863", "5864", "5865", "5866",
      "5867", "5868", "5869", "586A", "586B", "586C", "586D", "586E", "586F", "5870", "5871", "5872", "5873", "5874",
      "5875", "5876", "5877", "5878", "5879", "587A", "587B", "587C", "587D", "587E", "587F", "5880", "5881", "5882",
      "5883", "5884", "5885", "5886", "5887", "5888", "5889", "588A", "588B", "588C", "588D", "588E", "588F", "5890",
      "5891", "5892", "5893", "5894", "5895", "5896", "5897", "5898", "5899", "589A", "589B", "589C", "589D", "589E",
      "589F", "58A0", "58A1", "58A2", "58A3", "58A4", "58A5", "58A6", "58A7", "58A8", "58A9", "58AA", "58AB", "58AC",
      "58AD", "58AE", "58AF", "58B0", "58B1", "58B2", "58B3", "58B4", "58B5", "58B6", "58B7", "58B8", "58B9", "58BA",
      "58BB", "58BC", "58BD", "58BE", "58BF", "58C0", "58C1", "58C2", "58C3", "58C4", "58C5", "58C6", "58C7", "58C8",
      "58C9", "58CA", "58CB", "58CC", "58CD", "58CE", "58CF", "58D0", "58D1", "58D2", "58D3", "58D4", "58D5", "58D6",
      "58D7", "58D8", "58D9", "58DA", "58DB", "58DC", "58DD", "58DE", "58DF", "58E0", "58E1", "58E2", "58E3", "58E4",
      "58E5", "58E6", "58E7", "58E8", "58E9", "58EA", "58EB", "58EC", "58ED", "58EE", "58EF", "58F0", "58F1", "58F2",
      "58F3", "58F4", "58F5", "58F6", "58F7", "58F8", "58F9", "58FA", "58FB", "58FC", "58FD", "58FE", "58FF"
    };

    int value = -256;
    for (final String hex : expected) {
      final byte[] r = YAJBE_MAPPER.writeValueAsBytes(value);
      assertArrayEquals(HexFormat.of().parseHex(hex), r);
      value++;
    }
  }

  @Test
  public void testRandIntEncodeDecode() throws IOException {
    for (int i = 0; i < 1000; ++i) {
      final int input = RANDOM.nextInt();
      final byte[] enc = YAJBE_MAPPER.writeValueAsBytes(input);
      assertEquals(input, YAJBE_MAPPER.readValue(enc, int.class));
    }
  }

  @Test
  public void testRandLongEncodeDecode() throws IOException {
    for (int i = 0; i < 1000; ++i) {
      final long input = RANDOM.nextLong();
      final byte[] enc = YAJBE_MAPPER.writeValueAsBytes(input);
      assertEquals(input, YAJBE_MAPPER.readValue(enc, long.class));
    }
  }

  @Test
  public void testRandIntArrayEncodeDecode() throws IOException {
    for (int k = 0; k < 32; ++k) {
      final int length = RANDOM.nextInt(1 << 20);
      final int[] input = new int[length];
      for (int i = 0; i < length; ++i) {
        input[i] = RANDOM.nextInt();
      }
      final byte[] enc = YAJBE_MAPPER.writeValueAsBytes(input);
      assertArrayEquals(input, YAJBE_MAPPER.readValue(enc, int[].class));
    }
  }

  @Test
  public void testRandLongArrayEncodeDecode() throws IOException {
    for (int k = 0; k < 32; ++k) {
      final int length = RANDOM.nextInt(1 << 20);
      final long[] input = new long[length];
      for (int i = 0; i < length; ++i) {
        input[i] = RANDOM.nextLong();
      }
      final byte[] enc = YAJBE_MAPPER.writeValueAsBytes(input);
      assertArrayEquals(input, YAJBE_MAPPER.readValue(enc, long[].class));
    }
  }

  private void assertEncodeDecode(final int input, final String expectedEnc) throws IOException {
    final byte[] enc = YAJBE_MAPPER.writeValueAsBytes(input);
    assertHexEquals(expectedEnc, enc);
    assertEquals(input, YAJBE_MAPPER.readValue(enc, int.class));
  }

  private void assertEncodeDecode(final long input, final String expectedEnc) throws IOException {
    final byte[] enc = YAJBE_MAPPER.writeValueAsBytes(input);
    assertHexEquals(expectedEnc, enc);
    assertEquals(input, YAJBE_MAPPER.readValue(enc, long.class));
  }
}

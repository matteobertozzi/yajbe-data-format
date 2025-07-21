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
    assertEncodeDecode(25, "5800");
    assertEncodeDecode(127, "5866");
    assertEncodeDecode(128, "5867");
    assertEncodeDecode(0xff, "58e6");
    assertEncodeDecode(0xffff, "59e6ff");
    assertEncodeDecode(0xffffff, "5ae6ffff");
    assertEncodeDecode(0xffffffffL, "5be6ffffff");
    assertEncodeDecode(0xffffffffffL, "5ce6ffffffff");
    assertEncodeDecode(0xffffffffffffL, "5de6ffffffffff");
    assertEncodeDecode(0x1fffffffffffffL, "5ee6ffffffffff1f");
    assertEncodeDecode(0xffffffffffffffL, "5ee6ffffffffffff");
    assertEncodeDecode(0xfffffffffffffffL, "5fe6ffffffffffff0f");
    assertEncodeDecode(0x7fffffffffffffffL, "5fe6ffffffffffff7f");

    assertEncodeDecode(100, "584b");
    assertEncodeDecode(1000, "59cf03");
    assertEncodeDecode(1000000L, "5a27420f");
    assertEncodeDecode(1000000000000L, "5ce70fa5d4e8");
    assertEncodeDecode(100000000000000L, "5de73f7a10f35a");

    // negative ints
    assertEncodeDecode(0, "60");
    assertEncodeDecode(-1, "61");
    assertEncodeDecode(-7, "67");
    assertEncodeDecode(-23, "77");
    assertEncodeDecode(-24, "7800");
    assertEncodeDecode(-25, "7801");
    assertEncodeDecode(-0xff, "78e7");
    assertEncodeDecode(-0xffff, "79e7ff");
    assertEncodeDecode(-0xffffff, "7ae7ffff");
    assertEncodeDecode(-0xffffffffL, "7be7ffffff");
    assertEncodeDecode(-0xffffffffffL, "7ce7ffffffff");
    assertEncodeDecode(-0xffffffffffffL, "7de7ffffffffff");
    assertEncodeDecode(-0x1fffffffffffffL, "7ee7ffffffffff1f");
    assertEncodeDecode(-0xffffffffffffffL, "7ee7ffffffffffff");
    assertEncodeDecode(-0xfffffffffffffffL, "7fe7ffffffffffff0f");
    assertEncodeDecode(-0x7fffffffffffffffL, "7fe7ffffffffffff7f");

    assertEncodeDecode(-100, "784c");
    assertEncodeDecode(-1000, "79d003");
    assertEncodeDecode(-1000000L, "7a28420f");
    assertEncodeDecode(-1000000000000L, "7ce80fa5d4e8");
    assertEncodeDecode(-100000000000000L, "7de83f7a10f35a");
  }

  @Test
  public void testSmallInlineInt() throws JsonProcessingException {
    final String[] expected = new String[] {
      "790001",
      "78ff", "78fe", "78fd", "78fc", "78fb", "78fa", "78f9", "78f8", "78f7", "78f6", "78f5", "78f4", "78f3", "78f2", "78f1", "78f0",
      "78ef", "78ee", "78ed", "78ec", "78eb", "78ea", "78e9", "78e8", "78e7", "78e6", "78e5", "78e4", "78e3", "78e2", "78e1", "78e0",
      "78df", "78de", "78dd", "78dc", "78db", "78da", "78d9", "78d8", "78d7", "78d6", "78d5", "78d4", "78d3", "78d2", "78d1", "78d0",
      "78cf", "78ce", "78cd", "78cc", "78cb", "78ca", "78c9", "78c8", "78c7", "78c6", "78c5", "78c4", "78c3", "78c2", "78c1", "78c0",
      "78bf", "78be", "78bd", "78bc", "78bb", "78ba", "78b9", "78b8", "78b7", "78b6", "78b5", "78b4", "78b3", "78b2", "78b1", "78b0",
      "78af", "78ae", "78ad", "78ac", "78ab", "78aa", "78a9", "78a8", "78a7", "78a6", "78a5", "78a4", "78a3", "78a2", "78a1", "78a0",
      "789f", "789e", "789d", "789c", "789b", "789a", "7899", "7898", "7897", "7896", "7895", "7894", "7893", "7892", "7891", "7890",
      "788f", "788e", "788d", "788c", "788b", "788a", "7889", "7888", "7887", "7886", "7885", "7884", "7883", "7882", "7881", "7880",
      "787f", "787e", "787d", "787c", "787b", "787a", "7879", "7878", "7877", "7876", "7875", "7874", "7873", "7872", "7871", "7870",
      "786f", "786e", "786d", "786c", "786b", "786a", "7869", "7868", "7867", "7866", "7865", "7864", "7863", "7862", "7861", "7860",
      "785f", "785e", "785d", "785c", "785b", "785a", "7859", "7858", "7857", "7856", "7855", "7854", "7853", "7852", "7851", "7850",
      "784f", "784e", "784d", "784c", "784b", "784a", "7849", "7848", "7847", "7846", "7845", "7844", "7843", "7842", "7841", "7840",
      "783f", "783e", "783d", "783c", "783b", "783a", "7839", "7838", "7837", "7836", "7835", "7834", "7833", "7832", "7831", "7830",
      "782f", "782e", "782d", "782c", "782b", "782a", "7829", "7828", "7827", "7826", "7825", "7824", "7823", "7822", "7821", "7820",
      "781f", "781e", "781d", "781c", "781b", "781a", "7819", "7818", "7817", "7816", "7815", "7814", "7813", "7812", "7811", "7810",
      "780f", "780e", "780d", "780c", "780b", "780a", "7809", "7808", "7807", "7806", "7805", "7804", "7803", "7802", "7801", "7800",
      "77", "76", "75", "74", "73", "72", "71", "70", "6f", "6e", "6d", "6c", "6b", "6a", "69", "68", "67", "66", "65", "64", "63", "62", "61", "60",
      "40", "41", "42", "43", "44", "45", "46", "47", "48", "49", "4a", "4b", "4c", "4d", "4e", "4f", "50", "51", "52", "53", "54", "55", "56", "57", "5800",
      "5801", "5802", "5803", "5804", "5805", "5806", "5807", "5808", "5809", "580a", "580b", "580c", "580d", "580e", "580f", "5810",
      "5811", "5812", "5813", "5814", "5815", "5816", "5817", "5818", "5819", "581a", "581b", "581c", "581d", "581e", "581f", "5820",
      "5821", "5822", "5823", "5824", "5825", "5826", "5827", "5828", "5829", "582a", "582b", "582c", "582d", "582e", "582f", "5830",
      "5831", "5832", "5833", "5834", "5835", "5836", "5837", "5838", "5839", "583a", "583b", "583c", "583d", "583e", "583f", "5840",
      "5841", "5842", "5843", "5844", "5845", "5846", "5847", "5848", "5849", "584a", "584b", "584c", "584d", "584e", "584f", "5850",
      "5851", "5852", "5853", "5854", "5855", "5856", "5857", "5858", "5859", "585a", "585b", "585c", "585d", "585e", "585f", "5860",
      "5861", "5862", "5863", "5864", "5865", "5866", "5867", "5868", "5869", "586a", "586b", "586c", "586d", "586e", "586f", "5870",
      "5871", "5872", "5873", "5874", "5875", "5876", "5877", "5878", "5879", "587a", "587b", "587c", "587d", "587e", "587f", "5880",
      "5881", "5882", "5883", "5884", "5885", "5886", "5887", "5888", "5889", "588a", "588b", "588c", "588d", "588e", "588f", "5890",
      "5891", "5892", "5893", "5894", "5895", "5896", "5897", "5898", "5899", "589a", "589b", "589c", "589d", "589e", "589f", "58a0",
      "58a1", "58a2", "58a3", "58a4", "58a5", "58a6", "58a7", "58a8", "58a9", "58aa", "58ab", "58ac", "58ad", "58ae", "58af", "58b0",
      "58b1", "58b2", "58b3", "58b4", "58b5", "58b6", "58b7", "58b8", "58b9", "58ba", "58bb", "58bc", "58bd", "58be", "58bf", "58c0",
      "58c1", "58c2", "58c3", "58c4", "58c5", "58c6", "58c7", "58c8", "58c9", "58ca", "58cb", "58cc", "58cd", "58ce", "58cf", "58d0",
      "58d1", "58d2", "58d3", "58d4", "58d5", "58d6", "58d7", "58d8", "58d9", "58da", "58db", "58dc", "58dd", "58de", "58df", "58e0",
      "58e1", "58e2", "58e3", "58e4", "58e5", "58e6", "58e7", "58e8", "58e9", "58ea", "58eb", "58ec", "58ed", "58ee", "58ef", "58f0",
      "58f1", "58f2", "58f3", "58f4", "58f5", "58f6", "58f7", "58f8", "58f9", "58fa", "58fb", "58fc", "58fd", "58fe", "58ff",
      "590001"
    };

    int value = -280;
    for (final String hex : expected) {
      final byte[] r = YAJBE_MAPPER.writeValueAsBytes(value);
      assertArrayEquals(HexFormat.of().parseHex(hex), r);
      value++;
    }
  }

  @Test
  public void testSmallInlineIntArray() throws JsonProcessingException {
    final int[] items = new int[562];
    int value = -280;
    for (int i = 0; value <= 281; i++) {
      items[i] = value;
      value++;
    }
    final byte[] r = YAJBE_MAPPER.writeValueAsBytes(items);
    assertEquals("2c280279000178ff78fe78fd78fc78fb78fa78f978f878f778f678f578f478f378f278f178f078ef78ee78ed78ec78eb78ea78e978e878e778e678e578e478e378e278e178e078df78de78dd78dc78db78da78d978d878d778d678d578d478d378d278d178d078cf78ce78cd78cc78cb78ca78c978c878c778c678c578c478c378c278c178c078bf78be78bd78bc78bb78ba78b978b878b778b678b578b478b378b278b178b078af78ae78ad78ac78ab78aa78a978a878a778a678a578a478a378a278a178a0789f789e789d789c789b789a7899789878977896789578947893789278917890788f788e788d788c788b788a7889788878877886788578847883788278817880787f787e787d787c787b787a7879787878777876787578747873787278717870786f786e786d786c786b786a7869786878677866786578647863786278617860785f785e785d785c785b785a7859785878577856785578547853785278517850784f784e784d784c784b784a7849784878477846784578447843784278417840783f783e783d783c783b783a7839783878377836783578347833783278317830782f782e782d782c782b782a7829782878277826782578247823782278217820781f781e781d781c781b781a7819781878177816781578147813781278117810780f780e780d780c780b780a780978087807780678057804780378027801780077767574737271706f6e6d6c6b6a69686766656463626160404142434445464748494a4b4c4d4e4f50515253545556575800580158025803580458055806580758085809580a580b580c580d580e580f5810581158125813581458155816581758185819581a581b581c581d581e581f5820582158225823582458255826582758285829582a582b582c582d582e582f5830583158325833583458355836583758385839583a583b583c583d583e583f5840584158425843584458455846584758485849584a584b584c584d584e584f5850585158525853585458555856585758585859585a585b585c585d585e585f5860586158625863586458655866586758685869586a586b586c586d586e586f5870587158725873587458755876587758785879587a587b587c587d587e587f5880588158825883588458855886588758885889588a588b588c588d588e588f5890589158925893589458955896589758985899589a589b589c589d589e589f58a058a158a258a358a458a558a658a758a858a958aa58ab58ac58ad58ae58af58b058b158b258b358b458b558b658b758b858b958ba58bb58bc58bd58be58bf58c058c158c258c358c458c558c658c758c858c958ca58cb58cc58cd58ce58cf58d058d158d258d358d458d558d658d758d858d958da58db58dc58dd58de58df58e058e158e258e358e458e558e658e758e858e958ea58eb58ec58ed58ee58ef58f058f158f258f358f458f558f658f758f858f958fa58fb58fc58fd58fe58ff590001", HexFormat.of().formatHex(r));
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

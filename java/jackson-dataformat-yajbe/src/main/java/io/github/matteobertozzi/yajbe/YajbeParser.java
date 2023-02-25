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
import java.math.BigDecimal;
import java.math.BigInteger;
import java.util.Arrays;
import java.util.List;
import java.util.Map;

import com.fasterxml.jackson.core.Base64Variant;
import com.fasterxml.jackson.core.JsonLocation;
import com.fasterxml.jackson.core.JsonStreamContext;
import com.fasterxml.jackson.core.JsonToken;
import com.fasterxml.jackson.core.ObjectCodec;
import com.fasterxml.jackson.core.Version;
import com.fasterxml.jackson.core.base.ParserMinimalBase;
import com.fasterxml.jackson.core.io.IOContext;

/**
 * {@link ParserMinimalBase} implementation that reads YAJBE encoded content.
 */
final class YajbeParser extends ParserMinimalBase {
  private final YajbeFieldNameReader fieldNameReader;
  private final YajbeReader stream;
  private final ObjectCodec codec;

  private boolean isClosed = false;

  YajbeParser(final IOContext ctxt, final int features, final ObjectCodec codec, final YajbeReader stream) {
    super(features);
    this.stream = stream;
    this.fieldNameReader = new YajbeFieldNameReader(stream);
    this.codec = codec;
  }

  void setInitialFieldNames(final String[] names) {
    fieldNameReader.setInitialFieldNames(names);
  }

  @Override
  public void close() {
    if (isClosed) return;

    isClosed = true;
  }

  @Override
  public boolean isClosed() {
    return isClosed;
  }

  // ====================================================================================================
  //  Reader Stack Related
  //   - array fixed length (STACK_FLAG_ARRAY | length)
  //   - array eof (STACK_FLAG_ARRAY | STACK_FLAG_EOF)
  //   - object fixed length (length)
  //   - object eof (STACK_FLAG_EOF)
  // stackState used to know if we have to call the stackStateHandler
  //   - array length = 0
  //   - array/object eof check
  //   - object field/value check
  // ====================================================================================================
  interface StackStateHandler {
    JsonToken nextToken() throws IOException;
  }

  private static final long STACK_FLAG_ARRAY = (1L << 62);
  private static final long STACK_FLAG_EOF = (1L << 61);
  private static final long STACK_MASK_LENGTH = 0x7fffffffL;
  private static final long STACK_MASK_INFO = 0x7fffffff_00000000L;

  private long[] stackItem = new long[32];
  private int stackSize = -1;

  private StackStateHandler stackStateHandler;
  private long stackState = Long.MAX_VALUE;
  private int stackObjectAvail;

  private void stackPush(final long newItem) {
    if (stackSize != -1) {
      final long item = stackItem[stackSize];
      if ((item & STACK_FLAG_EOF) != STACK_FLAG_EOF) {
        final long length = ((item & STACK_FLAG_ARRAY) == STACK_FLAG_ARRAY) ? stackState : stackObjectAvail;
        stackItem[stackSize] = (item & STACK_MASK_INFO) | length;
      }
    }

    if (++stackSize == stackItem.length) {
      stackItem = Arrays.copyOf(stackItem, stackSize + 16);
    }
    stackItem[stackSize] = newItem;
  }

  private void stackPop() {
    if (stackSize-- == 0) {
      this.stackState = Long.MAX_VALUE;
      return;
    }

    final long item = stackItem[stackSize];
    if ((item & STACK_FLAG_ARRAY) == STACK_FLAG_ARRAY) {
      if ((item & STACK_FLAG_EOF) == STACK_FLAG_EOF) {
        this.stackStateHandler = this::stackEofArrayStateHandler;
        this.stackState = 0;
      } else {
        this.stackStateHandler = this::stackFixedArrayStateHandler;
        this.stackState = (int) (item & STACK_MASK_LENGTH);
      }
    } else {
      if ((item & STACK_FLAG_EOF) == STACK_FLAG_EOF) {
        this.stackStateHandler = this::stackEofObjectStateHandler;
      } else {
        this.stackStateHandler = this::stackFixedObjectStateHandler;
        this.stackObjectAvail = (int) (item & STACK_MASK_LENGTH);
      }
      this.stackState = 0;
    }
  }

  // ---------------------------------------------------------------------------
  //  Reader Stack - Object Related
  // ---------------------------------------------------------------------------
  private void startFixedObject(final int head) throws IOException {
    final int length = stream.readItemCount(head);
    stackPush(length);
    this.stackStateHandler = this::stackFixedObjectStateHandler;
    this.stackState = 0;
    this.stackObjectAvail = length;
  }

  private JsonToken stackFixedObjectStateHandler() {
    this.stackState = 1;
    if (stackObjectAvail-- != 0) {
      return JsonToken.FIELD_NAME;
    }

    stackPop();
    return JsonToken.END_OBJECT;
  }

  private void startEofObject() {
    stackPush(STACK_FLAG_EOF);
    this.stackStateHandler = this::stackEofObjectStateHandler;
    this.stackState = 0;
  }

  private JsonToken stackEofObjectStateHandler() throws IOException {
    this.stackState = 1;
    if (stream.peek() != 1) return JsonToken.FIELD_NAME;

    stream.read();
    stackPop();
    return JsonToken.END_OBJECT;
  }

  // ---------------------------------------------------------------------------
  //  Reader Stack - Array Related
  // ---------------------------------------------------------------------------
  private void startFixedArray(final int head) throws IOException {
    final int length = stream.readItemCount(head);
    stackPush(STACK_FLAG_ARRAY | length);
    this.stackStateHandler = this::stackFixedArrayStateHandler;
    this.stackState = length;
  }

  private JsonToken stackFixedArrayStateHandler() {
    stackPop();
    return JsonToken.END_ARRAY;
  }

  private void startEofArray() {
    stackPush(STACK_FLAG_ARRAY | STACK_FLAG_EOF);
    this.stackStateHandler = this::stackEofArrayStateHandler;
    this.stackState = 0;
  }

  private JsonToken stackEofArrayStateHandler() throws IOException {
    this.stackState = 0;
    if (stream.peek() != 1) return null;

    stream.read();
    stackPop();
    return JsonToken.END_ARRAY;
  }

  // =====================================================================================
  //  NOTE: to avoid too many ifs, we pre-build a map with the tokens.
  //  so we can find the token just by looking up TOKEN_MAP[head]
  private static final byte[] TOKEN_MAP = new byte[] {
    0, -1, 1, 2,
    12, 13, 14, 15,
    8, 9, 9,
    -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
    16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 17,
    18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 19,
    3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 4, 4, 4, 4, 4, 4, 4, 4,
    3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 5, 5, 5, 5, 5, 5, 5, 5,
    10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 11, 11, 11, 11,
    6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 7, 7, 7, 7
  };

  private static final int TOKEN_NULL         = 0;
  private static final int TOKEN_FALSE        = 1;
  private static final int TOKEN_TRUE         = 2;
  private static final int TOKEN_INT_SMALL    = 3;
  private static final int TOKEN_INT_POSITIVE = 4;
  private static final int TOKEN_INT_NEGATIVE = 5;
  private static final int TOKEN_SMALL_STRING = 6;
  private static final int TOKEN_STRING       = 7;
  private static final int TOKEN_ENUM_CONFIG  = 8;
  private static final int TOKEN_ENUM_STRING  = 9;
  private static final int TOKEN_SMALL_BYTES  = 10;
  private static final int TOKEN_BYTES        = 11;
  private static final int TOKEN_FLOAT_VLE    = 12;
  private static final int TOKEN_FLOAT_32     = 13;
  private static final int TOKEN_FLOAT_64     = 14;
  private static final int TOKEN_BIG_DECIMAL  = 15;
  private static final int TOKEN_ARRAY        = 16;
  private static final int TOKEN_ARRAY_EOF    = 17;
  private static final int TOKEN_OBJECT       = 18;
  private static final int TOKEN_OBJECT_EOF   = 19;

  private static final JsonToken[] JSON_TOKEN_MAP = new JsonToken[] {
    JsonToken.VALUE_NULL,
    JsonToken.VALUE_FALSE,
    JsonToken.VALUE_TRUE,
    JsonToken.VALUE_NUMBER_INT,       // small int
    JsonToken.VALUE_NUMBER_INT,       // int positive
    JsonToken.VALUE_NUMBER_INT,       // int negative
    JsonToken.VALUE_STRING,           // small string
    JsonToken.VALUE_STRING,           // string
    null,                             // enum config
    JsonToken.VALUE_STRING,           // enum string
    JsonToken.VALUE_EMBEDDED_OBJECT,  // small bytes
    JsonToken.VALUE_EMBEDDED_OBJECT,  // bytes
    JsonToken.VALUE_NUMBER_FLOAT,     // float vle
    JsonToken.VALUE_NUMBER_FLOAT,     // float32
    JsonToken.VALUE_NUMBER_FLOAT,     // float64
    JsonToken.VALUE_NUMBER_FLOAT,     // big decimal
    JsonToken.START_ARRAY,            // fixed array
    JsonToken.START_ARRAY,            // eof array
    JsonToken.START_OBJECT,           // fixed object
    JsonToken.START_OBJECT,           // eof object
  };

  @Override
  public JsonToken nextToken() throws IOException {
    if (stackState-- == 0) {
      if ((_currToken = stackStateHandler.nextToken()) != null) {
        return _currToken;
      }
    }

    final int head = stream.read();
    final int tokenId = TOKEN_MAP[head];
    switch (tokenId) {
      case TOKEN_INT_SMALL -> stream.decodeSmallInt(head);
      case TOKEN_INT_POSITIVE -> stream.decodeIntPositive(head);
      case TOKEN_INT_NEGATIVE -> stream.decodeIntNegative(head);
      case TOKEN_SMALL_STRING -> stream.decodeSmallString(head);
      case TOKEN_STRING -> stream.decodeString(head);
      case TOKEN_SMALL_BYTES -> stream.decodeSmallBytes(head);
      case TOKEN_BYTES -> stream.decodeBytes(head);
      case TOKEN_FLOAT_VLE -> stream.decodeFloatVle();
      case TOKEN_FLOAT_32 -> stream.decodeFloat32();
      case TOKEN_FLOAT_64 -> stream.decodeFloat64();
      case TOKEN_BIG_DECIMAL -> stream.decodeBigDecimal();
      case TOKEN_ARRAY -> startFixedArray(head);
      case TOKEN_ARRAY_EOF -> startEofArray();
      case TOKEN_OBJECT -> startFixedObject(head);
      case TOKEN_OBJECT_EOF -> startEofObject();
    }
    return _currToken = JSON_TOKEN_MAP[tokenId];
  }

  @Override
  protected void _handleEOF() {
    // TODO Auto-generated method stub
  }

  @Override
  public String getCurrentName() throws IOException {
    return fieldNameReader.read();
  }

  @Override
  public JsonStreamContext getParsingContext() {
    return null;
  }

  @Override
  public JsonLocation getTokenLocation() {
    return null;
  }

  @Override
  public void overrideCurrentName(final String name) {
    throw new UnsupportedOperationException();
  }

  @Override
  public String getText() {
    return stream.stringValue();
  }

  @Override
  public boolean hasTextCharacters() {
    return false;
  }

  @Override
  public char[] getTextCharacters() {
    throw new UnsupportedOperationException();
  }

  @Override
  public int getTextOffset() {
    throw new UnsupportedOperationException();
  }

  @Override
  public int getTextLength() {
    throw new UnsupportedOperationException();
  }

  @Override
  public byte[] getBinaryValue(final Base64Variant b64variant) {
    return stream.bytesValue().toByteArray();
  }

  @Override
  public Object getEmbeddedObject() {
    return switch (_currToken) {
      case START_ARRAY -> List.of();
      case START_OBJECT -> Map.of();
      case VALUE_EMBEDDED_OBJECT -> stream.bytesValue().toByteArray();
      default -> throw new IllegalArgumentException();
    };
  }

  @Override
  public ObjectCodec getCodec() {
    return codec;
  }

  @Override
  public void setCodec(final ObjectCodec oc) {
    throw new UnsupportedOperationException();
  }

  @Override
  public Version version() {
    throw new UnsupportedOperationException();
  }

  @Override
  public JsonLocation getCurrentLocation() {
    throw new UnsupportedOperationException();
  }

  @Override
  public Number getNumberValue() {
    return switch (stream.numberType()) {
      case INT -> stream.intValue();
      case LONG -> stream.longValue();
      case FLOAT -> stream.floatValue();
      case DOUBLE -> stream.doubleValue();
      case BIG_INTEGER -> stream.bigInteger();
      case BIG_DECIMAL -> stream.bigDecimal();
    };
  }

  @Override
  public NumberType getNumberType() {
    return stream.numberType();
  }

  @Override
  public int getIntValue() {
    return switch (stream.numberType()) {
      case INT -> stream.intValue();
      case LONG -> Math.toIntExact(stream.longValue());
      case FLOAT -> (int)stream.floatValue();
      case DOUBLE -> (int)stream.doubleValue();
      case BIG_INTEGER -> stream.bigInteger().intValueExact();
      case BIG_DECIMAL -> stream.bigDecimal().intValueExact();
    };
  }

  @Override
  public long getLongValue() {
    return switch (stream.numberType()) {
      case INT -> stream.intValue();
      case LONG -> stream.longValue();
      case FLOAT -> (long)stream.floatValue();
      case DOUBLE -> (long)stream.doubleValue();
      case BIG_INTEGER -> stream.bigInteger().longValueExact();
      case BIG_DECIMAL -> stream.bigDecimal().longValueExact();
    };
  }

  @Override
  public BigInteger getBigIntegerValue() {
    return switch (stream.numberType()) {
      case INT -> BigInteger.valueOf(stream.intValue());
      case LONG -> BigInteger.valueOf(stream.longValue());
      case FLOAT -> BigInteger.valueOf((long)stream.floatValue());
      case DOUBLE -> BigInteger.valueOf((long)stream.doubleValue());
      case BIG_INTEGER -> stream.bigInteger();
      case BIG_DECIMAL -> stream.bigDecimal().toBigIntegerExact();
    };
  }

  @Override
  public float getFloatValue() {
    return switch (stream.numberType()) {
      case INT -> stream.intValue();
      case LONG -> stream.longValue();
      case FLOAT -> stream.floatValue();
      case DOUBLE -> (float)stream.doubleValue();
      case BIG_INTEGER -> stream.bigInteger().floatValue();
      case BIG_DECIMAL -> stream.bigDecimal().floatValue();
    };
  }

  @Override
  public double getDoubleValue() {
    return switch (stream.numberType()) {
      case INT -> stream.intValue();
      case LONG -> stream.longValue();
      case FLOAT -> stream.floatValue();
      case DOUBLE -> stream.doubleValue();
      case BIG_INTEGER -> stream.bigInteger().doubleValue();
      case BIG_DECIMAL -> stream.bigDecimal().doubleValue();
    };
  }

  @Override
  public BigDecimal getDecimalValue() {
    return switch (stream.numberType()) {
      case INT -> BigDecimal.valueOf(stream.intValue());
      case LONG -> BigDecimal.valueOf(stream.longValue());
      case FLOAT -> BigDecimal.valueOf(stream.floatValue());
      case DOUBLE -> BigDecimal.valueOf(stream.doubleValue());
      case BIG_INTEGER -> new BigDecimal(stream.bigInteger());
      case BIG_DECIMAL -> stream.bigDecimal();
    };
  }

  public static void main(final String[] args) {
    final int[] tokens = new int[256];
    for (int i = 0; i <= 0xff; ++i) {
      final int head = i & 0xff;
      if ((head & 0b11_000000) == 0b11_000000) {
        final int w = head & 0b111111;
        if (w <= 59) {
          tokens[i] = TOKEN_SMALL_STRING;
        } else {
          tokens[i] = TOKEN_STRING;
        }
      } else if ((head & 0b10_000000) == 0b10_000000) {
        final int w = head & 0b111111;
        if (w <= 59) {
          tokens[i] = TOKEN_SMALL_BYTES;
        } else {
          tokens[i] = TOKEN_BYTES;
        }
      } else if ((head & 0b010_00000) == 0b010_00000) {
        final int w = head & 0b11111;
        if (w < 24) {
          tokens[i] = TOKEN_INT_SMALL;
        } else if ((head & 0b011_00000) == 0b011_00000) {
          tokens[i] = TOKEN_INT_NEGATIVE;
        } else {
          tokens[i] = TOKEN_INT_POSITIVE;
        }
      } else if ((head & 0b0011_1111) == 0b0011_1111) {
        tokens[i] = TOKEN_OBJECT_EOF;
      } else if ((head & 0b0011_0000) == 0b0011_0000) {
        tokens[i] = TOKEN_OBJECT;
      } else if ((head & 0b0010_1111) == 0b0010_1111) {
        tokens[i] = TOKEN_ARRAY_EOF;
      } else if ((head & 0b0010_0000) == 0b0010_0000) {
        tokens[i] = TOKEN_ARRAY;
      } else if ((head & 0b0001_0000) == 0b0001_0000) {
        tokens[i] = -1;
      } else if ((head & 0b00001_000) == 0b00001_000) {
        switch (head) {
          case 0b00001000 -> tokens[i] = TOKEN_ENUM_CONFIG;
          case 0b00001001 -> tokens[i] = TOKEN_ENUM_STRING;
          case 0b00001010 -> tokens[i] = TOKEN_ENUM_STRING;
          default -> tokens[i] = -1;
        }
      } else switch (head) {
        case 0b00000000 -> tokens[i] = TOKEN_NULL;
        case 0b00000001 -> tokens[i] = -1;
        case 0b00000010 -> tokens[i] = TOKEN_FALSE;
        case 0b00000011 -> tokens[i] = TOKEN_TRUE;
        case 0b00000100 -> tokens[i] = TOKEN_FLOAT_VLE;
        case 0b00000101 -> tokens[i] = TOKEN_FLOAT_32;
        case 0b00000110 -> tokens[i] = TOKEN_FLOAT_64;
        case 0b00000111 -> tokens[i] = TOKEN_BIG_DECIMAL;
        default -> throw new IllegalArgumentException("unhandled " + Integer.toBinaryString(i));
      }
    }

    System.out.println(Arrays.toString(tokens));
  }
}

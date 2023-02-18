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

public class YajbeParser extends ParserMinimalBase {
  private final YajbeFieldNameReader fieldNameReader;
  private final YajbeReader stream;
  private final ObjectCodec codec;
  private boolean isClosed = false;

  public YajbeParser(final IOContext ctxt, final int features, final ObjectCodec codec, final YajbeReader stream) {
    super(features);
    this.stream = stream;
    this.fieldNameReader = new YajbeFieldNameReader(stream);
    this.codec = codec;
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

  private int[] stackBlocks = new int[32]; // fields/length
  private int stackSize = 0;

  private void startBlock(final int fields, final int length) {
    if (stackSize == stackBlocks.length) {
      stackBlocks = Arrays.copyOf(stackBlocks, stackSize + 16);
    }
    stackBlocks[stackSize++] = fields;
    stackBlocks[stackSize++] = length;
  }

  private JsonToken endBlock() {
    stackSize -= 2;
    final boolean isObject = stackBlocks[stackSize] >= 0;
    return isObject ? JsonToken.END_OBJECT : JsonToken.END_ARRAY;
  }

  private boolean nextIsObjectField() throws IOException {
    final int blockObj = stackBlocks[stackSize - 2]++;
    return blockObj >= 0 && ((blockObj & 1) == 0) && stream.peek() != 0b0000000_1;
  }

  @Override
  public JsonToken nextToken() throws IOException {
    if (stackSize != 0) {
      if (stackBlocks[stackSize - 1]-- == 0) {
        return _currToken = endBlock();
      }

      if (nextIsObjectField()) {
        return _currToken = JsonToken.FIELD_NAME;
      }
    }

    final int head = stream.read();
    if ((head & 0b11_000000) == 0b11_000000) {
      stream.decodeString(head);
      return _currToken = JsonToken.VALUE_STRING;
    } else if ((head & 0b10_000000) == 0b10_000000) {
      stream.decodeBytes(head);
      return _currToken = JsonToken.VALUE_EMBEDDED_OBJECT;
    } else if ((head & 0b010_00000) == 0b010_00000) {
      stream.decodeInt(head);
      return _currToken = JsonToken.VALUE_NUMBER_INT;
    } else if ((head & 0b0011_0000) == 0b0011_0000) {
      startBlock(0, stream.readItemCount(head));
      return _currToken = JsonToken.START_OBJECT;
    } else if ((head & 0b0010_0000) == 0b0010_0000) {
      startBlock(Integer.MIN_VALUE, stream.readItemCount(head));
      return _currToken = JsonToken.START_ARRAY;
    } else if ((head & 0b000001_00) == 0b000001_00) {
      stream.decodeFloat(head);
      return _currToken = JsonToken.VALUE_NUMBER_FLOAT;
    } else return switch (head) {
      case 0b00000000 -> _currToken = JsonToken.VALUE_NULL;
      case 0b00000001 -> _currToken = endBlock();
      case 0b00000010 -> _currToken = JsonToken.VALUE_FALSE;
      case 0b00000011 -> _currToken = JsonToken.VALUE_TRUE;
      default -> throw new IOException("unsupported head " + Integer.toBinaryString(head));
    };
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
}

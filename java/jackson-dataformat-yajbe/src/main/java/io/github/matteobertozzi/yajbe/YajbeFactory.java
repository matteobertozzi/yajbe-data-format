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

import java.io.InputStream;
import java.io.OutputStream;
import java.io.Reader;
import java.io.Writer;

import com.fasterxml.jackson.core.JsonFactory;
import com.fasterxml.jackson.core.JsonParser;
import com.fasterxml.jackson.core.io.IOContext;

public class YajbeFactory extends JsonFactory {
  private static final long serialVersionUID = 1; // 2.6

  @Override public String getFormatName() { return "YAJBE"; }

  @Override public boolean requiresPropertyOrdering() { return false; }
  @Override public boolean canHandleBinaryNatively() { return true; }
  @Override public boolean canUseCharArrays() { return false; }

  @Override
  protected YajbeParser _createParser(final InputStream in, final IOContext ctxt) {
    return new YajbeParser(ctxt, _parserFeatures, _objectCodec, YajbeReader.fromStream(in));
  }

  @Override
  protected JsonParser _createParser(final Reader r, final IOContext ctxt) {
    throw new UnsupportedOperationException();
  }

  @Override
  protected JsonParser _createParser(final char[] data, final int offset, final int len, final IOContext ctxt,
                                     final boolean recyclable) {
    throw new UnsupportedOperationException();
  }

  @Override
  protected YajbeParser _createParser(final byte[] data, final int offset, final int len, final IOContext ctxt) {
    return new YajbeParser(ctxt, _parserFeatures, _objectCodec, YajbeReader.fromBytes(data, offset, len));
  }

  @Override
  protected YajbeGenerator _createGenerator(final Writer out, final IOContext ctxt) {
    throw new UnsupportedOperationException();
  }

  @Override
  protected YajbeGenerator _createUTF8Generator(final OutputStream out, final IOContext ctxt) {
    return new YajbeGenerator(ctxt, _generatorFeatures, _objectCodec, out);
  }
}

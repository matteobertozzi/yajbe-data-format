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

import java.io.DataInput;
import java.io.DataOutput;
import java.io.File;
import java.io.IOException;
import java.io.InputStream;
import java.io.OutputStream;
import java.io.Reader;
import java.io.Writer;
import java.net.URL;

import com.fasterxml.jackson.core.FormatSchema;
import com.fasterxml.jackson.core.JsonEncoding;
import com.fasterxml.jackson.core.JsonFactory;
import com.fasterxml.jackson.core.JsonGenerator;
import com.fasterxml.jackson.core.JsonParser;
import com.fasterxml.jackson.core.PrettyPrinter;
import com.fasterxml.jackson.databind.DeserializationConfig;
import com.fasterxml.jackson.databind.InjectableValues;
import com.fasterxml.jackson.databind.JavaType;
import com.fasterxml.jackson.databind.JsonDeserializer;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.databind.ObjectReader;
import com.fasterxml.jackson.databind.ObjectWriter;
import com.fasterxml.jackson.databind.SerializationConfig;
import com.fasterxml.jackson.databind.deser.DataFormatReaders;

/**
 * Specialized {@link ObjectMapper} to use with YAJBE data format.
 */
public class YajbeMapper extends ObjectMapper {
  private static final long serialVersionUID = 1L;

  public static final String CONFIG_MAP_FIELD_NAMES = "map.field.names";

  /**
   * Default constructor, which will construct the default {@link YajbeFactory}
   */
  public YajbeMapper() {
    this(new YajbeFactory());
  }

  /**
   * Constructs an instance that uses specified {@link YajbeFactory}
   * for constructing necessary {@link YajbeParser}s and/or
   * {@link YajbeGenerator}s.
   *
   * @param factory the {@link YajbeFactory}
   */
  public YajbeMapper(final YajbeFactory factory) {
    super(factory);
    // enable(SerializationFeature.ORDER_MAP_ENTRIES_BY_KEYS);
  }

  // ==========================================================================================
  // Writer
  // ==========================================================================================
  protected ObjectWriter _newWriter(final SerializationConfig config) {
    return new YajbeWriter(this, config);
  }

  protected ObjectWriter _newWriter(final SerializationConfig config, final FormatSchema schema) {
    return new YajbeWriter(this, config, schema);
  }

  protected ObjectWriter _newWriter(final SerializationConfig config, final JavaType rootType, final PrettyPrinter pp) {
    return new YajbeWriter(this, config, rootType, pp);
  }

  private static final class YajbeWriter extends ObjectWriter {
    YajbeWriter(final ObjectMapper mapper, final SerializationConfig config) {
      super(mapper, config);
    }

    public YajbeWriter(final ObjectMapper mapper, final SerializationConfig config, final FormatSchema schema) {
      super(mapper, config, schema);
    }

    public YajbeWriter(final YajbeMapper mapper, final SerializationConfig config, final JavaType rootType,
        final PrettyPrinter pp) {
      super(mapper, config, rootType, pp);
    }

    public YajbeWriter(final ObjectWriter base, final JsonFactory f) {
      super(base, f);
    }

    public YajbeWriter(final ObjectWriter base, final SerializationConfig config) {
      super(base, config);
    }

    public YajbeWriter(final ObjectWriter writer, final SerializationConfig _config,
        final GeneratorSettings genSettings, final Prefetch prefetch) {
      super(writer, _config, genSettings, prefetch);
    }

    @Override
    protected ObjectWriter _new(final ObjectWriter base, final JsonFactory f) {
      return new YajbeWriter(base, f);
    }

    @Override
    protected ObjectWriter _new(final ObjectWriter base, final SerializationConfig config) {
      if (config == _config) return this;
      return new YajbeWriter(base, config);
    }

    @Override
    protected ObjectWriter _new(final GeneratorSettings genSettings, final Prefetch prefetch) {
      if ((_generatorSettings == genSettings) && (_prefetch == prefetch)) return this;
      return new YajbeWriter(this, _config, genSettings, prefetch);
    }

    @Override
    public JsonGenerator createGenerator(final OutputStream out) throws IOException {
      return _configureAttrs(super.createGenerator(out));
    }

    @Override
    public JsonGenerator createGenerator(final OutputStream out, final JsonEncoding enc) throws IOException {
      return _configureAttrs(super.createGenerator(out, enc));
    }

    @Override
    public JsonGenerator createGenerator(final Writer w) throws IOException {
      return _configureAttrs(super.createGenerator(w));
    }

    @Override
    public JsonGenerator createGenerator(final File outputFile, final JsonEncoding enc) throws IOException {
      return _configureAttrs(super.createGenerator(outputFile, enc));
    }

    @Override
    public JsonGenerator createGenerator(final DataOutput out) throws IOException {
      return _configureAttrs(super.createGenerator(out));
    }

    private JsonGenerator _configureAttrs(final JsonGenerator g) {
      final Object initialFields = _config.getAttributes().getAttribute(CONFIG_MAP_FIELD_NAMES);
      if (initialFields != null) {
        if (initialFields instanceof final String[] names) {
          final YajbeGenerator yg = (YajbeGenerator) g;
          yg.setInitialFieldNames(names);
        } else {
          throw new IllegalArgumentException("expected String[] for " + CONFIG_MAP_FIELD_NAMES + ": " + initialFields);
        }
      }
      return g;
    }
  }

  // ==========================================================================================
  // Reader
  // ==========================================================================================
  protected ObjectReader _newReader(final DeserializationConfig config) {
    return new YajbeReader(this, config);
  }

  protected ObjectReader _newReader(final DeserializationConfig config,
      final JavaType valueType, final Object valueToUpdate,
      final FormatSchema schema, final InjectableValues injectableValues) {
    return new YajbeReader(this, config, valueType, valueToUpdate, schema, injectableValues);
  }

  private static class YajbeReader extends ObjectReader {
    public YajbeReader(final ObjectMapper mapper, final DeserializationConfig config) {
      super(mapper, config);
    }

    public YajbeReader(final ObjectMapper mapper, final DeserializationConfig config,
        final JavaType valueType, final Object valueToUpdate,
        final FormatSchema schema, final InjectableValues injectableValues) {
      super(mapper, config, valueType, valueToUpdate, schema, injectableValues);
    }

    public YajbeReader(final ObjectReader base, final JsonFactory f) {
      super(base, f);
    }

    public YajbeReader(final ObjectReader base, final DeserializationConfig config) {
      super(base, config);
    }

    public YajbeReader(final ObjectReader base, final DeserializationConfig config, final JavaType valueType,
        final JsonDeserializer<Object> rootDeser, final Object valueToUpdate, final FormatSchema schema,
        final InjectableValues injectableValues, final DataFormatReaders dataFormatReaders) {
      super(base, config, valueType, rootDeser, valueToUpdate, schema, injectableValues, dataFormatReaders);
    }

    @Override
    protected ObjectReader _new(final ObjectReader base, final JsonFactory f) {
      return new YajbeReader(base, f);
    }

    @Override
    protected ObjectReader _new(final ObjectReader base, final DeserializationConfig config) {
      return new YajbeReader(base, config);
    }

    @Override
    protected ObjectReader _new(final ObjectReader base, final DeserializationConfig config,
        final JavaType valueType, final JsonDeserializer<Object> rootDeser, final Object valueToUpdate,
        final FormatSchema schema, final InjectableValues injectableValues,
        final DataFormatReaders dataFormatReaders) {
      return new YajbeReader(base, config, valueType, rootDeser, valueToUpdate,
          schema, injectableValues, dataFormatReaders);
    }

    @Override
    public JsonParser createParser(final File src) throws IOException {
      return _configureAttrs(super.createParser(src));
    }

    @Override
    public JsonParser createParser(final URL src) throws IOException {
      return _configureAttrs(super.createParser(src));
    }

    @Override
    public JsonParser createParser(final InputStream in) throws IOException {
      return _configureAttrs(super.createParser(in));
    }

    @Override
    public JsonParser createParser(final Reader r) throws IOException {
      return _configureAttrs(super.createParser(r));
    }

    @Override
    public JsonParser createParser(final byte[] content) throws IOException {
      return _configureAttrs(super.createParser(content));
    }

    @Override
    public JsonParser createParser(final byte[] content, final int offset, final int len) throws IOException {
      return _configureAttrs(super.createParser(content, offset, len));
    }

    @Override
    public JsonParser createParser(final String content) throws IOException {
      return _configureAttrs(super.createParser(content));
    }

    @Override
    public JsonParser createParser(final char[] content) throws IOException {
      return _configureAttrs(super.createParser(content));
    }

    @Override
    public JsonParser createParser(final char[] content, final int offset, final int len) throws IOException {
      return _configureAttrs(super.createParser(content, offset, len));
    }

    @Override
    public JsonParser createParser(final DataInput content) throws IOException {
      return _configureAttrs(super.createParser(content));
    }

    @Override
    public JsonParser createNonBlockingByteArrayParser() throws IOException {
      return _configureAttrs(super.createNonBlockingByteArrayParser());
    }

    private JsonParser _configureAttrs(final JsonParser p) {
      final Object initialFields = _config.getAttributes().getAttribute(CONFIG_MAP_FIELD_NAMES);
      if (initialFields != null) {
        if (initialFields instanceof final String[] names) {
          final YajbeParser yp = (YajbeParser) p;
          yp.setInitialFieldNames(names);
        } else {
          throw new IllegalArgumentException("expected String[] for " + CONFIG_MAP_FIELD_NAMES + ": " + initialFields);
        }
      }
      return p;
    }
  }
}

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

/**
 * Interface implemented by enum mapping algorithms.
 */
public interface YajbeEnumMapping {
  /** Minimum length for the string to be taken in consideration for enum mapping */
  int MIN_ENUM_STRING_LENGTH = 3;

  /** Maximum length of the Enum Index */
  int MAX_INDEX_LENGTH = 0xffff;

  /**
   * @param index the index of the mapped string to lookup
   * @return the string mapped at the given index
   */
  String get(int index);

  /**
   * @param key the string to map
   * @return -1 if the string is not yet indexed, otherwise the index of the mapped string
   */
  int add(String key);

  /**
   * Creates an instance of the enum mapping algorithm given the configuration.
   * @param config the enum mapping algo configuration
   * @return the enum mapping instance
   */
  static YajbeEnumMapping fromConfig(final YajbeEnumMappingConfig config) {
    if (config instanceof final YajbeEnumLruMappingConfig lruConfig) {
      return new YajbeEnumLruMapping(lruConfig.lruSize(), lruConfig.minFreq());
    }
    throw new IllegalArgumentException("invalid config " + config);
  }

  /**
   * Base class to for the enum mapping config
   */
  interface YajbeEnumMappingConfig {
    /** types of enum maping */
    enum Type {
      /** mapping algo that uses an LRU to keep track of most common strings */
      LRU
    }

    /** @return The type of the enum mapping */
    Type type();
  }

  /**
   * Enum Mapping algorithm using an LRU
   * @param lruSize the size of LRU should be a power of 2 and larger than your string repetition
   * @param minFreq the minimum frequency after which the text will be added to the index
   */
  record YajbeEnumLruMappingConfig (int lruSize, int minFreq) implements YajbeEnumMappingConfig {
    public Type type() { return Type.LRU; }
  }
}

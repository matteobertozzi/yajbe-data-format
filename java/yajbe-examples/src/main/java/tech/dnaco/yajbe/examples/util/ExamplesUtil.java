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

package tech.dnaco.yajbe.examples.util;

import java.io.ByteArrayInputStream;
import java.io.ByteArrayOutputStream;
import java.io.IOException;
import java.util.concurrent.TimeUnit;
import java.util.zip.GZIPInputStream;
import java.util.zip.GZIPOutputStream;

import com.fasterxml.jackson.annotation.JsonAutoDetect.Visibility;
import com.fasterxml.jackson.annotation.JsonInclude;
import com.fasterxml.jackson.annotation.PropertyAccessor;
import com.fasterxml.jackson.databind.DeserializationFeature;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.databind.SerializationFeature;

public final class ExamplesUtil {
  private ExamplesUtil() {
    // no-op
  }

  // ===============================================================================================
  //  Jackson Util
  // ===============================================================================================
  public static ObjectMapper newObjectMapper(final ObjectMapper mapper) {
    mapper.setVisibility(PropertyAccessor.FIELD, Visibility.ANY);
    mapper.setVisibility(PropertyAccessor.GETTER, Visibility.NONE);
    mapper.setVisibility(PropertyAccessor.IS_GETTER, Visibility.NONE);

    // --- Deserialization ---
    // Just ignore unknown fields, don't stop parsing
    mapper.configure(DeserializationFeature.FAIL_ON_UNKNOWN_PROPERTIES, false);
    // Trying to deserialize value into an enum, don't fail on unknown value, use
    // null instead
    mapper.configure(DeserializationFeature.READ_UNKNOWN_ENUM_VALUES_AS_NULL, true);

    // --- Serialization ---
    // Don't include properties with null value in JSON output
    mapper.setSerializationInclusion(JsonInclude.Include.NON_NULL);
    // Use default pretty printer
    mapper.configure(SerializationFeature.INDENT_OUTPUT, false);
    mapper.configure(SerializationFeature.FAIL_ON_EMPTY_BEANS, false);

    //final String JSON_DATE_FORMAT_PATTERN = "YYYYMMddHHmmss";
    //mapper.setDateFormat(new SimpleDateFormat(JSON_DATE_FORMAT_PATTERN));

    // mapper.setAnnotationIntrospector(new ExtentedAnnotationIntrospector());
    return mapper;
  }

  // ===============================================================================================
  //  Gzip Util
  // ===============================================================================================
  public static byte[] compress(final byte[] data) throws IOException {
    try (ByteArrayOutputStream out = new ByteArrayOutputStream(data.length)) {
      try (GZIPOutputStream gz = new GZIPOutputStream(out)) {
        gz.write(data);
      }
      return out.toByteArray();
    }
  }

  public static byte[] decompress(final byte[] gzData) throws IOException {
    try (ByteArrayInputStream in = new ByteArrayInputStream(gzData)) {
      try (GZIPInputStream gz = new GZIPInputStream(in)) {
        return gz.readAllBytes();
      }
    }
  }

  // ===============================================================================================
  //  Dummy Bench Util (use jmh for a proper bench)
  // ===============================================================================================
  public interface BenchRunnable {
    Object run() throws Exception;
  }

  public static long runBench(final String name, final long count, final BenchRunnable runnable) throws Exception {
    final long startTime = System.nanoTime();
    for (long i = 0; i < count; ++i) {
      final Object r = runnable.run();
    }
    final long elapsed = System.nanoTime() - startTime;
    System.out.printf("[BENCH] %20s - %s runs took %s %s%n",
        name, humanCount(count), humanTimeNanos(elapsed), humanRate(count, elapsed, TimeUnit.NANOSECONDS));
    return elapsed;
  }

  // ===============================================================================================
  //  Humans Util
  // ===============================================================================================
  public static String humanSize(final long size) {
    if (size >= (1L << 60)) return String.format("%.2fEiB", (float) size / (1L << 60));
    if (size >= (1L << 50)) return String.format("%.2fPiB", (float) size / (1L << 50));
    if (size >= (1L << 40)) return String.format("%.2fTiB", (float) size / (1L << 40));
    if (size >= (1L << 30)) return String.format("%.2fGiB", (float) size / (1L << 30));
    if (size >= (1L << 20)) return String.format("%.2fMiB", (float) size / (1L << 20));
    if (size >= (1L << 10)) return String.format("%.2fKiB", (float) size / (1L << 10));
    return size > 0 ? size + "bytes" : "0";
  }

  public static String humanCount(final long size) {
    if (size >= 1000000) return String.format("%.2fM", (float) size / 1000000);
    if (size >= 1000) return String.format("%.2fK", (float) size / 1000);
    return Long.toString(size);
  }

  public static String humanRate(final double rate) {
    if (rate >= 1000000000000.0) return String.format("%.2fT/sec", rate / 1000000000000.0);
    if (rate >= 1000000000.0) return String.format("%.2fG/sec", rate / 1000000000.0);
    if (rate >= 1000000.0) return String.format("%.2fM/sec", rate / 1000000.0);
    if (rate >= 1000.0) return String.format("%.2fK/sec", rate / 1000.0f);
    return String.format("%.2f/sec", rate);
  }

  public static String humanRate(final long count, final long duration, final TimeUnit unit) {
    final double sec = unit.toNanos(duration) / 1000000000.0;
    return humanRate(count / sec);
  }

  public static String humanTimeNanos(final long timeNs) {
    if (timeNs < 1000) return (timeNs < 0) ? "unkown" : timeNs + "ns";
    return humanTimeMicros(timeNs / 1000);
  }

  public static String humanTimeMicros(final long timeUs) {
    if (timeUs < 1000) return (timeUs < 0) ? "unkown" : timeUs + "us";
    return humanTimeMillis(timeUs / 1000);
  }

  public static String humanTimeMillis(final long timeDiff) {
    return humanTime(timeDiff, TimeUnit.MILLISECONDS);
  }

  public static String humanTime(final long timeDiff, final TimeUnit unit) {
    final long msec = unit.toMillis(timeDiff);
    if (msec == 0) {
      final long micros = unit.toMicros(timeDiff);
      if (micros > 0)
        return micros + "us";
      return unit.toNanos(timeDiff) + "ns";
    }

    if (msec < 1000) {
      return msec + "ms";
    }

    final long hours = msec / (60 * 60 * 1000);
    long rem = (msec % (60 * 60 * 1000));
    final long minutes = rem / (60 * 1000);
    rem = rem % (60 * 1000);
    final float seconds = rem / 1000.0f;

    if ((hours > 0) || (minutes > 0)) {
      final StringBuilder buf = new StringBuilder(32);
      if (hours > 0) {
        buf.append(hours);
        buf.append("hrs, ");
      }
      if (minutes > 0) {
        buf.append(minutes);
        buf.append("min, ");
      }

      final String humanTime;
      if (seconds > 0) {
        buf.append(String.format("%.2fsec", seconds));
        humanTime = buf.toString();
      } else {
        humanTime = buf.substring(0, buf.length() - 2);
      }

      if (hours > 24) {
        return String.format("%s (%.1f days)", humanTime, (hours / 24.0));
      }
      return humanTime;
    }

    return String.format((seconds % 1) != 0 ? "%.4fsec" : "%.0fsec", seconds);
  }
}

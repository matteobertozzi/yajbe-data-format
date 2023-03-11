package io.github.matteobertozzi.yajbe.examples.util;

import java.io.File;
import java.io.FileInputStream;
import java.io.IOException;
import java.util.List;
import java.util.zip.GZIPInputStream;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.databind.json.JsonMapper;
import com.fasterxml.jackson.dataformat.cbor.databind.CBORMapper;
import com.fasterxml.jackson.dataformat.xml.XmlMapper;

import io.github.matteobertozzi.yajbe.YajbeEnumMapping.YajbeEnumLruMappingConfig;
import io.github.matteobertozzi.yajbe.YajbeFactory;
import io.github.matteobertozzi.yajbe.YajbeMapper;

public abstract class AbstractTestEncodeDecode {
  protected static final ObjectMapper YAJBE_ENUM_MAPPER = ExamplesUtil.newObjectMapper(new YajbeMapper(new YajbeFactory(new YajbeEnumLruMappingConfig(256, 5))));
  protected static final ObjectMapper YAJBE_MAPPER = ExamplesUtil.newObjectMapper(new YajbeMapper());
  protected static final ObjectMapper JSON_MAPPER = ExamplesUtil.newObjectMapper(new JsonMapper());
  protected static final ObjectMapper CBOR_MAPPER = ExamplesUtil.newObjectMapper(new CBORMapper());
  protected static final ObjectMapper XML_MAPPER = ExamplesUtil.newObjectMapper(new XmlMapper());

  public record TestData (String name, Class<?> inputType, Object input, byte[] jsonEnc, byte[] cborEnc, byte[] yajbeEnc, byte[] yajbeEnumEnc) {
    public TestData(final String name, final Class<?> inputType, final Object input) throws JsonProcessingException {
      this(
        name, inputType, input,
        JSON_MAPPER.writeValueAsBytes(input),
        CBOR_MAPPER.writeValueAsBytes(input),
        YAJBE_MAPPER.writeValueAsBytes(input),
        YAJBE_ENUM_MAPPER.writeValueAsBytes(input)
      );
    }
  }

  @FunctionalInterface
  protected interface TestDataConsumer {
    void accept(File file, JsonNode node) throws Exception;
  }

  protected static void foreachTestData(final File rootDir, final TestDataConsumer consumer) throws Exception {
    final File[] files = rootDir.listFiles();
    if (files == null || files.length == 0) return;

    for (final File file: files) {
      if (file.isDirectory()) {
        foreachTestData(file, consumer);
      } else if (file.getName().endsWith(".json")) {
        consumer.accept(file, JSON_MAPPER.readTree(file));
      } else if (file.getName().endsWith(".json.gz")) {
        consumer.accept(file, parseGzFile(JSON_MAPPER, file));
      } else if (file.getName().endsWith(".xml")) {
        consumer.accept(file, XML_MAPPER.readTree(file));
      } else if (file.getName().endsWith(".xml.gz")) {
        consumer.accept(file, parseGzFile(XML_MAPPER, file));
      }
    }
  }

  private static JsonNode parseGzFile(final ObjectMapper mapper, final File file) throws IOException {
    try (GZIPInputStream stream = new GZIPInputStream(new FileInputStream(file))) {
      return mapper.readTree(stream);
    }
  }

  protected static HumansTableView runEncodeDecode(final List<TestData> testData) throws Exception {
    final HumansTableView results = new HumansTableView();
    results.addColumn("file", null);
    results.addColumn("JSON Size", v -> ExamplesUtil.humanSize(((Number)v).longValue()));
    results.addColumn("CBOR Size", v -> ExamplesUtil.humanSize(((Number)v).longValue()));
    results.addColumn("Yajbe Size", v -> ExamplesUtil.humanSize(((Number)v).longValue()));
    results.addColumn("YajbeEnum Size", v -> ExamplesUtil.humanSize(((Number)v).longValue()));
    results.addColumn("JSON ops/sec", v -> ExamplesUtil.humanRate(((Number)v).longValue()));
    results.addColumn("CBOR ops/sec", v -> ExamplesUtil.humanRate(((Number)v).doubleValue()));
    results.addColumn("Yajbe ops/sec", v -> ExamplesUtil.humanRate(((Number)v).doubleValue()));
    results.addColumn("YajbeEnum ops/sec", v -> ExamplesUtil.humanRate(((Number)v).doubleValue()));
    results.addColumn("JSON time", v -> ExamplesUtil.humanTimeNanos(((Number)v).longValue()));
    results.addColumn("CBOR time", v -> ExamplesUtil.humanTimeNanos(((Number)v).longValue()));
    results.addColumn("Yajbe time", v -> ExamplesUtil.humanTimeNanos(((Number)v).longValue()));
    results.addColumn("YajbeEnum time", v -> ExamplesUtil.humanTimeNanos(((Number)v).longValue()));
    results.addColumn("Yajbe/JSON", v -> String.format("%.2f%%", (double)v * 100.0));
    results.addColumn("Yajbe/CBOR", v -> String.format("%.2f%%", (double)v * 100.0));
    results.addColumn("YajbeEnum/Yajbe", v -> String.format("%.2f%%", (double)v * 100.0));

    for (final TestData data: testData) {
      final int NRUNS = Math.max(10, (int)Math.ceil(2_000_000_000.0 / data.jsonEnc.length));

      final long jsonElapsed = ExamplesUtil.runEncodeDecodeBench(JSON_MAPPER, NRUNS, data.name(), data.input());
      final long cborElapsed = ExamplesUtil.runEncodeDecodeBench(CBOR_MAPPER, NRUNS, data.name(), data.input());
      final long yajbeElapsed = ExamplesUtil.runEncodeDecodeBench(YAJBE_MAPPER, NRUNS, data.name(), data.input());
      final long yajbeEnumElapsed = ExamplesUtil.runEncodeDecodeBench(YAJBE_ENUM_MAPPER, NRUNS, data.name(), data.input());

      if (yajbeElapsed < cborElapsed) {
        System.out.println(String.format(
          " ----> YAJBE FASTER THAN JSON:%.2f%% %s CBOR:%.2f%% %s",
          (1.0 - ((double)yajbeElapsed / jsonElapsed)) * 100,
          ExamplesUtil.humanTimeNanos(jsonElapsed - yajbeElapsed),
          (1.0 - ((double)yajbeElapsed / cborElapsed)) * 100,
          ExamplesUtil.humanTimeNanos(cborElapsed - yajbeElapsed)
        ));
      } else {
        System.out.println(String.format(
          " ----> CBOR FASTER THAN JSON:%.2f%% %s YABE:%.2f%% %s",
          (1.0 - ((double)cborElapsed / yajbeElapsed)) * 100,
          ExamplesUtil.humanTimeNanos(jsonElapsed - cborElapsed),
          (1.0 - ((double)cborElapsed / yajbeElapsed)) * 100,
          ExamplesUtil.humanTimeNanos(yajbeElapsed - cborElapsed)
        ));
      }

      results.addRow(data.name(),
        data.jsonEnc().length,
        data.cborEnc().length,
        data.yajbeEnc().length,
        data.yajbeEnumEnc().length,
        (double)NRUNS / (jsonElapsed / 1000000000.0),
        (double)NRUNS / (cborElapsed / 1000000000.0),
        (double)NRUNS / (yajbeElapsed / 1000000000.0),
        (double)NRUNS / (yajbeEnumElapsed / 1000000000.0),
        jsonElapsed,
        cborElapsed,
        yajbeElapsed,
        yajbeEnumElapsed,
        (1.0 - ((double)yajbeElapsed / jsonElapsed)),
        (1.0 - ((double)yajbeElapsed / cborElapsed)),
        (1.0 - ((double)yajbeEnumElapsed / yajbeElapsed))
      );
    }
    return results;
  }

  protected static HumansTableView runSeparateEncodeDecode(final List<TestData> testData) throws Exception {
    final HumansTableView results = new HumansTableView();
    results.addColumn("file", null);
    results.addColumn("JSON Size", v -> ExamplesUtil.humanSize(((Number)v).longValue()));
    results.addColumn("CBOR Size", v -> ExamplesUtil.humanSize(((Number)v).longValue()));
    results.addColumn("Yajbe Size", v -> ExamplesUtil.humanSize(((Number)v).longValue()));
    // encode
    results.addColumn("JSON Enc ops/sec", v -> ExamplesUtil.humanRate(((Number)v).longValue()));
    results.addColumn("CBOR Enc ops/sec", v -> ExamplesUtil.humanRate(((Number)v).doubleValue()));
    results.addColumn("Yajbe Enc ops/sec", v -> ExamplesUtil.humanRate(((Number)v).doubleValue()));
    results.addColumn("JSON Enc time", v -> ExamplesUtil.humanTimeNanos(((Number)v).longValue()));
    results.addColumn("CBOR Enc time", v -> ExamplesUtil.humanTimeNanos(((Number)v).longValue()));
    results.addColumn("Yajbe Enc time", v -> ExamplesUtil.humanTimeNanos(((Number)v).longValue()));
    results.addColumn("Yajbe/JSON Enc", v -> String.format("%.2f%%", (double)v * 100.0));
    results.addColumn("Yajbe/CBOR Enc", v -> String.format("%.2f%%", (double)v * 100.0));
    // decode
    results.addColumn("JSON Dec ops/sec", v -> ExamplesUtil.humanRate(((Number)v).longValue()));
    results.addColumn("CBOR Dec ops/sec", v -> ExamplesUtil.humanRate(((Number)v).doubleValue()));
    results.addColumn("Yajbe Dec ops/sec", v -> ExamplesUtil.humanRate(((Number)v).doubleValue()));
    results.addColumn("JSON Dec time", v -> ExamplesUtil.humanTimeNanos(((Number)v).longValue()));
    results.addColumn("CBOR Dec time", v -> ExamplesUtil.humanTimeNanos(((Number)v).longValue()));
    results.addColumn("Yajbe Dec time", v -> ExamplesUtil.humanTimeNanos(((Number)v).longValue()));
    results.addColumn("Yajbe/JSON Dec", v -> String.format("%.2f%%", (double)v * 100.0));
    results.addColumn("Yajbe/CBOR Dec", v -> String.format("%.2f%%", (double)v * 100.0));

    for (final TestData data: testData) {
      final long NRUNS = 10 + ((int)Math.ceil(2_000_000_000.0 / data.jsonEnc.length));

      final byte[] jsonEnc = JSON_MAPPER.writeValueAsBytes(data.input());
      final byte[] cborEnc = CBOR_MAPPER.writeValueAsBytes(data.input());
      final byte[] yajbeEnc = YAJBE_MAPPER.writeValueAsBytes(data.input());
      System.out.println(String.format("-> %s JSON:%s CBOR:%s YAJBE:%s",
        data.name(),
        ExamplesUtil.humanSize(jsonEnc.length),
        ExamplesUtil.humanSize(cborEnc.length),
        ExamplesUtil.humanSize(yajbeEnc.length)
      ));

      final long jsonEncodeElapsed = 1; //ExamplesUtil.runEncodeBench(JSON_MAPPER, NRUNS, data.name(), data.input());
      final long cborEncodeElapsed = ExamplesUtil.runEncodeBench(CBOR_MAPPER, NRUNS, data.name(), data.input());
      final long yajbeEncodeElapsed = ExamplesUtil.runEncodeBench(YAJBE_MAPPER, NRUNS, data.name(), data.input());

      if (yajbeEncodeElapsed < cborEncodeElapsed) {
        System.out.println(String.format(
          " ----> YAJBE ENCODER FASTER THAN JSON:%.2f%% %s CBOR:%.2f%% %s",
          (1.0 - ((double)yajbeEncodeElapsed / jsonEncodeElapsed)) * 100,
          ExamplesUtil.humanTimeNanos(jsonEncodeElapsed - yajbeEncodeElapsed),
          (1.0 - ((double)yajbeEncodeElapsed / cborEncodeElapsed)) * 100,
          ExamplesUtil.humanTimeNanos(cborEncodeElapsed - yajbeEncodeElapsed)
        ));
      } else {
        System.out.println(String.format(
          " ----> CBOR ENCODER FASTER THAN JSON:%.2f%% %s YABE:%.2f%% %s",
          (1.0 - ((double)cborEncodeElapsed / jsonEncodeElapsed)) * 100,
          ExamplesUtil.humanTimeNanos(jsonEncodeElapsed - cborEncodeElapsed),
          (1.0 - ((double)cborEncodeElapsed / yajbeEncodeElapsed)) * 100,
          ExamplesUtil.humanTimeNanos(yajbeEncodeElapsed - cborEncodeElapsed)
        ));
      }

      final long jsonDecodeElapsed = ExamplesUtil.runDecodeBench(JSON_MAPPER, NRUNS, data.name(), jsonEnc);
      final long cborDecodeElapsed = ExamplesUtil.runDecodeBench(CBOR_MAPPER, NRUNS, data.name(), cborEnc);
      final long yajbeDecodeElapsed = ExamplesUtil.runDecodeBench(YAJBE_MAPPER, NRUNS, data.name(), yajbeEnc);

      if (yajbeEncodeElapsed < cborEncodeElapsed) {
        System.out.println(String.format(
          " ----> YAJBE DECODER FASTER THAN JSON:%.2f%% %s CBOR:%.2f%% %s",
          (1.0 - ((double)yajbeDecodeElapsed / jsonDecodeElapsed)) * 100,
          ExamplesUtil.humanTimeNanos(jsonDecodeElapsed - yajbeDecodeElapsed),
          (1.0 - ((double)yajbeDecodeElapsed / cborDecodeElapsed)) * 100,
          ExamplesUtil.humanTimeNanos(cborDecodeElapsed - yajbeDecodeElapsed)
        ));
      } else {
        System.out.println(String.format(
          " ----> CBOR DECODER FASTER THAN JSON:%.2f%% %s YABE:%.2f%% %s",
          (1.0 - ((double)cborDecodeElapsed / jsonDecodeElapsed)) * 100,
          ExamplesUtil.humanTimeNanos(jsonDecodeElapsed - cborDecodeElapsed),
          (1.0 - ((double)cborDecodeElapsed / yajbeDecodeElapsed)) * 100,
          ExamplesUtil.humanTimeNanos(yajbeDecodeElapsed - cborDecodeElapsed)
        ));
      }

      results.addRow(data.name(),
        jsonEnc.length,
        cborEnc.length,
        yajbeEnc.length,
        (double)NRUNS / (jsonEncodeElapsed / 1000000000.0),
        (double)NRUNS / (cborEncodeElapsed / 1000000000.0),
        (double)NRUNS / (yajbeEncodeElapsed / 1000000000.0),
        jsonEncodeElapsed,
        cborEncodeElapsed,
        yajbeEncodeElapsed,
        (1.0 - ((double)yajbeEncodeElapsed / jsonEncodeElapsed)),
        (1.0 - ((double)yajbeEncodeElapsed / cborEncodeElapsed)),
        (double)NRUNS / (jsonDecodeElapsed / 1000000000.0),
        (double)NRUNS / (cborDecodeElapsed / 1000000000.0),
        (double)NRUNS / (yajbeDecodeElapsed / 1000000000.0),
        jsonDecodeElapsed,
        cborDecodeElapsed,
        yajbeDecodeElapsed,
        (1.0 - ((double)yajbeDecodeElapsed / jsonDecodeElapsed)),
        (1.0 - ((double)yajbeDecodeElapsed / cborDecodeElapsed))
      );
    }

    return results;
  }
}

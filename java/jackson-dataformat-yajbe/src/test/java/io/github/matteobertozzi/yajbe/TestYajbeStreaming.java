package io.github.matteobertozzi.yajbe;

import static org.junit.jupiter.api.Assertions.assertEquals;

import java.io.ByteArrayOutputStream;
import java.io.IOException;
import java.util.ArrayList;
import java.util.HexFormat;
import java.util.Map;

import org.junit.jupiter.api.Test;

import com.fasterxml.jackson.core.JsonGenerator;
import com.fasterxml.jackson.core.JsonParser;

public class TestYajbeStreaming extends BaseYajbeTest {
  @Test
  public void testSimple() throws IOException {
    final ArrayList<Map<String, Object>> rows = new ArrayList<>();
    rows.add(Map.of("aaa", 1)); // 3f836161614001
    rows.add(Map.of("bbb", 2)); // 3f836262624101
    rows.add(Map.of("aaa", 3)); // 3fa04201
    rows.add(Map.of("a", 4));   // 3f81614301
    rows.add(Map.of("b", 5));   // 3f81624401
    rows.add(Map.of("a", 6));   // 3fa24501
    rows.add(Map.of("b", 7));   // 3fa34601

    try (ByteArrayOutputStream wstream = new ByteArrayOutputStream()) {
      try (JsonGenerator generator = YAJBE_MAPPER.createGenerator(wstream)) {
        for (final Map<String, Object> row: rows) {
          generator.writeObject(row);
        }
      }
      assertEquals("3f8361616140013f8362626241013fa042013f816143013f816244013fa245013fa34601", HexFormat.of().formatHex(wstream.toByteArray()));

      try (JsonParser parser = YAJBE_MAPPER.createParser(wstream.toByteArray())) {
        for (final Map<String, Object> row: rows) {
          assertEquals(row, parser.readValueAs(Map.class));
        }
      }
    }
  }
}

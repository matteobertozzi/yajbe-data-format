# YAJBE for Java

There is a jackson-dataformat-yajbe library uploaded to the maven repo.

You should include in your pom.xml the yajbe-data-format repository.
```xml
<repositories>
  <repository>
    <id>central</id>
    <url>https://repo1.maven.org/maven2</url>
  </repository>
  <repository>
    <id>github</id>
    <url>https://maven.pkg.github.com/matteo.bertozzi/yajbe-data-format</url>
    <snapshots>
      <enabled>true</enabled>
    </snapshots>
  </repository>
</repositories>
```

and include the package as dependency
```xml
<dependencies>
  <dependency>
    <groupId>tech.dnaco</groupId>
    <artifactId>jackson-dataformat-yajbe</artifactId>
    <version>0.9.0-SNAPSHOT</version>
  </dependency>
</dependencies>
```

To used YAJBE you can just create an instance of YajbeMapper as you do for the JsonMapper. and then use with the writeValue() methods or the readValue() methods as you always do. the only difference is that the write output will be a byte-array and not a string.
```java
import com.fasterxml.jackson.databind.json.JsonMapper;

import tech.dnaco.yajbe.YajbeMapper;

public class MyTest {
  public record TestObj (int a, float b, String c) {}

  public static void main(String[] args) throws Exception {
    final JsonMapper json = new JsonMapper();
    final YajbeMapper yajbe = new YajbeMapper(); // the YAJBE mapper to be used for encode/decode

    // encode/decode using the JSON mapper
    final String j1 = json.writeValueAsString(Map.of("a", 10, "b", 20));
    System.out.println(j1); // { "a": 10, "b": 20 }
    System.out.println(json.readValue(j1, Map.class)); // {a=10, b=20}

    // encode/decode using the YAJBE mapper
    final byte[] y1 = yajbe.writeValueAsBytes(Map.of("a", 10, "b", 20));
    System.out.println(HexFormat.of().formatHex(y1)); // 3f81614981625301
    System.out.println(yajbe.readValue(y1, Map.class)); // {a=10, b=20}

    // encode decode a java record
    final byte[] y2 = yajbe.writeValueAsBytes(new TestObj(1, 5.23f, "test"));
    System.out.println(HexFormat.of().formatHex(y2)); // 3f816140816205295ca7408163c47465737401
    System.out.println(yajbe.readValue(y2, TestObj.class)); // TestObj[a=1, b=5.23, c=test]
  }
}
```
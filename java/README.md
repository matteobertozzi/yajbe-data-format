# YAJBE for Java

There is a jackson-dataformat-yajbe library uploaded to the maven repo.
[https://repo1.maven.org/maven2/io/github/matteobertozzi/jackson-dataformat-yajbe/](https://repo1.maven.org/maven2/io/github/matteobertozzi/jackson-dataformat-yajbe/)

so you'll just probably need to include the package as dependency in your pom.xml
```xml
<dependencies>
  <dependency>
    <groupId>io.github.matteobertozzi</groupId>
    <artifactId>jackson-dataformat-yajbe</artifactId>
    <version>2.17.0</version>
  </dependency>
</dependencies>
```

In case you are not able to download the jar, you should check your repositories and add them to your pom.xml if needed.
```xml
<repositories>
  <repository>
    <id>central</id>
    <url>https://repo1.maven.org/maven2</url>
  </repository>
  <repository>
    <id>oss.sonatype</id>
    <url>https://s01.oss.sonatype.org/content/repositories/releases</url>
  </repository>
  <repository>
    <id>oss.sonatype.snapshots</id>
    <url>https://s01.oss.sonatype.org/content/repositories/snapshots</url>
    <snapshots>
      <enabled>true</enabled>
    </snapshots>
  </repository>
</repositories>
```

To used YAJBE you can just create an instance of YajbeMapper as you do for the JsonMapper. and then use with the writeValue() methods or the readValue() methods as you always do. the only difference is that the write output will be a byte-array and not a string.
```java
import com.fasterxml.jackson.databind.json.JsonMapper;

import io.github.matteobertozzi.yajbe.YajbeMapper;

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
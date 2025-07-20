# YAJBE Go Implementation

A Go implementation of the YAJBE (Yet Another JSON Binary Encoding) format.

## Features

- **Drop-in replacement for JSON**: Use `yajbe.Marshal()` and `yajbe.Unmarshal()` just like `json.Marshal()` and `json.Unmarshal()`
- **Binary encoding**: Compact binary format for efficient storage and transmission
- **Type support**: Supports all Go primitive types, structs, slices, maps, and big integers
- **Streaming**: Low-level streaming API with `Writer` and `Reader` for performance-critical applications

## Usage

### Basic Marshal/Unmarshal

```go
package main

import (
    "fmt"
    "github.com/matteobertozzi/yajbe-data-format/golang/yajbe"
)

type Person struct {
    Name string `json:"name"`
    Age  int    `json:"age"`
    City string `json:"city"`
}

func main() {
    person := Person{
        Name: "Alice",
        Age:  30,
        City: "New York",
    }

    // Marshal to YAJBE binary format
    data, err := yajbe.Marshal(person)
    if err != nil {
        panic(err)
    }

    // Unmarshal from YAJBE binary format
    var result Person
    err = yajbe.Unmarshal(data, &result)
    if err != nil {
        panic(err)
    }

    fmt.Printf("%+v\n", result)
}
```

### Streaming API

```go
package main

import (
    "bytes"
    "github.com/matteobertozzi/yajbe-data-format/golang/yajbe"
)

func main() {
    buf := &bytes.Buffer{}
    w := yajbe.NewWriter(buf)

    // Write values directly
    w.WriteString("hello")
    w.WriteInt(42)
    w.WriteBool(true)
    w.Flush()

    // Read values back
    r := yajbe.NewReader(bytes.NewReader(buf.Bytes()))

    str, _ := r.ReadValue()  // "hello"
    num, _ := r.ReadValue()  // 42
    flag, _ := r.ReadValue() // true
}
```


## JSON Compatibility

The implementation respects JSON struct tags:

```go
type User struct {
    ID       int    `json:"id"`
    Username string `json:"username"`
    Email    string `json:"email,omitempty"`
    Internal string `json:"-"` // ignored
}
```

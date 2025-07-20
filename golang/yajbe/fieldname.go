// Licensed to the Apache Software Foundation (ASF) under one or more
// contributor license agreements.  See the NOTICE file distributed with
// this work for additional information regarding copyright ownership.
// The ASF licenses this file to You under the Apache License, Version 2.0
// (the "License"); you may not use this file except in compliance with
// the License.  You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package yajbe

// FieldNameWriter handles writing of field names with compression
type FieldNameWriter struct {
	writer     RawWriter
	indexedMap map[string]int
	lastKey    []byte
}

// RawWriter interface for writing raw bytes
type RawWriter interface {
	WriteRaw([]byte) error
}

// BufferWriter implements RawWriter
func (w *BufferWriter) WriteRaw(data []byte) error {
	w.buf = append(w.buf, data...)
	return nil
}

// StreamWriter implements RawWriter
func (w *StreamWriter) WriteRaw(data []byte) error {
	_, err := w.writer.Write(data)
	return err
}

// NewFieldNameWriter creates a new field name writer
func NewFieldNameWriter(writer RawWriter) *FieldNameWriter {
	return &FieldNameWriter{
		writer:     writer,
		indexedMap: make(map[string]int, 32), // Pre-size for common usage
		lastKey:    make([]byte, 0, 64),      // Pre-size for typical field names
	}
}

// Write writes a field name using compression similar to Dart implementation
func (w *FieldNameWriter) Write(name string) error {
	// Fast path for short common field names - no compression needed
	if len(name) <= 4 {
		// Check if we have an indexed version for short names
		if index, exists := w.indexedMap[name]; exists {
			w.writeIndexedFieldName(index)
			utf8data := stringToBytes(name)
			w.lastKey = utf8data
			return nil
		}
		// Write short names directly
		utf8data := stringToBytes(name)
		w.writeFullFieldName(utf8data)
		// Add to index if we have space
		if len(w.indexedMap) < 65819 {
			w.indexedMap[name] = len(w.indexedMap)
		}
		w.lastKey = utf8data
		return nil
	}

	// Use zero-copy string to bytes conversion
	utf8data := stringToBytes(name)

	// Check if we have an indexed version
	if index, exists := w.indexedMap[name]; exists {
		w.writeIndexedFieldName(index)
		w.lastKey = utf8data
		return nil
	}

	// Compression logic for longer keys
	if len(w.lastKey) > 0 {
		prefix := minInt(0xff, w.prefix(utf8data))
		suffix := w.suffix(utf8data, prefix)

		if suffix > 2 {
			w.writePrefixSuffix(utf8data, prefix, minInt(0xff, suffix))
		} else if prefix > 2 {
			w.writePrefix(utf8data, prefix)
		} else {
			w.writeFullFieldName(utf8data)
		}
	} else {
		w.writeFullFieldName(utf8data)
	}

	// Add to index if we have space
	if len(w.indexedMap) < 65819 {
		w.indexedMap[name] = len(w.indexedMap)
	}
	w.lastKey = utf8data
	return nil
}

func (w *FieldNameWriter) writeFullFieldName(fieldName []byte) error {
	// 100----- Full Field Name (0-29 length - 1, 30 1b-len, 31 2b-len)
	// Inline simple cases to avoid function call overhead
	length := len(fieldName)
	if length < 30 {
		// Write header byte directly
		if err := w.writer.WriteRaw([]byte{0x80 | byte(length)}); err != nil {
			return err
		}
		// Write field name directly - no temp buffer needed
		return w.writer.WriteRaw(fieldName)
	}
	return w.writeFieldLength(0x80, length, fieldName)
}

func (w *FieldNameWriter) writeIndexedFieldName(fieldIndex int) error {
	// 101----- Field Offset (0-29 field, 30 1b-len, 31 2b-len)
	return w.writeFieldLength(0xa0, fieldIndex, nil)
}

func (w *FieldNameWriter) writePrefix(fieldName []byte, prefix int) error {
	// 110----- Prefix (1byte prefix, 0-29 length - 1, 30 1b-len, 31 2b-len)
	length := len(fieldName) - prefix
	// Optimize for common case where length < 30
	if length < 30 {
		// Write header and prefix bytes directly - no temp buffer
		header := [2]byte{0xc0 | byte(length), byte(prefix)}
		if err := w.writer.WriteRaw(header[:]); err != nil {
			return err
		}
		// Write remaining field name data directly
		return w.writer.WriteRaw(fieldName[prefix:])
	}

	header, err := w.buildFieldHeader(0xc0, length)
	if err != nil {
		return err
	}

	// Write separately to avoid allocation
	if err := w.writer.WriteRaw(header); err != nil {
		return err
	}
	prefixByte := [1]byte{byte(prefix)}
	if err := w.writer.WriteRaw(prefixByte[:]); err != nil {
		return err
	}
	return w.writer.WriteRaw(fieldName[prefix:])
}

func (w *FieldNameWriter) writePrefixSuffix(fieldName []byte, prefix, suffix int) error {
	// 111----- Prefix/Suffix (1byte prefix, 1byte suffix, 0-29 length - 1, 30 1b-len, 31 2b-len)
	length := len(fieldName) - prefix - suffix
	// Optimize for common case where length < 30
	if length < 30 {
		// Write header, prefix, and suffix bytes directly - no temp buffer
		header := [3]byte{0xe0 | byte(length), byte(prefix), byte(suffix)}
		if err := w.writer.WriteRaw(header[:]); err != nil {
			return err
		}
		// Write middle part of field name directly
		return w.writer.WriteRaw(fieldName[prefix : len(fieldName)-suffix])
	}

	header, err := w.buildFieldHeader(0xe0, length)
	if err != nil {
		return err
	}

	// Write separately to avoid allocation
	if err := w.writer.WriteRaw(header); err != nil {
		return err
	}
	prefixSuffixBytes := [2]byte{byte(prefix), byte(suffix)}
	if err := w.writer.WriteRaw(prefixSuffixBytes[:]); err != nil {
		return err
	}
	return w.writer.WriteRaw(fieldName[prefix : len(fieldName)-suffix])
}

func (w *FieldNameWriter) writeFieldLength(head byte, length int, data []byte) error {
	header, err := w.buildFieldHeader(head, length)
	if err != nil {
		return err
	}

	if data != nil {
		// Write header and data separately to avoid allocation
		if err := w.writer.WriteRaw(header); err != nil {
			return err
		}
		return w.writer.WriteRaw(data)
	}

	return w.writer.WriteRaw(header)
}

func (w *FieldNameWriter) buildFieldHeader(head byte, length int) ([]byte, error) {
	if length < 30 {
		return []byte{head | byte(length)}, nil
	} else if length <= 284 {
		return []byte{head | 0x1e, byte((length - 29) & 0xff)}, nil
	} else if length <= 65819 {
		delta := length - 284
		return []byte{head | 0x1f, byte(delta / 256), byte(delta & 255)}, nil
	}
	return nil, ErrInvalidFormat
}

func (w *FieldNameWriter) prefix(key []byte) int {
	a := w.lastKey
	b := key
	minLen := minInt(len(a), len(b))
	for i := 0; i < minLen; i++ {
		if a[i] != b[i] {
			return i
		}
	}
	return minLen
}

func (w *FieldNameWriter) suffix(key []byte, keyPrefix int) int {
	a := w.lastKey
	b := key
	bLen := len(b) - keyPrefix
	minLen := minInt(len(a), bLen)
	for i := 1; i <= minLen; i++ {
		if a[len(a)-i] != b[keyPrefix+(bLen-i)] {
			return i - 1
		}
	}
	return minLen
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// FieldNameReader handles reading of field names with decompression
type FieldNameReader struct {
	reader       Reader
	indexedNames []string
	lastKey      []byte
}

// NewFieldNameReader creates a new field name reader
func NewFieldNameReader(reader Reader) *FieldNameReader {
	return &FieldNameReader{
		reader:       reader,
		indexedNames: make([]string, 0, 32),
		lastKey:      make([]byte, 0),
	}
}

// Read reads and decompresses a field name
func (r *FieldNameReader) Read() (string, error) {
	head, err := r.reader.ReadByte()
	if err != nil {
		return "", err
	}

	switch head & 0xe0 {
	case 0x80: // Full field name
		return r.readFullFieldName(head)
	case 0xa0: // Indexed field name
		return r.readIndexedFieldName(head)
	case 0xc0: // Prefix field name
		return r.readPrefixFieldName(head)
	case 0xe0: // Prefix/Suffix field name
		return r.readPrefixSuffixFieldName(head)
	default:
		return "", ErrInvalidFormat
	}
}

func (r *FieldNameReader) readFullFieldName(head byte) (string, error) {
	length, err := r.readFieldLength(head)
	if err != nil {
		return "", err
	}

	data, err := r.reader.ReadBytes(length)
	if err != nil {
		return "", err
	}

	name := string(data)
	r.addToIndex(name)
	r.lastKey = data
	return name, nil
}

func (r *FieldNameReader) readIndexedFieldName(head byte) (string, error) {
	index, err := r.readFieldLength(head)
	if err != nil {
		return "", err
	}

	if index >= len(r.indexedNames) {
		return "", ErrInvalidFormat
	}

	name := r.indexedNames[index]
	// Use zero-copy string to bytes conversion
	r.lastKey = stringToBytes(name)
	return name, nil
}

func (r *FieldNameReader) readPrefixFieldName(head byte) (string, error) {
	length, err := r.readFieldLength(head)
	if err != nil {
		return "", err
	}

	prefixByte, err := r.reader.ReadByte()
	if err != nil {
		return "", err
	}
	prefix := int(prefixByte)

	data, err := r.reader.ReadBytes(length)
	if err != nil {
		return "", err
	}

	// Reconstruct full field name
	fullName := make([]byte, prefix+length)
	copy(fullName[:prefix], r.lastKey[:prefix])
	copy(fullName[prefix:], data)

	name := string(fullName)
	r.addToIndex(name)
	r.lastKey = fullName
	return name, nil
}

func (r *FieldNameReader) readPrefixSuffixFieldName(head byte) (string, error) {
	length, err := r.readFieldLength(head)
	if err != nil {
		return "", err
	}

	prefixByte, err := r.reader.ReadByte()
	if err != nil {
		return "", err
	}
	prefix := int(prefixByte)

	suffixByte, err := r.reader.ReadByte()
	if err != nil {
		return "", err
	}
	suffix := int(suffixByte)

	data, err := r.reader.ReadBytes(length)
	if err != nil {
		return "", err
	}

	// Reconstruct full field name
	fullName := make([]byte, prefix+length+suffix)
	copy(fullName[:prefix], r.lastKey[:prefix])
	copy(fullName[prefix:prefix+length], data)
	lastKeyLen := len(r.lastKey)
	copy(fullName[prefix+length:], r.lastKey[lastKeyLen-suffix:])

	name := string(fullName)
	r.addToIndex(name)
	r.lastKey = fullName
	return name, nil
}

func (r *FieldNameReader) readFieldLength(head byte) (int, error) {
	length := int(head & 0x1f)
	if length < 30 {
		return length, nil
	} else if length == 30 {
		b, err := r.reader.ReadByte()
		if err != nil {
			return 0, err
		}
		return 29 + int(b), nil
	} else { // length == 31
		b1, err := r.reader.ReadByte()
		if err != nil {
			return 0, err
		}
		b2, err := r.reader.ReadByte()
		if err != nil {
			return 0, err
		}
		return 284 + int(b1)*256 + int(b2), nil
	}
}

func (r *FieldNameReader) addToIndex(name string) {
	if len(r.indexedNames) < 65819 {
		r.indexedNames = append(r.indexedNames, name)
	}
}

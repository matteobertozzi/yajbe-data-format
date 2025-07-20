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

// Package yajbe provides encoding and decoding of YAJBE (Yet Another JSON Binary Encoding) format.
// YAJBE is a compact binary data format built to be a drop-in replacement for JSON.
package yajbe

import (
	"errors"
	"fmt"
	"io"
	"reflect"
	"sort"
	"strings"
)

var (
	ErrInvalidFormat  = errors.New("invalid YAJBE format")
	ErrBufferTooSmall = errors.New("buffer too small")
)

// Marshal returns the YAJBE encoding of v.
func Marshal(v any) ([]byte, error) {
	// Use pooled buffer for writer
	buf := getWriterBuffer()
	writer := NewWriterFromBuffer(buf)
	encoder := getEncoder(writer)
	defer func() {
		putEncoder(encoder)
		// Don't put the buffer back since we're returning it to the caller
	}()

	if err := encoder.Encode(v); err != nil {
		return nil, err
	}

	// Make a copy of the result since we can't return the pooled buffer
	result := make([]byte, len(writer.Bytes()))
	copy(result, writer.Bytes())
	putWriterBuffer(buf)
	return result, nil
}

// Unmarshal parses the YAJBE-encoded data and stores the result
// in the value pointed to by v.
func Unmarshal(data []byte, v any) error {
	reader := NewReaderFromBytes(data)
	decoder := getDecoder(reader)
	defer putDecoder(decoder)

	return decoder.Decode(v)
}

// MarshalToWriter writes the YAJBE encoding of v to the provided writer.
func MarshalToWriter(w io.Writer, v any) error {
	// Use Marshal and then write to avoid StreamWriter bugs
	data, err := Marshal(v)
	if err != nil {
		return err
	}
	_, err = w.Write(data)
	return err
}

// UnmarshalFromReader reads YAJBE-encoded data from the reader and stores
// the result in the value pointed to by v.
func UnmarshalFromReader(r io.Reader, v any) error {
	reader := NewReaderFromReader(r)
	decoder := &Decoder{reader: reader}
	return decoder.Decode(v)
}

// Encoder writes YAJBE values to an output stream.
type Encoder struct {
	writer Writer
}

// NewEncoder creates a new YAJBE encoder that writes to the provided Writer.
func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{writer: NewWriterFromWriter(w)}
}

// Encode writes the YAJBE encoding of v to the stream.
func (e *Encoder) Encode(v any) error {
	return e.encodeValue(v)
}

// Decoder reads and decodes YAJBE values from an input stream.
type Decoder struct {
	reader Reader
}

// NewDecoder creates a new YAJBE decoder that reads from the provided Reader.
func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{reader: NewReaderFromReader(r)}
}

// Decode reads the next YAJBE-encoded value from its input and stores it in v.
func (d *Decoder) Decode(v any) error {
	return d.decodeValue(v)
}

func (e *Encoder) encodeValue(v any) error {
	if v == nil {
		return e.writer.WriteNull()
	}

	// Optimize common cases first to reduce type switch overhead
	switch val := v.(type) {
	case string:
		return e.writer.WriteString(val)
	case int64:
		return e.writer.WriteInt(val)
	case float64:
		return e.writer.WriteFloat64(val)
	case bool:
		return e.writer.WriteBool(val)
	case map[string]interface{}:
		return e.encodeMap(val)
	case []interface{}:
		return e.encodeSlice(val)
	case []byte:
		if val == nil {
			return e.writer.WriteNull()
		}
		return e.writer.WriteBytes(val)
	case []int:
		return encodeNullableSlice(e, val, e.encodeIntSlice)
	case map[string]int:
		return encodeNullableMap(e, val, e.encodeStringIntMap)
	case map[string]string:
		return encodeNullableMap(e, val, e.encodeStringStringMap)
	case float32:
		return e.writer.WriteFloat32(val)
	case int:
		converted, err := convertToInt64(val)
		if err != nil {
			return err
		}
		return e.writer.WriteInt(converted)
	case int32:
		converted, err := convertToInt64(val)
		if err != nil {
			return err
		}
		return e.writer.WriteInt(converted)
	case uint64:
		converted, err := handleUint64(val)
		if err != nil {
			return err
		}
		return e.writer.WriteInt(converted)
	case []string:
		return e.encodeStringSlice(val)
	case []int64:
		return e.encodeInt64Slice(val)
	case []bool:
		return e.encodeBoolSlice(val)
	case int8:
		converted, err := convertToInt64(val)
		if err != nil {
			return err
		}
		return e.writer.WriteInt(converted)
	case int16:
		converted, err := convertToInt64(val)
		if err != nil {
			return err
		}
		return e.writer.WriteInt(converted)
	case uint:
		converted, err := convertToInt64(val)
		if err != nil {
			return err
		}
		return e.writer.WriteInt(converted)
	case uint8:
		converted, err := convertToInt64(val)
		if err != nil {
			return err
		}
		return e.writer.WriteInt(converted)
	case uint16:
		converted, err := convertToInt64(val)
		if err != nil {
			return err
		}
		return e.writer.WriteInt(converted)
	case uint32:
		converted, err := convertToInt64(val)
		if err != nil {
			return err
		}
		return e.writer.WriteInt(converted)
	default:
		// Use reflection for structs to avoid JSON round-trip
		if rv := reflect.ValueOf(v); rv.Kind() == reflect.Struct {
			return e.encodeStruct(rv)
		}
		// Handle other types with direct reflection instead of JSON fallback
		return e.encodeReflectValue(reflect.ValueOf(v))
	}
}

// Specialized encoding functions to reduce function call overhead
func (e *Encoder) encodeMap(val map[string]any) error {
	if err := e.writer.WriteObjectStart(len(val)); err != nil {
		return err
	}

	// Sort keys for deterministic encoding
	keys := make([]string, 0, len(val))
	for key := range val {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	// Encode in sorted key order
	for _, key := range keys {
		if err := e.writer.WriteFieldName(key); err != nil {
			return err
		}
		if err := e.encodeValue(val[key]); err != nil {
			return err
		}
	}
	return e.writer.WriteObjectEnd()
}

func (e *Encoder) encodeSlice(val []any) error {
	if err := e.writer.WriteArrayStart(len(val)); err != nil {
		return err
	}
	for _, item := range val {
		if err := e.encodeValue(item); err != nil {
			return err
		}
	}
	return e.writer.WriteArrayEnd()
}

func (e *Encoder) encodeStringSlice(val []string) error {
	return encodeSliceGeneric(e, val, e.writer.WriteString)
}

func (e *Encoder) encodeInt64Slice(val []int64) error {
	return encodeSliceGeneric(e, val, e.writer.WriteInt)
}

func (e *Encoder) encodeBoolSlice(val []bool) error {
	return encodeSliceGeneric(e, val, e.writer.WriteBool)
}

func (e *Encoder) encodeIntSlice(val []int) error {
	return encodeSliceGeneric(e, val, func(item int) error {
		return e.writer.WriteInt(int64(item))
	})
}

func (e *Encoder) encodeStringIntMap(val map[string]int) error {
	return encodeMapGeneric(e, val, func(item int) error {
		return e.writer.WriteInt(int64(item))
	})
}

func (e *Encoder) encodeStringStringMap(val map[string]string) error {
	return encodeMapGeneric(e, val, e.writer.WriteString)
}

// encodeStruct efficiently encodes structs using reflection without JSON roundtrip
func (e *Encoder) encodeStruct(rv reflect.Value) error {
	rt := rv.Type()
	numFields := rt.NumField()

	// Count visible fields
	visibleFields := 0
	for i := range numFields {
		field := rt.Field(i)
		if field.PkgPath != "" && !field.Anonymous { // Skip unexported fields
			continue
		}
		visibleFields++
	}

	if err := e.writer.WriteObjectStart(visibleFields); err != nil {
		return err
	}

	for i := range numFields {
		field := rt.Field(i)
		fieldValue := rv.Field(i)

		if field.PkgPath != "" && !field.Anonymous { // Skip unexported fields
			continue
		}

		// Get field name from json tag or use field name
		fieldName := field.Name
		if tag := field.Tag.Get("json"); tag != "" && tag != "-" {
			if idx := strings.Index(tag, ","); idx >= 0 {
				fieldName = tag[:idx]
			} else {
				fieldName = tag
			}
		}

		if err := e.writer.WriteFieldName(fieldName); err != nil {
			return err
		}

		if err := e.encodeReflectValue(fieldValue); err != nil {
			return err
		}
	}

	return e.writer.WriteObjectEnd()
}

// encodeReflectValue efficiently encodes values using reflection
func (e *Encoder) encodeReflectValue(rv reflect.Value) error {
	if !rv.IsValid() {
		return e.writer.WriteNull()
	}

	switch rv.Kind() {
	case reflect.Ptr, reflect.Interface:
		if rv.IsNil() {
			return e.writer.WriteNull()
		}
		return e.encodeReflectValue(rv.Elem())

	case reflect.Bool:
		return e.writer.WriteBool(rv.Bool())

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return e.writer.WriteInt(rv.Int())

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		uval := rv.Uint()
		if uval > 9223372036854775807 {
			return valueTooLargeError("uint64")
		}
		return e.writer.WriteInt(int64(uval))

	case reflect.Float32:
		return e.writer.WriteFloat32(float32(rv.Float()))

	case reflect.Float64:
		return e.writer.WriteFloat64(rv.Float())

	case reflect.String:
		return e.writer.WriteString(rv.String())

	case reflect.Slice:
		if rv.Type().Elem().Kind() == reflect.Uint8 {
			// []byte case
			return e.writer.WriteBytes(rv.Bytes())
		}
		fallthrough

	case reflect.Array:
		n := rv.Len()
		if err := e.writer.WriteArrayStart(n); err != nil {
			return err
		}
		for i := range n {
			if err := e.encodeReflectValue(rv.Index(i)); err != nil {
				return err
			}
		}
		return e.writer.WriteArrayEnd()

	case reflect.Map:
		keys := rv.MapKeys()
		if err := e.writer.WriteObjectStart(len(keys)); err != nil {
			return err
		}
		for _, key := range keys {
			keyStr := fmt.Sprintf("%v", key.Interface())
			if err := e.writer.WriteFieldName(keyStr); err != nil {
				return err
			}
			if err := e.encodeReflectValue(rv.MapIndex(key)); err != nil {
				return err
			}
		}
		return e.writer.WriteObjectEnd()

	case reflect.Struct:
		return e.encodeStruct(rv)

	default:
		// Unsupported type - return error instead of JSON fallback
		return fmt.Errorf("unsupported type for YAJBE encoding: %v", rv.Type())
	}
}

func (d *Decoder) decodeValue(v any) error {
	token, err := d.reader.NextToken()
	if err != nil {
		return err
	}
	return d.decodeToken(token, v)
}

func (d *Decoder) decodeToken(token Token, v any) error {
	switch token.Type {
	case TokenNull:
		return d.setNull(v)
	case TokenBool:
		return d.setBool(v, token.BoolValue)
	case TokenInt:
		return d.setInt(v, token.IntValue)
	case TokenFloat32:
		return d.setFloat32(v, token.Float32Value)
	case TokenFloat64:
		return d.setFloat64(v, token.Float64Value)
	case TokenString:
		return d.setString(v, token.StringValue)
	case TokenBytes:
		return d.setBytes(v, token.BytesValue)
	case TokenArrayStart:
		return d.decodeArray(v, token.Length)
	case TokenObjectStart:
		return d.decodeObject(v, token.Length)
	default:
		return ErrInvalidFormat
	}
}

func (d *Decoder) setNull(v any) error {
	// Handle null value assignment based on the type of v
	switch ptr := v.(type) {
	case *[]int:
		*ptr = nil
	case *[]int64:
		*ptr = nil
	case *[]string:
		*ptr = nil
	case *[]bool:
		*ptr = nil
	case *[]interface{}:
		*ptr = nil
	case *map[string]int:
		*ptr = nil
	case *map[string]interface{}:
		*ptr = nil
	case *interface{}:
		*ptr = nil
	case **string:
		*ptr = nil
	case **int:
		*ptr = nil
	case **int64:
		*ptr = nil
	case **bool:
		*ptr = nil
	case **float32:
		*ptr = nil
	case **float64:
		*ptr = nil
	default:
		// For other types, we don't set them to avoid breaking non-pointer types
	}
	return nil
}

func (d *Decoder) setBool(v interface{}, value bool) error {
	switch ptr := v.(type) {
	case *bool:
		*ptr = value
	case *interface{}:
		*ptr = value
	default:
		return boolDecodeError(v)
	}
	return nil
}

func (d *Decoder) setInt(v interface{}, value int64) error {
	switch ptr := v.(type) {
	case *int:
		*ptr = int(value)
	case *int8:
		*ptr = int8(value)
	case *int16:
		*ptr = int16(value)
	case *int32:
		*ptr = int32(value)
	case *int64:
		*ptr = value
	case *uint:
		*ptr = uint(value)
	case *uint8:
		*ptr = uint8(value)
	case *uint16:
		*ptr = uint16(value)
	case *uint32:
		*ptr = uint32(value)
	case *uint64:
		*ptr = uint64(value)
	case *interface{}:
		// For JSON compatibility, decode integers as float64 when target is interface{}
		*ptr = float64(value)
	default:
		return intDecodeError(v)
	}
	return nil
}

func (d *Decoder) setFloat32(v interface{}, value float32) error {
	switch ptr := v.(type) {
	case *float32:
		*ptr = value
	case *float64:
		*ptr = float64(value)
	case *interface{}:
		// For JSON compatibility, convert float32 to float64 when unmarshaling to interface{}
		*ptr = float64(value)
	default:
		return float32DecodeError(v)
	}
	return nil
}

func (d *Decoder) setFloat64(v interface{}, value float64) error {
	switch ptr := v.(type) {
	case *float32:
		*ptr = float32(value)
	case *float64:
		*ptr = value
	case *interface{}:
		*ptr = value
	default:
		return float64DecodeError(v)
	}
	return nil
}

func (d *Decoder) setString(v interface{}, value string) error {
	switch ptr := v.(type) {
	case *string:
		*ptr = value
	case *interface{}:
		*ptr = value
	default:
		return stringDecodeError(v)
	}
	return nil
}

func (d *Decoder) setBytes(v interface{}, value []byte) error {
	switch ptr := v.(type) {
	case *[]byte:
		*ptr = value
	case *interface{}:
		*ptr = value
	default:
		return bytesDecodeError(v)
	}
	return nil
}

func (d *Decoder) decodeEOFArray(v interface{}) error {
	switch ptr := v.(type) {
	case *[]interface{}:
		arr := make([]interface{}, 0) // Initialize to empty slice, not nil
		for {
			token, err := d.reader.NextToken()
			if err == io.EOF {
				break // End of array
			}
			if err != nil {
				return err
			}

			var element interface{}
			if err := d.decodeToken(token, &element); err != nil {
				return err
			}
			arr = append(arr, element)
		}
		*ptr = arr
	case *interface{}:
		arr := make([]interface{}, 0) // Initialize to empty slice, not nil
		for {
			token, err := d.reader.NextToken()
			if err == io.EOF {
				break // End of array
			}
			if err != nil {
				return err
			}

			var element interface{}
			if err := d.decodeToken(token, &element); err != nil {
				return err
			}
			arr = append(arr, element)
		}
		*ptr = arr
	default:
		return eofArrayDecodeError(v)
	}
	return nil
}

func (d *Decoder) decodeArray(v interface{}, length int) error {
	if length < -1 {
		return arrayLengthError(length)
	}

	// Handle EOF arrays (length == -1)
	if length == -1 {
		return d.decodeEOFArray(v)
	}

	switch ptr := v.(type) {
	case *[]interface{}:
		// Use pooled slice for better performance
		arr := getInterfaceSlice(length)
		for i := range length {
			if err := d.decodeValue(&arr[i]); err != nil {
				return err
			}
		}
		*ptr = arr
		// Note: We don't put the slice back to pool since it's now owned by the caller
	case *[]bool:
		// Use pooled slice for better performance
		arr := getBoolSlice(length)
		for i := range length {
			if err := d.decodeValue(&arr[i]); err != nil {
				return err
			}
		}
		*ptr = arr
		// Note: We don't put the slice back to pool since it's now owned by the caller
	case *[]int64:
		// Use pooled slice for better performance
		arr := getInt64Slice(length)
		for i := range length {
			if err := d.decodeValue(&arr[i]); err != nil {
				return err
			}
		}
		*ptr = arr
		// Note: We don't put the slice back to pool since it's now owned by the caller
	case *[]string:
		// Use pooled slice for better performance
		arr := getStringSlice(length)
		for i := range length {
			if err := d.decodeValue(&arr[i]); err != nil {
				return err
			}
		}
		*ptr = arr
		// Note: We don't put the slice back to pool since it's now owned by the caller
	case *[]int:
		arr := make([]int, length)
		for i := range length {
			if err := d.decodeValue(&arr[i]); err != nil {
				return err
			}
		}
		*ptr = arr
	case *interface{}:
		// Use pooled slice for better performance
		arr := getInterfaceSlice(length)
		for i := range length {
			if err := d.decodeValue(&arr[i]); err != nil {
				return err
			}
		}
		*ptr = arr
		// Note: We don't put the slice back to pool since it's now owned by the caller
	default:
		// Handle slice of structs using reflection
		rv := reflect.ValueOf(v)
		if rv.Kind() != reflect.Ptr {
			return arrayDecodeError(v)
		}
		rv = rv.Elem()
		if rv.Kind() != reflect.Slice {
			return arrayDecodeError(v)
		}
		return d.decodeSlice(rv, length)
	}
	return nil
}

func (d *Decoder) decodeObject(v interface{}, length int) error {
	switch ptr := v.(type) {
	case *map[string]interface{}:
		// Use pooled map for better performance and generic helper
		obj, err := decodeObjectGeneric(d, length,
			getStringInterfaceMap,
			func(v *interface{}) error { return d.decodeValue(v) },
		)
		if err != nil {
			return err
		}
		*ptr = obj
	case *map[string]int:
		obj, err := decodeObjectGeneric(d, length,
			func(length int) map[string]int { return make(map[string]int, length) },
			func(v *int) error { return d.decodeValue(v) },
		)
		if err != nil {
			return err
		}
		*ptr = obj
	case *map[string]string:
		obj, err := decodeObjectGeneric(d, length,
			func(length int) map[string]string { return make(map[string]string, length) },
			func(v *string) error { return d.decodeValue(v) },
		)
		if err != nil {
			return err
		}
		*ptr = obj
	case *interface{}:
		// Use pooled map for better performance and generic helper
		obj, err := decodeObjectGeneric(d, length,
			getStringInterfaceMap,
			func(v *interface{}) error { return d.decodeValue(v) },
		)
		if err != nil {
			return err
		}
		*ptr = obj
	default:
		// Handle struct pointers using reflection
		rv := reflect.ValueOf(v)
		if rv.Kind() != reflect.Ptr || rv.Elem().Kind() != reflect.Struct {
			return objectDecodeError(v)
		}
		return d.decodeStruct(rv.Elem(), length)
	}
	return nil
}

func getTypeName(v any) string {
	if v == nil {
		return "nil"
	}
	return fmt.Sprintf("%T", v)
}

// decodeStruct decodes an object into a struct using reflection
func (d *Decoder) decodeStruct(rv reflect.Value, length int) error {
	rt := rv.Type()

	// Create a map to store field mappings for efficient lookup
	fieldMap := make(map[string]reflect.Value)
	for i := 0; i < rt.NumField(); i++ {
		field := rt.Field(i)
		fieldValue := rv.Field(i)

		// Skip unexported fields
		if field.PkgPath != "" && !field.Anonymous {
			continue
		}

		// Only process settable fields
		if !fieldValue.CanSet() {
			continue
		}

		// Get field name from json tag or use field name
		fieldName := field.Name
		if tag := field.Tag.Get("json"); tag != "" && tag != "-" {
			if idx := strings.Index(tag, ","); idx >= 0 {
				fieldName = tag[:idx]
			} else {
				fieldName = tag
			}
		}

		fieldMap[fieldName] = fieldValue
	}

	if length == -1 {
		// EOF object - read until we hit end marker (0x01)
		for {
			// Check if next byte is end marker
			nextByte, err := d.reader.Peek()
			if err != nil {
				return err
			}
			if nextByte == 0x01 {
				// Consume the end marker
				_, err := d.reader.ReadByte()
				if err != nil {
					return err
				}
				break
			}

			// Read field name using the proper field name reader
			fieldNameReader := d.reader.GetFieldNameReader()
			if fieldNameReader == nil {
				return ErrInvalidFormat
			}

			keyName, err := fieldNameReader.Read()
			if err != nil {
				return err
			}

			// Find the corresponding struct field
			if fieldValue, exists := fieldMap[keyName]; exists {
				// Handle pointer fields specially
				if fieldValue.Kind() == reflect.Ptr {
					// Create new instance if field is nil
					if fieldValue.IsNil() {
						newValue := reflect.New(fieldValue.Type().Elem())
						fieldValue.Set(newValue)
					}
					// Decode into the pointed-to value
					if err := d.decodeValue(fieldValue.Interface()); err != nil {
						return err
					}
				} else {
					// Decode directly into the field
					if err := d.decodeValue(fieldValue.Addr().Interface()); err != nil {
						return err
					}
				}
			} else {
				// Skip unknown fields by decoding into a throwaway interface{}
				var throwaway interface{}
				if err := d.decodeValue(&throwaway); err != nil {
					return err
				}
			}
		}
	} else {
		// Fixed length object
		for range length {
			// Read field name using the proper field name reader
			fieldNameReader := d.reader.GetFieldNameReader()
			if fieldNameReader == nil {
				return ErrInvalidFormat
			}

			keyName, err := fieldNameReader.Read()
			if err != nil {
				return err
			}

			// Find the corresponding struct field
			if fieldValue, exists := fieldMap[keyName]; exists {
				// Handle pointer fields specially
				if fieldValue.Kind() == reflect.Ptr {
					// Create new instance if field is nil
					if fieldValue.IsNil() {
						newValue := reflect.New(fieldValue.Type().Elem())
						fieldValue.Set(newValue)
					}
					// Decode into the pointed-to value
					if err := d.decodeValue(fieldValue.Interface()); err != nil {
						return err
					}
				} else {
					// Decode directly into the field
					if err := d.decodeValue(fieldValue.Addr().Interface()); err != nil {
						return err
					}
				}
			} else {
				// Skip unknown fields by decoding into a throwaway interface{}
				var throwaway interface{}
				if err := d.decodeValue(&throwaway); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// decodeSlice decodes an array into a slice using reflection
func (d *Decoder) decodeSlice(rv reflect.Value, length int) error {
	if length < -1 {
		return arrayLengthError(length)
	}

	elemType := rv.Type().Elem()

	// Handle EOF arrays (length == -1)
	if length == -1 {
		slice := reflect.MakeSlice(rv.Type(), 0, 0)
		for {
			token, err := d.reader.NextToken()
			if err == io.EOF {
				break // End of array
			}
			if err != nil {
				return err
			}

			// Create new element
			elem := reflect.New(elemType).Elem()
			if err := d.decodeToken(token, elem.Addr().Interface()); err != nil {
				return err
			}
			slice = reflect.Append(slice, elem)
		}
		rv.Set(slice)
		return nil
	}

	// Fixed length array
	slice := reflect.MakeSlice(rv.Type(), length, length)
	for i := range length {
		elem := slice.Index(i)
		if err := d.decodeValue(elem.Addr().Interface()); err != nil {
			return err
		}
	}
	rv.Set(slice)
	return nil
}

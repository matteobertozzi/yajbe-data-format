/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the 'License'); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an 'AS IS' BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package yajbe

import (
	"bytes"
	"fmt"
	"io"
	"math/big"
	"reflect"
	"strconv"
	"sync"
)

// Object pools for optimization
var (
	bufferPool = sync.Pool{
		New: func() any {
			return &bytes.Buffer{}
		},
	}

	writerPool = sync.Pool{
		New: func() any {
			return &Writer{}
		},
	}

	readerPool = sync.Pool{
		New: func() any {
			return &Reader{}
		},
	}
)

// Optimized caching
var (
	typeCache sync.Map
)

type structFieldInfo struct {
	Index int
	Name  string
	Type  reflect.Type
}

type structInfo struct {
	Fields []structFieldInfo
}

func Marshal(v any) ([]byte, error) {
	buf := bufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer bufferPool.Put(buf)

	w := writerPool.Get().(*Writer)
	w.reset(buf)
	defer func() {
		w.Close()
		writerPool.Put(w)
	}()

	if err := marshalValueFast(w, v); err != nil {
		return nil, err
	}

	if err := w.Flush(); err != nil {
		return nil, err
	}

	// Copy the result since we're returning the buffer to the pool
	result := make([]byte, buf.Len())
	copy(result, buf.Bytes())
	return result, nil
}

func Unmarshal(data []byte, v any) error {
	r := readerPool.Get().(*Reader)
	r.reset(data)
	defer readerPool.Put(r)

	value, err := r.ReadValue()
	if err != nil {
		return err
	}

	return assignValueFast(v, value)
}

type Decoder struct {
	r io.Reader
}

func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{r: r}
}

func (d *Decoder) Decode(v any) error {
	buf, err := io.ReadAll(d.r)
	if err != nil {
		return err
	}

	return Unmarshal(buf, v)
}

type Encoder struct {
	w io.Writer
}

func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{w: w}
}

func (e *Encoder) Encode(v any) error {
	data, err := Marshal(v)
	if err != nil {
		return err
	}

	for len(data) > 0 {
		bytesWritten, writeErr := e.w.Write(data)
		if writeErr != nil {
			return writeErr
		}

		if bytesWritten == 0 && len(data) > 0 {
			return io.ErrShortWrite
		}

		data = data[bytesWritten:]
	}
	return nil
}

// Optimized marshalValueFast with caching and fast paths
func marshalValueFast(w *Writer, v any) error {
	if v == nil {
		return w.WriteNull()
	}

	// Check for special types first
	if bigInt, ok := v.(*big.Int); ok {
		return w.WriteBigInt(bigInt)
	}

	val := reflect.ValueOf(v)
	typ := val.Type()

	// Fast path for common types
	switch typ.Kind() {
	case reflect.Bool:
		return w.WriteBool(val.Bool())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return w.WriteInt(val.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return w.WriteInt(int64(val.Uint()))
	case reflect.Float32:
		return w.WriteFloat32(float32(val.Float()))
	case reflect.Float64:
		return w.WriteFloat64(val.Float())
	case reflect.String:
		str := val.String()
		if str == "" {
			return w.WriteEmptyString()
		}
		return w.WriteString(str)
	case reflect.Slice, reflect.Array:
		if typ.Elem().Kind() == reflect.Uint8 {
			return w.WriteBytes(val.Bytes())
		}
		return marshalArrayFast(w, val)
	case reflect.Map:
		return marshalMapFast(w, val)
	case reflect.Struct:
		return marshalStructFast(w, val)
	case reflect.Ptr:
		if val.IsNil() {
			return w.WriteNull()
		}
		return marshalValueFast(w, val.Elem().Interface())
	case reflect.Interface:
		if val.IsNil() {
			return w.WriteNull()
		}
		return marshalValueFast(w, val.Elem().Interface())
	default:
		return fmt.Errorf("unsupported type: %T", v)
	}
}

func marshalArrayFast(w *Writer, val reflect.Value) error {
	length := val.Len()
	if err := w.WriteArrayHeader(length); err != nil {
		return err
	}

	for i := 0; i < length; i++ {
		if err := marshalValueFast(w, val.Index(i).Interface()); err != nil {
			return err
		}
	}

	return nil
}

func marshalMapFast(w *Writer, val reflect.Value) error {
	keys := val.MapKeys()
	if err := w.WriteObjectHeader(len(keys)); err != nil {
		return err
	}

	for _, key := range keys {
		if err := marshalValueFast(w, key.Interface()); err != nil {
			return err
		}
		if err := marshalValueFast(w, val.MapIndex(key).Interface()); err != nil {
			return err
		}
	}

	return nil
}

func marshalStructFast(w *Writer, val reflect.Value) error {
	typ := val.Type()

	var info *structInfo
	if cached, ok := typeCache.Load(typ); ok {
		info = cached.(*structInfo)
	} else {
		info = &structInfo{}
		numFields := typ.NumField()

		for i := 0; i < numFields; i++ {
			field := typ.Field(i)
			if field.IsExported() {
				fieldName := field.Name
				if tag := field.Tag.Get("json"); tag != "" && tag != "-" {
					if commaIdx := len(tag); commaIdx > 0 {
						for j, r := range tag {
							if r == ',' {
								commaIdx = j
								break
							}
						}
						fieldName = tag[:commaIdx]
					}
				}

				info.Fields = append(info.Fields, structFieldInfo{
					Index: i,
					Name:  fieldName,
					Type:  field.Type,
				})
			}
		}

		typeCache.Store(typ, info)
	}

	if err := w.WriteObjectHeader(len(info.Fields)); err != nil {
		return err
	}

	for _, fieldInfo := range info.Fields {
		if err := w.WriteString(fieldInfo.Name); err != nil {
			return err
		}
		if err := marshalValueFast(w, val.Field(fieldInfo.Index).Interface()); err != nil {
			return err
		}
	}

	return nil
}

// Optimized assignValueFast with type caching
func assignValueFast(dst any, src any) error {
	dstVal := reflect.ValueOf(dst)
	if dstVal.Kind() != reflect.Ptr {
		return fmt.Errorf("destination must be a pointer")
	}

	dstVal = dstVal.Elem()
	if !dstVal.CanSet() {
		return fmt.Errorf("destination is not settable")
	}

	return assignReflectValueFast(dstVal, src)
}

func assignReflectValueFast(dst reflect.Value, src any) error {
	if src == nil {
		dst.Set(reflect.Zero(dst.Type()))
		return nil
	}

	srcVal := reflect.ValueOf(src)
	dstType := dst.Type()

	if srcVal.Type().AssignableTo(dstType) {
		dst.Set(srcVal)
		return nil
	}

	// Fast path for common conversions
	switch dstType.Kind() {
	case reflect.Bool:
		if b, ok := src.(bool); ok {
			dst.SetBool(b)
			return nil
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return assignIntFast(dst, src)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return assignUintFast(dst, src)
	case reflect.Float32, reflect.Float64:
		return assignFloatFast(dst, src)
	case reflect.String:
		if s, ok := src.(string); ok {
			dst.SetString(s)
			return nil
		}
	case reflect.Slice:
		return assignSliceFast(dst, src)
	case reflect.Array:
		return assignArrayFast(dst, src)
	case reflect.Map:
		return assignMapFast(dst, src)
	case reflect.Struct:
		return assignStructFast(dst, src)
	case reflect.Ptr:
		if dst.IsNil() {
			dst.Set(reflect.New(dstType.Elem()))
		}
		return assignReflectValueFast(dst.Elem(), src)
	case reflect.Interface:
		dst.Set(srcVal)
		return nil
	}

	if dstType == reflect.TypeOf((*big.Int)(nil)) {
		if bigInt, ok := src.(*big.Int); ok {
			dst.Set(reflect.ValueOf(bigInt))
			return nil
		}
	}

	return fmt.Errorf("cannot assign %T to %s", src, dstType)
}

func assignIntFast(dst reflect.Value, src any) error {
	var val int64
	switch s := src.(type) {
	case int:
		val = int64(s)
	case int8:
		val = int64(s)
	case int16:
		val = int64(s)
	case int32:
		val = int64(s)
	case int64:
		val = s
	case uint:
		val = int64(s)
	case uint8:
		val = int64(s)
	case uint16:
		val = int64(s)
	case uint32:
		val = int64(s)
	case uint64:
		val = int64(s)
	case float32:
		val = int64(s)
	case float64:
		val = int64(s)
	case string:
		parsed, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return err
		}
		val = parsed
	default:
		return fmt.Errorf("cannot convert %T to int", src)
	}
	dst.SetInt(val)
	return nil
}

func assignUintFast(dst reflect.Value, src any) error {
	var val uint64
	switch s := src.(type) {
	case int:
		val = uint64(s)
	case int8:
		val = uint64(s)
	case int16:
		val = uint64(s)
	case int32:
		val = uint64(s)
	case int64:
		val = uint64(s)
	case uint:
		val = uint64(s)
	case uint8:
		val = uint64(s)
	case uint16:
		val = uint64(s)
	case uint32:
		val = uint64(s)
	case uint64:
		val = s
	case float32:
		val = uint64(s)
	case float64:
		val = uint64(s)
	case string:
		parsed, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			return err
		}
		val = parsed
	default:
		return fmt.Errorf("cannot convert %T to uint", src)
	}
	dst.SetUint(val)
	return nil
}

func assignFloatFast(dst reflect.Value, src any) error {
	var val float64
	switch s := src.(type) {
	case int:
		val = float64(s)
	case int8:
		val = float64(s)
	case int16:
		val = float64(s)
	case int32:
		val = float64(s)
	case int64:
		val = float64(s)
	case uint:
		val = float64(s)
	case uint8:
		val = float64(s)
	case uint16:
		val = float64(s)
	case uint32:
		val = float64(s)
	case uint64:
		val = float64(s)
	case float32:
		val = float64(s)
	case float64:
		val = s
	case string:
		parsed, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return err
		}
		val = parsed
	default:
		return fmt.Errorf("cannot convert %T to float", src)
	}
	dst.SetFloat(val)
	return nil
}

func assignSliceFast(dst reflect.Value, src any) error {
	srcSlice, ok := src.([]any)
	if !ok {
		return fmt.Errorf("expected []any, got %T", src)
	}

	slice := reflect.MakeSlice(dst.Type(), len(srcSlice), len(srcSlice))

	for i, item := range srcSlice {
		if err := assignReflectValueFast(slice.Index(i), item); err != nil {
			return err
		}
	}

	dst.Set(slice)
	return nil
}

func assignArrayFast(dst reflect.Value, src any) error {
	srcSlice, ok := src.([]any)
	if !ok {
		return fmt.Errorf("expected []any, got %T", src)
	}

	if len(srcSlice) != dst.Len() {
		return fmt.Errorf("array length mismatch: expected %d, got %d", dst.Len(), len(srcSlice))
	}

	for i, item := range srcSlice {
		if err := assignReflectValueFast(dst.Index(i), item); err != nil {
			return err
		}
	}

	return nil
}

func assignMapFast(dst reflect.Value, src any) error {
	srcMap, ok := src.(map[string]any)
	if !ok {
		return fmt.Errorf("expected map[string]any, got %T", src)
	}

	if dst.IsNil() {
		dst.Set(reflect.MakeMap(dst.Type()))
	}

	for key, value := range srcMap {
		keyVal := reflect.ValueOf(key)
		valueVal := reflect.New(dst.Type().Elem()).Elem()

		if err := assignReflectValueFast(valueVal, value); err != nil {
			return err
		}

		dst.SetMapIndex(keyVal, valueVal)
	}

	return nil
}

func assignStructFast(dst reflect.Value, src any) error {
	srcMap, ok := src.(map[string]any)
	if !ok {
		return fmt.Errorf("expected map[string]any, got %T", src)
	}

	dstType := dst.Type()

	// Check cache for struct info
	var info *structInfo
	if cached, ok := typeCache.Load(dstType); ok {
		info = cached.(*structInfo)
	} else {
		info = &structInfo{}
		numFields := dstType.NumField()

		for i := 0; i < numFields; i++ {
			field := dstType.Field(i)
			if field.IsExported() {
				fieldName := field.Name
				if tag := field.Tag.Get("json"); tag != "" && tag != "-" {
					if commaIdx := len(tag); commaIdx > 0 {
						for j, r := range tag {
							if r == ',' {
								commaIdx = j
								break
							}
						}
						fieldName = tag[:commaIdx]
					}
				}

				info.Fields = append(info.Fields, structFieldInfo{
					Index: i,
					Name:  fieldName,
					Type:  field.Type,
				})
			}
		}

		typeCache.Store(dstType, info)
	}

	for _, fieldInfo := range info.Fields {
		if value, exists := srcMap[fieldInfo.Name]; exists {
			if err := assignReflectValueFast(dst.Field(fieldInfo.Index), value); err != nil {
				return err
			}
		}
	}

	return nil
}

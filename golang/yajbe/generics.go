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

// Generic encoding helper functions to reduce code duplication

// encodeSliceGeneric encodes any slice type using a provided encoding function
func encodeSliceGeneric[T any](e *Encoder, slice []T, encodeItem func(T) error) error {
	if err := e.writer.WriteArrayStart(len(slice)); err != nil {
		return err
	}
	for _, item := range slice {
		if err := encodeItem(item); err != nil {
			return err
		}
	}
	return e.writer.WriteArrayEnd()
}

// encodeMapGeneric encodes any map[string]T type using a provided encoding function
func encodeMapGeneric[T any](e *Encoder, m map[string]T, encodeValue func(T) error) error {
	if err := e.writer.WriteObjectStart(len(m)); err != nil {
		return err
	}
	for key, value := range m {
		if err := e.writer.WriteFieldName(key); err != nil {
			return err
		}
		if err := encodeValue(value); err != nil {
			return err
		}
	}
	return e.writer.WriteObjectEnd()
}

// encodeNullableSlice handles null check and encoding for slice types
func encodeNullableSlice[T any](e *Encoder, slice []T, encodeSlice func([]T) error) error {
	if slice == nil {
		return e.writer.WriteNull()
	}
	return encodeSlice(slice)
}

// encodeNullableMap handles null check and encoding for map types
func encodeNullableMap[T any](e *Encoder, m map[string]T, encodeMap func(map[string]T) error) error {
	if m == nil {
		return e.writer.WriteNull()
	}
	return encodeMap(m)
}

// Generic decoding helper functions

// decodeObjectGeneric handles the common pattern of decoding objects for different map types
func decodeObjectGeneric[T any](
	d *Decoder,
	length int,
	makeMap func(int) map[string]T,
	decodeValue func(*T) error,
) (map[string]T, error) {
	obj := makeMap(length)
	
	if length == -1 {
		// EOF object - read until we hit end marker (0x01)
		for {
			// Check if next byte is end marker
			nextByte, err := d.reader.Peek()
			if err != nil {
				return nil, err
			}
			if nextByte == 0x01 {
				// Consume the end marker
				_, err := d.reader.ReadByte()
				if err != nil {
					return nil, err
				}
				break
			}
			
			// Read field name using the proper field name reader
			fieldNameReader := d.reader.GetFieldNameReader()
			if fieldNameReader == nil {
				return nil, ErrInvalidFormat
			}

			keyName, err := fieldNameReader.Read()
			if err != nil {
				return nil, err
			}

			var value T
			if err := decodeValue(&value); err != nil {
				return nil, err
			}
			obj[keyName] = value
		}
	} else {
		// Fixed length object
		for i := 0; i < length; i++ {
			// Read field name using the proper field name reader
			fieldNameReader := d.reader.GetFieldNameReader()
			if fieldNameReader == nil {
				return nil, ErrInvalidFormat
			}

			keyName, err := fieldNameReader.Read()
			if err != nil {
				return nil, err
			}

			var value T
			if err := decodeValue(&value); err != nil {
				return nil, err
			}
			obj[keyName] = value
		}
	}
	
	return obj, nil
}

// convertToInt64 converts various integer types to int64
func convertToInt64[T ~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32](val T) (int64, error) {
	result := int64(val)
	// Check for overflow in uint32 and larger unsigned types
	if uint64(val) > 9223372036854775807 { // max int64
		return 0, valueTooLargeError("integer")
	}
	return result, nil
}

// handleUint64 specifically handles uint64 with overflow check
func handleUint64(val uint64) (int64, error) {
	if val > 9223372036854775807 { // max int64
		return 0, valueTooLargeError("uint64")
	}
	return int64(val), nil
}
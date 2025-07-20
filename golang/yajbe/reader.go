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
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"math/big"
)

// Lookup table entry for fast value decoding
type lookupEntry struct {
	value any  // Pre-computed value for direct return (nil means needs decoding)
	kind  byte // Decoder kind: 0=direct, 1=intPos, 2=intNeg, 3=float32, etc.
}

// Pre-computed lookup table for all 256 possible header bytes
var valueLookup [256]lookupEntry

func init() {
	// Initialize the lookup table
	
	// Fixed values
	valueLookup[0] = lookupEntry{value: nil, kind: 0}           // null
	valueLookup[1] = lookupEntry{value: "EOF", kind: 0}         // EOF
	valueLookup[2] = lookupEntry{value: false, kind: 0}         // false
	valueLookup[3] = lookupEntry{value: true, kind: 0}          // true
	
	// Small positive integers (0x40-0x57): values 1-24
	for i := 0x40; i <= 0x57; i++ {
		value := 1 + (i - 0x40)
		valueLookup[i] = lookupEntry{value: value, kind: 0}
	}
	
	// Small negative integers (0x60-0x77): values 0 to -23
	for i := 0x60; i <= 0x77; i++ {
		value := -(i - 0x60)
		valueLookup[i] = lookupEntry{value: value, kind: 0}
	}
	
	// Larger positive integers (0x58-0x5F)
	for i := 0x58; i <= 0x5F; i++ {
		valueLookup[i] = lookupEntry{value: nil, kind: 1} // intPos
	}
	
	// Larger negative integers (0x78-0x7F)
	for i := 0x78; i <= 0x7F; i++ {
		valueLookup[i] = lookupEntry{value: nil, kind: 2} // intNeg
	}
	
	// Special types
	valueLookup[0x05] = lookupEntry{value: nil, kind: 3} // float32
	valueLookup[0x06] = lookupEntry{value: nil, kind: 4} // float64
	valueLookup[0x07] = lookupEntry{value: nil, kind: 5} // bigDecimal
	
	// Bytes (0x80-0xBF)
	for i := 0x80; i <= 0xBF; i++ {
		valueLookup[i] = lookupEntry{value: nil, kind: 6} // bytes
	}
	
	// Strings (0xC0-0xFF)
	for i := 0xC0; i <= 0xFF; i++ {
		valueLookup[i] = lookupEntry{value: nil, kind: 7} // string
	}
	
	// Arrays (0x20-0x2E, 0x2F)
	for i := 0x20; i <= 0x2E; i++ {
		valueLookup[i] = lookupEntry{value: nil, kind: 8} // array
	}
	valueLookup[0x2F] = lookupEntry{value: nil, kind: 8} // array
	
	// Objects (0x30-0x3E, 0x3F)
	for i := 0x30; i <= 0x3E; i++ {
		valueLookup[i] = lookupEntry{value: nil, kind: 9} // object
	}
	valueLookup[0x3F] = lookupEntry{value: nil, kind: 9} // object
}


type Reader struct {
	data []byte
	pos  int
}

func NewReader(data []byte) *Reader {
	return &Reader{
		data: data,
		pos:  0,
	}
}

func (r *Reader) Close() error {
	return nil
}

func (r *Reader) reset(data []byte) {
	r.data = data
	r.pos = 0
}

func (r *Reader) ensureBytes(n int) error {
	if r.pos+n > len(r.data) {
		return io.EOF
	}
	return nil
}

func (r *Reader) readByte() (byte, error) {
	if err := r.ensureBytes(1); err != nil {
		return 0, err
	}
	b := r.data[r.pos]
	r.pos++
	return b, nil
}

func (r *Reader) readBytes(n int) ([]byte, error) {
	if err := r.ensureBytes(n); err != nil {
		return nil, err
	}
	result := make([]byte, n)
	copy(result, r.data[r.pos:r.pos+n])
	r.pos += n
	return result, nil
}

// readBytesUnsafe returns a slice directly from the data without copying.
// Safe to use since we own the entire data slice.
func (r *Reader) readBytesUnsafe(n int) ([]byte, error) {
	if err := r.ensureBytes(n); err != nil {
		return nil, err
	}
	result := r.data[r.pos : r.pos+n]
	r.pos += n
	return result, nil
}

func readFixed(buf []byte, offset, width int) uint64 {
	// Use binary.BigEndian for consistent multi-byte operations
	switch width {
	case 1:
		return uint64(buf[offset])
	case 2:
		return uint64(binary.LittleEndian.Uint16(buf[offset:]))
	case 4:
		return uint64(binary.LittleEndian.Uint32(buf[offset:]))
	case 8:
		return binary.LittleEndian.Uint64(buf[offset:])
	default:
		// Fallback for other widths
		var result uint64
		for i := 0; i < width; i++ {
			result |= uint64(buf[offset+i]) << (8 * i)
		}
		return result
	}
}

func (r *Reader) readFixed(width int) (uint64, error) {
	data, err := r.readBytesUnsafe(width)
	if err != nil {
		return 0, err
	}
	return readFixed(data, 0, width), nil
}

func (r *Reader) readFixedInt(width int) (int, error) {
	val, err := r.readFixed(width)
	return int(val), err
}

func (r *Reader) ReadValue() (any, error) {
	head, err := r.readByte()
	if err != nil {
		return nil, err
	}

	entry := valueLookup[head]
	
	// Fast path: pre-computed value
	if entry.value != nil || entry.kind == 0 {
		return entry.value, nil
	}
	
	// Dispatch to specific decoder based on kind
	switch entry.kind {
	case 1: // intPos
		return r.decodeIntPositive(head)
	case 2: // intNeg
		return r.decodeIntNegative(head)
	case 3: // float32
		return r.decodeFloat32()
	case 4: // float64
		return r.decodeFloat64()
	case 5: // bigDecimal
		return r.decodeBigDecimal()
	case 6: // bytes
		return r.decodeBytes(head)
	case 7: // string
		return r.decodeString(head)
	case 8: // array
		return r.decodeArray(head)
	case 9: // object
		return r.decodeObject(head)
	default:
		return nil, fmt.Errorf("unknown type header: 0x%02x", head)
	}
}


func (r *Reader) decodeIntPositive(head byte) (any, error) {
	w := int(head & 0b11111)
	v, err := r.readFixed(w - 23)
	if err != nil {
		return nil, err
	}
	result := int64(25 + v)

	if result <= math.MaxInt32 {
		return int(result), nil
	}
	return result, nil
}

func (r *Reader) decodeIntNegative(head byte) (any, error) {
	w := int(head & 0b11111)
	v, err := r.readFixed(w - 23)
	if err != nil {
		return nil, err
	}
	result := -int64(v + 24)

	if result >= math.MinInt32 {
		return int(result), nil
	}
	return result, nil
}

func (r *Reader) decodeFloat32() (any, error) {
	bits, err := r.readFixedInt(4)
	if err != nil {
		return nil, err
	}
	return math.Float32frombits(uint32(bits)), nil
}

func (r *Reader) decodeFloat64() (any, error) {
	bits, err := r.readFixed(8)
	if err != nil {
		return nil, err
	}
	return math.Float64frombits(bits), nil
}

func (r *Reader) decodeBigDecimal() (any, error) {
	head, err := r.readByte()
	if err != nil {
		return nil, err
	}

	signedScale := (head & 0x80) == 0x80
	scaleBytes := 1 + int((head>>5)&3)
	precisionBytes := 1 + int((head>>3)&3)
	signedValue := (head & 4) == 4
	vDataBytes := 1 + int(head&3)

	scale, err := r.readFixedInt(scaleBytes)
	if err != nil {
		return nil, err
	}

	precision, err := r.readFixedInt(precisionBytes)
	if err != nil {
		return nil, err
	}

	vDataLength, err := r.readFixedInt(vDataBytes)
	if err != nil {
		return nil, err
	}

	data, err := r.readBytes(vDataLength)
	if err != nil {
		return nil, err
	}

	// Handle Java-style BigInteger encoding (with potential leading zero)
	unscaled := new(big.Int)
	if len(data) > 1 && data[0] == 0 && data[1]&0x80 != 0 {
		// Remove leading zero that Java adds for positive numbers
		unscaled.SetBytes(data[1:])
	} else {
		unscaled.SetBytes(data)
	}

	if signedValue {
		unscaled = unscaled.Neg(unscaled)
	}

	if scale == 0 && precision == 0 {
		return unscaled, nil
	}

	if signedScale {
		scale = -scale
	}

	return &BigDecimal{
		Unscaled:  unscaled,
		Scale:     scale,
		Precision: precision,
	}, nil
}

func (r *Reader) decodeBytes(head byte) (any, error) {
	length := int(head & 0b111111)
	if length > 59 {
		deltaLength, err := r.readFixedInt(length - 59)
		if err != nil {
			return nil, err
		}
		length = 59 + deltaLength
	}

	return r.readBytes(length)
}

func (r *Reader) decodeString(head byte) (any, error) {
	length := int(head & 0b111111)
	if length == 0 {
		return "", nil
	}

	if length > 59 {
		deltaLength, err := r.readFixedInt(length - 59)
		if err != nil {
			return nil, err
		}
		length = 59 + deltaLength
	}

	data, err := r.readBytesUnsafe(length)
	if err != nil {
		return nil, err
	}

	return string(data), nil
}

func (r *Reader) readItemCount(head byte) (int, error) {
	w := int(head & 0b1111)
	if w <= 10 {
		return w, nil
	}
	delta, err := r.readFixedInt(w - 10)
	if err != nil {
		return 0, err
	}
	return 10 + delta, nil
}

func (r *Reader) decodeArray(head byte) (any, error) {
	if head == 0b00101111 {
		result := make([]any, 0, 8)
		for {
			value, err := r.ReadValue()
			if err != nil {
				return nil, err
			}
			if value == "EOF" {
				break
			}
			result = append(result, value)
		}
		return result, nil
	}

	count, err := r.readItemCount(head)
	if err != nil {
		return nil, err
	}

	if count == 0 {
		return []any{}, nil
	}

	result := make([]any, count)
	for i := 0; i < count; i++ {
		value, err := r.ReadValue()
		if err != nil {
			return nil, err
		}
		result[i] = value
	}

	return result, nil
}

func (r *Reader) decodeObject(head byte) (any, error) {
	if head == 0b00111111 {
		result := make(map[string]any)
		for {
			key, err := r.ReadValue()
			if err != nil {
				return nil, err
			}
			if key == "EOF" {
				break
			}
			value, err := r.ReadValue()
			if err != nil {
				return nil, err
			}
			if keyStr, ok := key.(string); ok {
				result[keyStr] = value
			} else {
				return nil, fmt.Errorf("object key must be string, got %T", key)
			}
		}
		return result, nil
	}

	count, err := r.readItemCount(head)
	if err != nil {
		return nil, err
	}

	result := make(map[string]any, count)
	for i := 0; i < count; i++ {
		key, err := r.ReadValue()
		if err != nil {
			return nil, err
		}
		value, err := r.ReadValue()
		if err != nil {
			return nil, err
		}
		if keyStr, ok := key.(string); ok {
			result[keyStr] = value
		} else {
			return nil, fmt.Errorf("object key must be string, got %T", key)
		}
	}

	return result, nil
}

type BigDecimal struct {
	Unscaled  *big.Int
	Scale     int
	Precision int
}

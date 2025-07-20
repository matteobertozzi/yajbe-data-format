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

import (
	"encoding/binary"
	"io"
	"math"
)

// TokenType represents the type of a YAJBE token
type TokenType int

const (
	TokenNull TokenType = iota
	TokenBool
	TokenInt
	TokenFloat32
	TokenFloat64
	TokenString
	TokenBytes
	TokenArrayStart
	TokenArrayEnd
	TokenObjectStart
	TokenObjectEnd
	TokenFieldName
)

// Token represents a decoded YAJBE token
type Token struct {
	Type         TokenType
	BoolValue    bool
	IntValue     int64
	Float32Value float32
	Float64Value float64
	StringValue  string
	BytesValue   []byte
	Length       int // for arrays and objects
}

// Reader interface for reading YAJBE tokens
type Reader interface {
	NextToken() (Token, error)
	Peek() (byte, error)
	ReadByte() (byte, error)
	ReadString(n int) (string, error)
	ReadBytes(n int) ([]byte, error)
	ReadFixed(width int) (int64, error)
	ReadFixedInt(width int) (int32, error)
	GetFieldNameReader() *FieldNameReader
}

// ByteArrayReader implements Reader for byte arrays
type ByteArrayReader struct {
	buf []byte
	pos int
	enumMapping EnumMapping
	fieldNameReader *FieldNameReader
	tempBuf [16]byte // reusable buffer for number parsing
}

// StreamReader implements Reader for io.Reader
type StreamReader struct {
	reader io.Reader
	enumMapping EnumMapping
	fieldNameReader *FieldNameReader
}

// NewReaderFromBytes creates a new Reader from a byte array
func NewReaderFromBytes(data []byte) Reader {
	r := &ByteArrayReader{
		buf: data,
		pos: 0,
	}
	r.fieldNameReader = NewFieldNameReader(r)
	return r
}

// NewReaderFromReader creates a new Reader from an io.Reader
func NewReaderFromReader(reader io.Reader) Reader {
	r := &StreamReader{
		reader: reader,
	}
	r.fieldNameReader = NewFieldNameReader(r)
	return r
}

// ByteArrayReader implementation

func (r *ByteArrayReader) Peek() (byte, error) {
	if r.pos >= len(r.buf) {
		return 0, io.EOF
	}
	return r.buf[r.pos], nil
}

func (r *ByteArrayReader) ReadByte() (byte, error) {
	if r.pos >= len(r.buf) {
		return 0, io.EOF
	}
	b := r.buf[r.pos]
	r.pos++
	return b, nil
}

func (r *ByteArrayReader) ReadString(n int) (string, error) {
	if r.pos+n > len(r.buf) {
		return "", io.EOF
	}
	str := string(r.buf[r.pos : r.pos+n])
	r.pos += n
	return str, nil
}

func (r *ByteArrayReader) ReadBytes(n int) ([]byte, error) {
	if r.pos+n > len(r.buf) {
		return nil, io.EOF
	}
	// Return slice directly instead of copying for better performance
	// This is safe as long as callers don't modify the returned slice
	data := r.buf[r.pos : r.pos+n]
	r.pos += n
	return data, nil
}

func (r *ByteArrayReader) ReadFixed(width int) (int64, error) {
	if r.pos+width > len(r.buf) {
		return 0, io.EOF
	}
	result := readFixed(r.buf, r.pos, width)
	r.pos += width
	return result, nil
}

func (r *ByteArrayReader) ReadFixedInt(width int) (int32, error) {
	if r.pos+width > len(r.buf) {
		return 0, io.EOF
	}
	result := readFixedInt(r.buf, r.pos, width)
	r.pos += width
	return result, nil
}

func (r *ByteArrayReader) GetFieldNameReader() *FieldNameReader {
	return r.fieldNameReader
}

// StreamReader implementation

func (r *StreamReader) Peek() (byte, error) {
	// Note: This is a simplified implementation
	// In a real implementation, you'd need buffering
	var buf [1]byte
	n, err := r.reader.Read(buf[:])
	if err != nil {
		return 0, err
	}
	if n == 0 {
		return 0, io.EOF
	}
	// This is problematic as we can't "unread" the byte
	// A proper implementation would use a buffered reader
	return buf[0], nil
}

func (r *StreamReader) ReadByte() (byte, error) {
	var buf [1]byte
	n, err := r.reader.Read(buf[:])
	if err != nil {
		return 0, err
	}
	if n == 0 {
		return 0, io.EOF
	}
	return buf[0], nil
}

func (r *StreamReader) ReadString(n int) (string, error) {
	if n <= 16 {
		// Use stack allocation for small strings
		var stackBuf [16]byte
		buf := stackBuf[:n]
		_, err := io.ReadFull(r.reader, buf)
		if err != nil {
			return "", err
		}
		return string(buf), nil
	}
	// Fall back to heap allocation for larger strings
	buf := make([]byte, n)
	_, err := io.ReadFull(r.reader, buf)
	if err != nil {
		return "", err
	}
	return string(buf), nil
}

func (r *StreamReader) ReadBytes(n int) ([]byte, error) {
	if n <= 16 {
		// Use stack allocation for small byte arrays
		var stackBuf [16]byte
		buf := stackBuf[:n]
		_, err := io.ReadFull(r.reader, buf)
		if err != nil {
			return nil, err
		}
		// Must copy since we're returning a slice of stack memory
		result := make([]byte, n)
		copy(result, buf)
		return result, nil
	}
	// Fall back to heap allocation for larger arrays
	buf := make([]byte, n)
	_, err := io.ReadFull(r.reader, buf)
	if err != nil {
		return nil, err
	}
	return buf, nil
}

func (r *StreamReader) ReadFixed(width int) (int64, error) {
	if width <= 8 {
		// Use stack allocation for small fixed integers
		var stackBuf [8]byte
		buf := stackBuf[:width]
		_, err := io.ReadFull(r.reader, buf)
		if err != nil {
			return 0, err
		}
		return readFixed(buf, 0, width), nil
	}
	// Fall back to heap allocation for unusual widths
	buf := make([]byte, width)
	_, err := io.ReadFull(r.reader, buf)
	if err != nil {
		return 0, err
	}
	return readFixed(buf, 0, width), nil
}

func (r *StreamReader) ReadFixedInt(width int) (int32, error) {
	if width <= 4 {
		// Use stack allocation for small fixed integers
		var stackBuf [4]byte
		buf := stackBuf[:width]
		_, err := io.ReadFull(r.reader, buf)
		if err != nil {
			return 0, err
		}
		return readFixedInt(buf, 0, width), nil
	}
	// Fall back to heap allocation for unusual widths
	buf := make([]byte, width)
	_, err := io.ReadFull(r.reader, buf)
	if err != nil {
		return 0, err
	}
	return readFixedInt(buf, 0, width), nil
}

func (r *StreamReader) GetFieldNameReader() *FieldNameReader {
	return r.fieldNameReader
}

// TokenHandler represents different token parsing strategies
type TokenHandler int

const (
	HandlerNull TokenHandler = iota
	HandlerEOF
	HandlerBoolFalse
	HandlerBoolTrue
	HandlerVLEFloat // not implemented
	HandlerFloat32
	HandlerFloat64
	HandlerBigDecimal // not implemented
	HandlerEnumConfig
	HandlerEnumString8
	HandlerEnumString16
	HandlerInvalid
	HandlerSmallIntPos
	HandlerSmallIntNeg
	HandlerExternalIntPos
	HandlerExternalIntNeg
	HandlerBytes
	HandlerString
	HandlerArray
	HandlerObject
)

// TokenInfo precomputes parsing information for each byte value
type TokenInfo struct {
	Handler TokenHandler
	Value   int64 // for small ints, this is the decoded value directly
	Width   int   // for external ints, bytes, strings, arrays, objects
}

// Precomputed lookup table for all 256 possible header bytes
var tokenLookup [256]TokenInfo

// Initialize lookup table
func init() {
	// Initialize all to invalid first
	for i := range tokenLookup {
		tokenLookup[i] = TokenInfo{Handler: HandlerInvalid}
	}
	
	// Fixed tokens
	tokenLookup[0x00] = TokenInfo{Handler: HandlerNull}
	tokenLookup[0x01] = TokenInfo{Handler: HandlerEOF}
	tokenLookup[0x02] = TokenInfo{Handler: HandlerBoolFalse}
	tokenLookup[0x03] = TokenInfo{Handler: HandlerBoolTrue}
	tokenLookup[0x04] = TokenInfo{Handler: HandlerVLEFloat}
	tokenLookup[0x05] = TokenInfo{Handler: HandlerFloat32}
	tokenLookup[0x06] = TokenInfo{Handler: HandlerFloat64}
	tokenLookup[0x07] = TokenInfo{Handler: HandlerBigDecimal}
	tokenLookup[0x08] = TokenInfo{Handler: HandlerEnumConfig}
	tokenLookup[0x09] = TokenInfo{Handler: HandlerEnumString8}
	tokenLookup[0x0A] = TokenInfo{Handler: HandlerEnumString16}
	
	// Small positive integers (0x40-0x57): 1 to 24
	for i := 0x40; i <= 0x57; i++ {
		value := int64(i - 0x40 + 1)
		tokenLookup[i] = TokenInfo{Handler: HandlerSmallIntPos, Value: value}
	}
	
	// External positive integers (0x58-0x5F)
	for i := 0x58; i <= 0x5F; i++ {
		width := i - 0x58 + 1
		tokenLookup[i] = TokenInfo{Handler: HandlerExternalIntPos, Width: width}
	}
	
	// Small negative integers (0x60-0x77): 0 to -23
	for i := 0x60; i <= 0x77; i++ {
		value := -int64(i - 0x60)
		tokenLookup[i] = TokenInfo{Handler: HandlerSmallIntNeg, Value: value}
	}
	
	// External negative integers (0x78-0x7F)
	for i := 0x78; i <= 0x7F; i++ {
		width := i - 0x78 + 1
		tokenLookup[i] = TokenInfo{Handler: HandlerExternalIntNeg, Width: width}
	}
	
	// Bytes (0x80-0xBF)
	for i := 0x80; i <= 0xBF; i++ {
		length := i - 0x80
		tokenLookup[i] = TokenInfo{Handler: HandlerBytes, Width: length}
	}
	
	// Strings (0xC0-0xFF)
	for i := 0xC0; i <= 0xFF; i++ {
		length := i - 0xC0
		tokenLookup[i] = TokenInfo{Handler: HandlerString, Width: length}
	}
	
	// Arrays (0x20-0x2F)
	for i := 0x20; i <= 0x2F; i++ {
		if i == 0x2F {
			// EOF array
			tokenLookup[i] = TokenInfo{Handler: HandlerArray, Width: -1}
		} else {
			length := i - 0x20
			tokenLookup[i] = TokenInfo{Handler: HandlerArray, Width: length}
		}
	}
	
	// Objects (0x30-0x3F)
	for i := 0x30; i <= 0x3F; i++ {
		if i == 0x3F {
			// EOF object
			tokenLookup[i] = TokenInfo{Handler: HandlerObject, Width: -1}
		} else {
			length := i - 0x30
			tokenLookup[i] = TokenInfo{Handler: HandlerObject, Width: length}
		}
	}
}

// Common token parsing logic

func (r *ByteArrayReader) NextToken() (Token, error) {
	return r.nextToken()
}

func (r *StreamReader) NextToken() (Token, error) {
	return r.nextToken()
}

func (r *ByteArrayReader) nextToken() (Token, error) {
	head, err := r.ReadByte()
	if err != nil {
		return Token{}, err
	}
	return r.parseTokenFast(head)
}

func (r *StreamReader) nextToken() (Token, error) {
	head, err := r.ReadByte()
	if err != nil {
		return Token{}, err
	}
	return r.parseTokenFast(head)
}

func (r *ByteArrayReader) parseTokenFast(head byte) (Token, error) {
	return parseTokenFast(r, head)
}

func (r *StreamReader) parseTokenFast(head byte) (Token, error) {
	return parseTokenFast(r, head)
}

func parseTokenFast(r Reader, head byte) (Token, error) {
	info := tokenLookup[head]
	
	switch info.Handler {
	case HandlerNull:
		return Token{Type: TokenNull}, nil
	case HandlerEOF:
		return Token{}, io.EOF
	case HandlerBoolFalse:
		return Token{Type: TokenBool, BoolValue: false}, nil
	case HandlerBoolTrue:
		return Token{Type: TokenBool, BoolValue: true}, nil
	case HandlerVLEFloat:
		return Token{}, ErrInvalidFormat // VLE float not implemented
	case HandlerFloat32:
		// Read 4 bytes for float32 using ReadFixed for better performance
		bits, err := r.ReadFixed(4)
		if err != nil {
			return Token{}, err
		}
		// bits is already in little-endian format (ReadFixed) - use directly to match Dart implementation
		return Token{Type: TokenFloat32, Float32Value: math.Float32frombits(uint32(bits))}, nil
	case HandlerFloat64:
		// Read 8 bytes for float64 using ReadFixed for better performance
		bits, err := r.ReadFixed(8)
		if err != nil {
			return Token{}, err
		}
		// bits is already in little-endian format (ReadFixed)
		return Token{Type: TokenFloat64, Float64Value: math.Float64frombits(uint64(bits))}, nil
	case HandlerBigDecimal:
		return Token{}, ErrInvalidFormat // Big decimal not implemented
	case HandlerEnumConfig:
		if err := parseEnumConfig(r, head); err != nil {
			return Token{}, err
		}
		return r.NextToken() // Continue to next token
	case HandlerEnumString8, HandlerEnumString16:
		str, err := parseEnumString(r, head)
		if err != nil {
			return Token{}, err
		}
		return Token{Type: TokenString, StringValue: str}, nil
	case HandlerSmallIntPos, HandlerSmallIntNeg:
		// Value is precomputed in lookup table
		return Token{Type: TokenInt, IntValue: info.Value}, nil
	case HandlerExternalIntPos:
		v, err := r.ReadFixed(info.Width)
		if err != nil {
			return Token{}, err
		}
		return Token{Type: TokenInt, IntValue: 25 + v}, nil
	case HandlerExternalIntNeg:
		v, err := r.ReadFixed(info.Width)
		if err != nil {
			return Token{}, err
		}
		return Token{Type: TokenInt, IntValue: -(v + 24)}, nil
	case HandlerBytes:
		length, data, err := decodeBytesWithLength(r, info.Width)
		if err != nil {
			return Token{}, err
		}
		_ = length
		return Token{Type: TokenBytes, BytesValue: data}, nil
	case HandlerString:
		str, err := decodeStringWithLength(r, info.Width)
		if err != nil {
			return Token{}, err
		}
		return Token{Type: TokenString, StringValue: str}, nil
	case HandlerArray:
		if info.Width == -1 {
			return Token{Type: TokenArrayStart, Length: -1}, nil
		}
		length, err := readItemCountWithBase(r, info.Width)
		if err != nil {
			return Token{}, err
		}
		return Token{Type: TokenArrayStart, Length: length}, nil
	case HandlerObject:
		if info.Width == -1 {
			return Token{Type: TokenObjectStart, Length: -1}, nil
		}
		length, err := readItemCountWithBase(r, info.Width)
		if err != nil {
			return Token{}, err
		}
		return Token{Type: TokenObjectStart, Length: length}, nil
	default:
		return Token{}, ErrInvalidFormat
	}
}

// Helper functions

func readFixed(buf []byte, off, width int) int64 {
	// Use binary.LittleEndian for better performance than manual unrolling
	switch width {
	case 1:
		return int64(buf[off])
	case 2:
		return int64(binary.LittleEndian.Uint16(buf[off:]))
	case 4:
		return int64(binary.LittleEndian.Uint32(buf[off:]))
	case 8:
		return int64(binary.LittleEndian.Uint64(buf[off:]))
	default:
		// Fallback for unusual widths
		var result int64
		for i := 0; i < width; i++ {
			result |= int64(buf[off+i]) << (i * 8)
		}
		return result
	}
}

func readFixedInt(buf []byte, off, width int) int32 {
	// Use binary.LittleEndian for better performance than manual unrolling
	switch width {
	case 1:
		return int32(buf[off])
	case 2:
		return int32(binary.LittleEndian.Uint16(buf[off:]))
	case 4:
		return int32(binary.LittleEndian.Uint32(buf[off:]))
	default:
		// Fallback for unusual widths
		var result int32
		for i := 0; i < width; i++ {
			result |= int32(buf[off+i]) << (i * 8)
		}
		return result
	}
}



func decodeBytesWithLength(r Reader, baseLength int) (int, []byte, error) {
	var length int
	
	if baseLength <= 59 {
		length = baseLength
	} else {
		deltaLen, err := r.ReadFixedInt(baseLength - 59)
		if err != nil {
			return 0, nil, err
		}
		length = 59 + int(deltaLen)
	}
	
	data, err := r.ReadBytes(length)
	if err != nil {
		return 0, nil, err
	}
	
	return length, data, nil
}

func decodeStringWithLength(r Reader, baseLength int) (string, error) {
	var length int
	
	if baseLength <= 59 {
		length = baseLength
	} else {
		deltaLen, err := r.ReadFixedInt(baseLength - 59)
		if err != nil {
			return "", err
		}
		length = 59 + int(deltaLen)
	}
	
	return r.ReadString(length)
}

func readItemCountWithBase(r Reader, baseLength int) (int, error) {
	if baseLength <= 10 {
		return baseLength, nil
	}
	deltaLen, err := r.ReadFixedInt(baseLength - 10)
	if err != nil {
		return 0, err
	}
	return 10 + int(deltaLen), nil
}

func parseEnumConfig(r Reader, head byte) error {
	// Simplified enum config parsing
	h1, err := r.ReadByte()
	if err != nil {
		return err
	}
	
	switch (h1 >> 4) & 0x0F {
	case 0: // LRU
		freq, err := r.ReadByte()
		if err != nil {
			return err
		}
		// Create LRU enum mapping
		_ = freq // TODO: implement enum mapping
	}
	
	return nil
}

func parseEnumString(r Reader, head byte) (string, error) {
	switch head {
	case 0x09:
		index, err := r.ReadByte()
		if err != nil {
			return "", err
		}
		// TODO: get string from enum mapping
		_ = index
		return "", ErrInvalidFormat // Not implemented
	case 0x0A:
		index, err := r.ReadFixedInt(2)
		if err != nil {
			return "", err
		}
		// TODO: get string from enum mapping
		_ = index
		return "", ErrInvalidFormat // Not implemented
	}
	return "", ErrInvalidFormat
}
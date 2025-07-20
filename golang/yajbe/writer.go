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
	"math/bits"
)

// Writer interface for writing YAJBE tokens
type Writer interface {
	WriteNull() error
	WriteBool(value bool) error
	WriteInt(value int64) error
	WriteFloat32(value float32) error
	WriteFloat64(value float64) error
	WriteString(value string) error
	WriteBytes(value []byte) error
	WriteArrayStart(size int) error
	WriteArrayEnd() error
	WriteObjectStart(size int) error
	WriteObjectEnd() error
	WriteFieldName(name string) error
	Flush() error
}

// BufferWriter implements Writer using a byte buffer
type BufferWriter struct {
	buf             []byte
	enumMapping     EnumMapping
	fieldNameWriter *FieldNameWriter
	tempBuf         [16]byte // reusable buffer for encoding numbers
}

// StreamWriter implements Writer using an io.Writer
type StreamWriter struct {
	writer          io.Writer
	buf             []byte
	enumMapping     EnumMapping
	fieldNameWriter *FieldNameWriter
	tempBuf         [16]byte // reusable buffer for encoding numbers
}

// NewWriter creates a new BufferWriter
func NewWriter() *BufferWriter {
	w := &BufferWriter{
		buf: make([]byte, 0, 4096), // larger initial capacity
	}
	w.fieldNameWriter = NewFieldNameWriter(w)
	return w
}

// NewWriterFromBuffer creates a BufferWriter with an existing buffer
func NewWriterFromBuffer(buf []byte) *BufferWriter {
	w := &BufferWriter{
		buf: buf,
	}
	w.fieldNameWriter = NewFieldNameWriter(w)
	return w
}

// NewWriterFromWriter creates a new StreamWriter
func NewWriterFromWriter(writer io.Writer) *StreamWriter {
	w := &StreamWriter{
		writer: writer,
		buf:    make([]byte, 0, 4096), // larger initial capacity
	}
	w.fieldNameWriter = NewFieldNameWriter(w)
	return w
}

// BufferWriter implementation

func (w *BufferWriter) Bytes() []byte {
	return w.buf
}

func (w *BufferWriter) WriteNull() error {
	w.buf = append(w.buf, 0x00)
	return nil
}

func (w *BufferWriter) WriteBool(value bool) error {
	if value {
		w.buf = append(w.buf, 0x03)
	} else {
		w.buf = append(w.buf, 0x02)
	}
	return nil
}

func (w *BufferWriter) WriteInt(value int64) error {
	if value >= -23 && value <= 24 {
		return w.writeSmallInt(value)
	} else if value > 0 {
		return w.writeExternalInt(0x40, value-25)
	} else {
		return w.writeExternalInt(0x60, (-value)-24)
	}
}

func (w *BufferWriter) writeSmallInt(value int64) error {
	// Pre-calculate space and grow buffer once
	n := len(w.buf)
	if cap(w.buf) < n+1 {
		newBuf := make([]byte, n, (n+1)*2)
		copy(newBuf, w.buf)
		w.buf = newBuf
	}
	w.buf = w.buf[:n+1]

	if value > 0 {
		w.buf[n] = byte(0x40 | (value - 1))
	} else {
		w.buf[n] = byte(0x60 | (-value))
	}
	return nil
}

func (w *BufferWriter) writeExternalInt(head byte, value int64) error {
	width := calculateWidth(value)
	// Pre-calculate required space and grow buffer once
	n := len(w.buf)
	requiredSpace := 1 + width
	if cap(w.buf) < n+requiredSpace {
		newBuf := make([]byte, n, (n+requiredSpace)*2)
		copy(newBuf, w.buf)
		w.buf = newBuf
	}
	w.buf = w.buf[:n+requiredSpace]
	w.buf[n] = head | byte(23+width)
	// Inline writeFixed for better performance
	switch width {
	case 1:
		w.buf[n+1] = byte(value)
	case 2:
		binary.LittleEndian.PutUint16(w.buf[n+1:], uint16(value))
	case 4:
		binary.LittleEndian.PutUint32(w.buf[n+1:], uint32(value))
	case 8:
		binary.LittleEndian.PutUint64(w.buf[n+1:], uint64(value))
	default:
		// Fallback for unusual widths
		for i := range width {
			w.buf[n+1+i] = byte(value >> (i * 8))
		}
	}
	return nil
}

func (w *BufferWriter) WriteFloat32(value float32) error {
	bits := math.Float32bits(value)
	// Pre-calculate required space and grow buffer once
	n := len(w.buf)
	if cap(w.buf) < n+5 {
		newBuf := make([]byte, n, (n+5)*2)
		copy(newBuf, w.buf)
		w.buf = newBuf
	}
	w.buf = w.buf[:n+5]
	// Write float bits in little-endian order to match Dart implementation
	w.buf[n] = 0x05
	w.buf[n+1] = byte(bits)
	w.buf[n+2] = byte(bits >> 8)
	w.buf[n+3] = byte(bits >> 16)
	w.buf[n+4] = byte(bits >> 24)
	return nil
}

func (w *BufferWriter) WriteFloat64(value float64) error {
	bits := math.Float64bits(value)
	// Pre-calculate required space and grow buffer once
	n := len(w.buf)
	if cap(w.buf) < n+9 {
		newBuf := make([]byte, n, (n+9)*2)
		copy(newBuf, w.buf)
		w.buf = newBuf
	}
	w.buf = w.buf[:n+9]
	// Write float bits in little-endian order to match Dart implementation
	w.buf[n] = 0x06
	w.buf[n+1] = byte(bits)
	w.buf[n+2] = byte(bits >> 8)
	w.buf[n+3] = byte(bits >> 16)
	w.buf[n+4] = byte(bits >> 24)
	w.buf[n+5] = byte(bits >> 32)
	w.buf[n+6] = byte(bits >> 40)
	w.buf[n+7] = byte(bits >> 48)
	w.buf[n+8] = byte(bits >> 56)
	return nil
}

func (w *BufferWriter) WriteString(value string) error {
	if len(value) == 0 {
		w.buf = append(w.buf, 0xC0)
		return nil
	}

	// Use zero-copy string to bytes conversion
	utf8 := stringToBytes(value)
	return w.writeLength(0xC0, 59, len(utf8), utf8)
}

func (w *BufferWriter) WriteBytes(value []byte) error {
	return w.writeLength(0x80, 59, len(value), value)
}

func (w *BufferWriter) writeLength(head byte, inlineMax, length int, data []byte) error {
	var headerSize int
	if length <= inlineMax {
		headerSize = 1
	} else {
		deltaLength := length - inlineMax
		width := calculateWidth(int64(deltaLength))
		headerSize = 1 + width
	}

	// Pre-calculate total space needed and grow buffer once
	totalSize := headerSize + len(data)
	n := len(w.buf)
	if cap(w.buf) < n+totalSize {
		newBuf := make([]byte, n, (n+totalSize)*2)
		copy(newBuf, w.buf)
		w.buf = newBuf
	}
	w.buf = w.buf[:n+totalSize]

	if length <= inlineMax {
		w.buf[n] = head | byte(length)
		copy(w.buf[n+1:], data)
	} else {
		deltaLength := length - inlineMax
		width := calculateWidth(int64(deltaLength))
		w.buf[n] = head | byte(inlineMax+width)
		// Inline writeFixed for small widths
		switch width {
		case 1:
			w.buf[n+1] = byte(deltaLength)
		case 2:
			w.buf[n+1] = byte(deltaLength)
			w.buf[n+2] = byte(deltaLength >> 8)
		default:
			for i := range width {
				w.buf[n+1+i] = byte(deltaLength >> (i * 8))
			}
		}
		copy(w.buf[n+headerSize:], data)
	}
	return nil
}

func (w *BufferWriter) WriteArrayStart(size int) error {
	if size < 0 {
		// EOF array
		w.buf = append(w.buf, 0x2F)
		return nil
	}

	if size <= 10 {
		w.buf = append(w.buf, byte(0x20|size))
	} else {
		deltaLength := size - 10
		width := calculateWidth(int64(deltaLength))
		// Pre-calculate required space and grow buffer once
		n := len(w.buf)
		requiredSpace := 1 + width
		if cap(w.buf) < n+requiredSpace {
			newBuf := make([]byte, n, (n+requiredSpace)*2)
			copy(newBuf, w.buf)
			w.buf = newBuf
		}
		w.buf = w.buf[:n+requiredSpace]
		w.buf[n] = byte(0x20 | (10 + width))
		// Inline writeFixed for better performance
		switch width {
		case 1:
			w.buf[n+1] = byte(deltaLength)
		case 2:
			binary.LittleEndian.PutUint16(w.buf[n+1:], uint16(deltaLength))
		case 4:
			binary.LittleEndian.PutUint32(w.buf[n+1:], uint32(deltaLength))
		case 8:
			binary.LittleEndian.PutUint64(w.buf[n+1:], uint64(deltaLength))
		default:
			// Fallback for unusual widths
			for i := range width {
				w.buf[n+1+i] = byte(deltaLength >> (i * 8))
			}
		}
	}
	return nil
}

func (w *BufferWriter) WriteArrayEnd() error {
	// Array end is a no-op for fixed-size arrays
	// EOF arrays are handled by the caller if needed
	return nil
}

func (w *BufferWriter) WriteObjectStart(size int) error {
	if size < 0 {
		// EOF object
		w.buf = append(w.buf, 0x3F)
		return nil
	}

	if size <= 10 {
		w.buf = append(w.buf, byte(0x30|size))
	} else {
		deltaLength := size - 10
		width := calculateWidth(int64(deltaLength))
		// Pre-calculate required space and grow buffer once
		n := len(w.buf)
		requiredSpace := 1 + width
		if cap(w.buf) < n+requiredSpace {
			newBuf := make([]byte, n, (n+requiredSpace)*2)
			copy(newBuf, w.buf)
			w.buf = newBuf
		}
		w.buf = w.buf[:n+requiredSpace]
		w.buf[n] = byte(0x30 | (10 + width))
		// Inline writeFixed for better performance
		switch width {
		case 1:
			w.buf[n+1] = byte(deltaLength)
		case 2:
			binary.LittleEndian.PutUint16(w.buf[n+1:], uint16(deltaLength))
		case 4:
			binary.LittleEndian.PutUint32(w.buf[n+1:], uint32(deltaLength))
		case 8:
			binary.LittleEndian.PutUint64(w.buf[n+1:], uint64(deltaLength))
		default:
			// Fallback for unusual widths
			for i := range width {
				w.buf[n+1+i] = byte(deltaLength >> (i * 8))
			}
		}
	}
	return nil
}

func (w *BufferWriter) WriteObjectEnd() error {
	// Object end is a no-op for fixed-size objects
	// EOF objects are handled by the caller if needed
	return nil
}

func (w *BufferWriter) WriteFieldName(name string) error {
	return w.fieldNameWriter.Write(name)
}

func (w *BufferWriter) Flush() error {
	return nil
}

func (w *BufferWriter) writeFixed(value int64, width int) {
	// Pre-calculate required space and grow buffer once
	n := len(w.buf)
	if cap(w.buf) < n+width {
		newBuf := make([]byte, n, (n+width)*2)
		copy(newBuf, w.buf)
		w.buf = newBuf
	}
	w.buf = w.buf[:n+width]
	// Use binary.LittleEndian for better performance
	switch width {
	case 1:
		w.buf[n] = byte(value)
	case 2:
		binary.LittleEndian.PutUint16(w.buf[n:], uint16(value))
	case 4:
		binary.LittleEndian.PutUint32(w.buf[n:], uint32(value))
	case 8:
		binary.LittleEndian.PutUint64(w.buf[n:], uint64(value))
	default:
		// Fallback for unusual widths
		for i := range width {
			w.buf[n+i] = byte(value >> (i * 8))
		}
	}
}

func (w *BufferWriter) writeFixedInt(value int32, width int) {
	// Pre-calculate required space and grow buffer once
	n := len(w.buf)
	if cap(w.buf) < n+width {
		newBuf := make([]byte, n, (n+width)*2)
		copy(newBuf, w.buf)
		w.buf = newBuf
	}
	w.buf = w.buf[:n+width]
	// Use binary.LittleEndian for better performance
	switch width {
	case 1:
		w.buf[n] = byte(value)
	case 2:
		binary.LittleEndian.PutUint16(w.buf[n:], uint16(value))
	case 4:
		binary.LittleEndian.PutUint32(w.buf[n:], uint32(value))
	default:
		// Fallback for unusual widths
		for i := range width {
			w.buf[n+i] = byte(value >> (i * 8))
		}
	}
}

// StreamWriter implementation

func (w *StreamWriter) WriteNull() error {
	_, err := w.writer.Write([]byte{0x00})
	return err
}

func (w *StreamWriter) WriteBool(value bool) error {
	var b byte = 0x02
	if value {
		b = 0x03
	}
	_, err := w.writer.Write([]byte{b})
	return err
}

func (w *StreamWriter) WriteInt(value int64) error {
	if value >= -23 && value <= 24 {
		return w.writeSmallInt(value)
	} else if value > 0 {
		return w.writeExternalInt(0x40, value-25)
	} else {
		return w.writeExternalInt(0x60, (-value)-24)
	}
}

func (w *StreamWriter) writeSmallInt(value int64) error {
	var b byte
	if value > 0 {
		b = byte(0x40 | (value - 1))
	} else {
		b = byte(0x60 | (-value))
	}
	_, err := w.writer.Write([]byte{b})
	return err
}

func (w *StreamWriter) writeExternalInt(head byte, value int64) error {
	width := calculateWidth(value)
	w.buf = w.buf[:0] // reset buffer
	w.buf = append(w.buf, head|byte(23+width))
	w.writeFixedToBuf(value, width)
	_, err := w.writer.Write(w.buf)
	return err
}

func (w *StreamWriter) WriteFloat32(value float32) error {
	bits := math.Float32bits(value)
	// Use temp buffer to avoid allocation
	w.tempBuf[0] = 0x05
	w.tempBuf[1] = byte(bits)
	w.tempBuf[2] = byte(bits >> 8)
	w.tempBuf[3] = byte(bits >> 16)
	w.tempBuf[4] = byte(bits >> 24)
	_, err := w.writer.Write(w.tempBuf[:5])
	return err
}

func (w *StreamWriter) WriteFloat64(value float64) error {
	bits := math.Float64bits(value)
	// Use temp buffer to avoid allocation
	w.tempBuf[0] = 0x06
	w.tempBuf[1] = byte(bits)
	w.tempBuf[2] = byte(bits >> 8)
	w.tempBuf[3] = byte(bits >> 16)
	w.tempBuf[4] = byte(bits >> 24)
	w.tempBuf[5] = byte(bits >> 32)
	w.tempBuf[6] = byte(bits >> 40)
	w.tempBuf[7] = byte(bits >> 48)
	w.tempBuf[8] = byte(bits >> 56)
	_, err := w.writer.Write(w.tempBuf[:9])
	return err
}

func (w *StreamWriter) WriteString(value string) error {
	if len(value) == 0 {
		_, err := w.writer.Write([]byte{0xC0})
		return err
	}

	// Use zero-copy string to bytes conversion
	utf8 := stringToBytes(value)
	return w.writeLength(0xC0, 59, len(utf8), utf8)
}

func (w *StreamWriter) WriteBytes(value []byte) error {
	return w.writeLength(0x80, 59, len(value), value)
}

func (w *StreamWriter) writeLength(head byte, inlineMax, length int, data []byte) error {
	w.buf = w.buf[:0] // reset buffer

	if length <= inlineMax {
		w.buf = append(w.buf, head|byte(length))
	} else {
		deltaLength := length - inlineMax
		width := calculateWidth(int64(deltaLength))
		w.buf = append(w.buf, head|byte(inlineMax+width))
		w.writeFixedToBuf(int64(deltaLength), width)
	}

	if _, err := w.writer.Write(w.buf); err != nil {
		return err
	}
	_, err := w.writer.Write(data)
	return err
}

func (w *StreamWriter) WriteArrayStart(size int) error {
	if size < 0 {
		// EOF array
		_, err := w.writer.Write([]byte{0x2F})
		return err
	}

	w.buf = w.buf[:0] // reset buffer
	if size <= 10 {
		w.buf = append(w.buf, byte(0x20|size))
	} else {
		deltaLength := size - 10
		width := calculateWidth(int64(deltaLength))
		w.buf = append(w.buf, byte(0x20|(10+width)))
		w.writeFixedToBuf(int64(deltaLength), width)
	}
	_, err := w.writer.Write(w.buf)
	return err
}

func (w *StreamWriter) WriteArrayEnd() error {
	// Array end is a no-op for fixed-size arrays
	// EOF arrays are handled by the caller if needed
	return nil
}

func (w *StreamWriter) WriteObjectStart(size int) error {
	if size < 0 {
		// EOF object
		_, err := w.writer.Write([]byte{0x3F})
		return err
	}

	w.buf = w.buf[:0] // reset buffer
	if size <= 10 {
		w.buf = append(w.buf, byte(0x30|size))
	} else {
		deltaLength := size - 10
		width := calculateWidth(int64(deltaLength))
		w.buf = append(w.buf, byte(0x30|(10+width)))
		w.writeFixedToBuf(int64(deltaLength), width)
	}
	_, err := w.writer.Write(w.buf)
	return err
}

func (w *StreamWriter) WriteObjectEnd() error {
	// Object end is a no-op for fixed-size objects
	// EOF objects are handled by the caller if needed
	return nil
}

func (w *StreamWriter) WriteFieldName(name string) error {
	return w.fieldNameWriter.Write(name)
}

func (w *StreamWriter) Flush() error {
	if flusher, ok := w.writer.(interface{ Flush() error }); ok {
		return flusher.Flush()
	}
	return nil
}

func (w *StreamWriter) writeFixedToBuf(value int64, width int) {
	// Ensure buffer has enough space
	for len(w.buf) < width {
		w.buf = append(w.buf, 0)
	}
	// Use binary.LittleEndian for better performance
	switch width {
	case 1:
		w.buf[0] = byte(value)
	case 2:
		binary.LittleEndian.PutUint16(w.buf[:2], uint16(value))
	case 4:
		binary.LittleEndian.PutUint32(w.buf[:4], uint32(value))
	case 8:
		binary.LittleEndian.PutUint64(w.buf[:8], uint64(value))
	default:
		// Fallback for unusual widths
		for i := range width {
			w.buf[i] = byte(value >> (i * 8))
		}
	}
	w.buf = w.buf[:width] // trim to exact size
}

func (w *StreamWriter) writeFixedIntToBuf(value int32, width int) {
	// Ensure buffer has enough space
	for len(w.buf) < width {
		w.buf = append(w.buf, 0)
	}
	// Use binary.LittleEndian for better performance
	switch width {
	case 1:
		w.buf[0] = byte(value)
	case 2:
		binary.LittleEndian.PutUint16(w.buf[:2], uint16(value))
	case 4:
		binary.LittleEndian.PutUint32(w.buf[:4], uint32(value))
	default:
		// Fallback for unusual widths
		for i := range width {
			w.buf[i] = byte(value >> (i * 8))
		}
	}
	w.buf = w.buf[:width] // trim to exact size
}

// Helper functions

// Optimized width calculation using lookup table for common values
func calculateWidth(value int64) int {
	if value == 0 {
		return 1
	}
	uval := uint64(value)
	// Use faster bit manipulation for common cases
	switch {
	case uval < 256: // 1 byte
		return 1
	case uval < 65536: // 2 bytes
		return 2
	case uval < 16777216: // 3 bytes
		return 3
	case uval < 4294967296: // 4 bytes
		return 4
	default:
		// Use bit counting for larger values
		return (64 - bits.LeadingZeros64(uval) + 7) / 8
	}
}

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
	"io"
	"math"
	"math/big"
	"math/bits"
	"sync"
)

var writerBufferPool = sync.Pool{
	New: func() interface{} {
		return make([]byte, 4096)
	},
}

type Writer struct {
	w      io.Writer
	buffer []byte
	pos    int
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{
		w:      w,
		buffer: writerBufferPool.Get().([]byte),
		pos:    0,
	}
}

func (w *Writer) Close() error {
	if err := w.Flush(); err != nil {
		return err
	}
	if w.buffer != nil {
		writerBufferPool.Put(w.buffer[:cap(w.buffer)])
		w.buffer = nil
	}
	return nil
}

func (w *Writer) reset(writer io.Writer) {
	w.w = writer
	w.pos = 0
	if w.buffer == nil {
		w.buffer = writerBufferPool.Get().([]byte)
	}
}

func (w *Writer) Flush() error {
	if w.pos > 0 {
		_, err := w.w.Write(w.buffer[:w.pos])
		w.pos = 0
		return err
	}
	return nil
}

func (w *Writer) ensureSpace(n int) error {
	if w.pos+n > len(w.buffer) {
		if err := w.Flush(); err != nil {
			return err
		}
		if n > len(w.buffer) {
			w.buffer = make([]byte, n*2)
		}
	}
	return nil
}

func (w *Writer) writeByte(b byte) error {
	if err := w.ensureSpace(1); err != nil {
		return err
	}
	w.buffer[w.pos] = b
	w.pos++
	return nil
}

func (w *Writer) writeBytes(data []byte) error {
	if err := w.ensureSpace(len(data)); err != nil {
		return err
	}
	copy(w.buffer[w.pos:], data)
	w.pos += len(data)
	return nil
}

func writeFixed(buf []byte, offset int, value uint64, width int) {
	// Use binary.LittleEndian for consistent multi-byte operations
	switch width {
	case 1:
		buf[offset] = byte(value)
	case 2:
		binary.LittleEndian.PutUint16(buf[offset:], uint16(value))
	case 4:
		binary.LittleEndian.PutUint32(buf[offset:], uint32(value))
	case 8:
		binary.LittleEndian.PutUint64(buf[offset:], value)
	default:
		// Fallback for other widths
		for i := 0; i < width; i++ {
			buf[offset+i] = byte(value >> (8 * i))
		}
	}
}

func (w *Writer) WriteNull() error {
	return w.writeByte(0)
}

func (w *Writer) WriteEof() error {
	return w.writeByte(1)
}

func (w *Writer) WriteBool(value bool) error {
	if value {
		return w.writeByte(0b11)
	}
	return w.writeByte(0b10)
}

func (w *Writer) WriteFloat32(value float32) error {
	if err := w.ensureSpace(5); err != nil {
		return err
	}
	w.buffer[w.pos] = 0b00000_101
	bits := math.Float32bits(value)
	writeFixed(w.buffer, w.pos+1, uint64(bits), 4)
	w.pos += 5
	return nil
}

func (w *Writer) WriteFloat64(value float64) error {
	if err := w.ensureSpace(9); err != nil {
		return err
	}
	w.buffer[w.pos] = 0b00000_110
	bits := math.Float64bits(value)
	writeFixed(w.buffer, w.pos+1, bits, 8)
	w.pos += 9
	return nil
}

func (w *Writer) WriteBigInt(value *big.Int) error {
	return w.writeBigDecimal(0, 0, value)
}

func (w *Writer) writeBigDecimal(scale, precision int, unscaled *big.Int) error {
	signedValue := unscaled.Sign() < 0
	if signedValue {
		unscaled = new(big.Int).Abs(unscaled)
	}

	signedScale := scale < 0
	if signedScale {
		scale = -scale
	}

	vData := unscaled.Bytes()
	if len(vData) == 0 {
		vData = []byte{0}
	} else if len(vData) > 0 && vData[0]&0x80 != 0 {
		// Java adds leading zero if MSB is set (for both positive and absolute value of negative)
		vData = append([]byte{0}, vData...)
	}
	// Use Java's calculation: ((32 - Integer.numberOfLeadingZeros(vData.length)) + 7) >> 3
	vDataLengthBytes := 1
	if len(vData) > 0 {
		vDataLengthBytes = max(1, (32-bits.LeadingZeros32(uint32(len(vData)))+7)/8)
	}
	scaleBytes := 1
	if scale > 0 {
		scaleBytes = max(1, (32-bits.LeadingZeros32(uint32(scale))+7)/8)
	}
	precisionBytes := 1
	if precision > 0 {
		precisionBytes = max(1, (32-bits.LeadingZeros32(uint32(precision))+7)/8)
	}

	totalSize := 2 + scaleBytes + precisionBytes + vDataLengthBytes + len(vData)
	if err := w.ensureSpace(totalSize); err != nil {
		return err
	}

	w.buffer[w.pos] = 0b00000_111
	w.pos++

	flags := byte(0)
	if signedScale {
		flags |= 0x80
	}
	flags |= byte((scaleBytes - 1) << 5)
	flags |= byte((precisionBytes - 1) << 3)
	if signedValue {
		flags |= 4
	}
	flags |= byte(vDataLengthBytes - 1)

	w.buffer[w.pos] = flags
	w.pos++

	writeFixed(w.buffer, w.pos, uint64(scale), scaleBytes)
	w.pos += scaleBytes
	writeFixed(w.buffer, w.pos, uint64(precision), precisionBytes)
	w.pos += precisionBytes
	writeFixed(w.buffer, w.pos, uint64(len(vData)), vDataLengthBytes)
	w.pos += vDataLengthBytes

	copy(w.buffer[w.pos:], vData)
	w.pos += len(vData)
	return nil
}

func (w *Writer) WriteInt(value int64) error {
	// Fast path for small integers - no additional bytes needed
	if value >= -23 && value <= 24 {
		return w.writeSmallInt(int(value))
	}
	
	// Fast path for single-byte integers
	if value > 24 && value <= 280 {
		if err := w.ensureSpace(2); err != nil {
			return err
		}
		w.buffer[w.pos] = 0b010_00000 | 24
		w.buffer[w.pos+1] = byte(value - 25)
		w.pos += 2
		return nil
	}
	
	if value < -23 && value >= -279 {
		if err := w.ensureSpace(2); err != nil {
			return err
		}
		w.buffer[w.pos] = 0b011_00000 | 24
		w.buffer[w.pos+1] = byte((-value) - 24)
		w.pos += 2
		return nil
	}
	
	// General case
	if value > 0 {
		return w.writeExternalInt(0b010_00000, uint64(value-25))
	} else {
		return w.writeExternalInt(0b011_00000, uint64((-value)-24))
	}
}

func (w *Writer) writeSmallInt(value int) error {
	if value > 0 {
		return w.writeByte(byte(0b010_00000 | (value - 1)))
	}
	return w.writeByte(byte(0b011_00000 | (-value)))
}

func (w *Writer) writeExternalInt(head byte, value uint64) error {
	width := max(1, (bits.Len64(value)+7)/8)
	if err := w.ensureSpace(1 + width); err != nil {
		return err
	}

	w.buffer[w.pos] = head | byte(23+width)
	writeFixed(w.buffer, w.pos+1, value, width)
	w.pos += 1 + width
	return nil
}

func (w *Writer) writeLength(head byte, inlineMax int, length int) error {
	if length <= inlineMax {
		return w.writeByte(head | byte(length))
	}

	deltaLength := length - inlineMax
	bytes := max(1, (bits.Len(uint(deltaLength))+7)/8)
	if err := w.ensureSpace(1 + bytes); err != nil {
		return err
	}

	w.buffer[w.pos] = head | byte(inlineMax+bytes)
	writeFixed(w.buffer, w.pos+1, uint64(deltaLength), bytes)
	w.pos += 1 + bytes
	return nil
}

func (w *Writer) WriteBytes(data []byte) error {
	if err := w.writeLength(0b10_000000, 59, len(data)); err != nil {
		return err
	}
	return w.writeBytes(data)
}

func (w *Writer) WriteEmptyString() error {
	return w.writeByte(0b11_000000)
}

func (w *Writer) WriteString(text string) error {
	utf8 := []byte(text)
	if err := w.writeLength(0b11_000000, 59, len(utf8)); err != nil {
		return err
	}
	return w.writeBytes(utf8)
}

func (w *Writer) WriteArrayHeader(size int) error {
	if size < 0 {
		return w.writeByte(0b0010_1111)
	}
	return w.writeLength(0b0010_0000, 10, size)
}

func (w *Writer) WriteObjectHeader(size int) error {
	if size < 0 {
		return w.writeByte(0b0011_1111)
	}
	return w.writeLength(0b0011_0000, 10, size)
}


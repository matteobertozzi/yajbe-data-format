package yajbe

import (
	"sync"
	"unsafe"
)

// stringToBytes converts string to []byte without allocation using unsafe
// This is safe because we never modify the returned slice
func stringToBytes(s string) []byte {
	if len(s) == 0 {
		return nil
	}
	return unsafe.Slice((*byte)(unsafe.StringData(s)), len(s))
}

// Pool management for reducing allocations
var (
	// Slice pools for different types and sizes
	interfaceSlicePool = sync.Pool{
		New: func() interface{} { return make([]interface{}, 0, 16) },
	}
	boolSlicePool = sync.Pool{
		New: func() interface{} { return make([]bool, 0, 16) },
	}
	int64SlicePool = sync.Pool{
		New: func() interface{} { return make([]int64, 0, 16) },
	}
	stringSlicePool = sync.Pool{
		New: func() interface{} { return make([]string, 0, 16) },
	}
	
	// Map pools for objects
	stringInterfaceMapPool = sync.Pool{
		New: func() interface{} { return make(map[string]interface{}, 16) },
	}
	
	// Byte buffer pools for field names and temporary operations
	smallBytePool = sync.Pool{
		New: func() interface{} { return make([]byte, 0, 64) },
	}
	mediumBytePool = sync.Pool{
		New: func() interface{} { return make([]byte, 0, 256) },
	}
	largeBytePool = sync.Pool{
		New: func() interface{} { return make([]byte, 0, 1024) },
	}
	
	// Writer buffer pools
	writerBufferPool = sync.Pool{
		New: func() interface{} { return make([]byte, 0, 4096) },
	}
	encoderPool = sync.Pool{
		New: func() interface{} { return &Encoder{} },
	}
	decoderPool = sync.Pool{
		New: func() interface{} { return &Decoder{} },
	}
)

// getInterfaceSlice gets a []interface{} from pool, ensuring it has the required capacity
func getInterfaceSlice(length int) []interface{} {
	if length < 0 {
		return nil
	}
	slice := interfaceSlicePool.Get().([]interface{})
	if cap(slice) < length {
		// Pool slice too small, allocate new one
		return make([]interface{}, length)
	}
	return slice[:length]
}

// putInterfaceSlice returns a []interface{} to the pool
func putInterfaceSlice(slice []interface{}) {
	if cap(slice) <= 64 { // Only pool reasonably sized slices
		slice = slice[:0] // Reset length but keep capacity
		interfaceSlicePool.Put(slice)
	}
}

// getBoolSlice gets a []bool from pool, ensuring it has the required capacity
func getBoolSlice(length int) []bool {
	if length < 0 {
		return nil
	}
	slice := boolSlicePool.Get().([]bool)
	if cap(slice) < length {
		return make([]bool, length)
	}
	return slice[:length]
}

// putBoolSlice returns a []bool to the pool
func putBoolSlice(slice []bool) {
	if cap(slice) <= 64 {
		slice = slice[:0]
		boolSlicePool.Put(slice)
	}
}

// getInt64Slice gets a []int64 from pool, ensuring it has the required capacity
func getInt64Slice(length int) []int64 {
	if length < 0 {
		return nil
	}
	slice := int64SlicePool.Get().([]int64)
	if cap(slice) < length {
		return make([]int64, length)
	}
	return slice[:length]
}

// putInt64Slice returns a []int64 to the pool
func putInt64Slice(slice []int64) {
	if cap(slice) <= 64 {
		slice = slice[:0]
		int64SlicePool.Put(slice)
	}
}

// getStringSlice gets a []string from pool, ensuring it has the required capacity
func getStringSlice(length int) []string {
	if length < 0 {
		return nil
	}
	slice := stringSlicePool.Get().([]string)
	if cap(slice) < length {
		return make([]string, length)
	}
	return slice[:length]
}

// putStringSlice returns a []string to the pool
func putStringSlice(slice []string) {
	if cap(slice) <= 64 {
		// Clear strings to avoid memory leaks
		for i := range slice {
			slice[i] = ""
		}
		slice = slice[:0]
		stringSlicePool.Put(slice)
	}
}

// getStringInterfaceMap gets a map[string]interface{} from pool
func getStringInterfaceMap(length int) map[string]interface{} {
	m := stringInterfaceMapPool.Get().(map[string]interface{})
	// Clear existing entries
	for k := range m {
		delete(m, k)
	}
	return m
}

// putStringInterfaceMap returns a map[string]interface{} to the pool
func putStringInterfaceMap(m map[string]interface{}) {
	if len(m) <= 32 { // Only pool reasonably sized maps
		stringInterfaceMapPool.Put(m)
	}
}

// getByteBuffer gets a []byte buffer from the appropriate pool based on size
func getByteBuffer(size int) []byte {
	switch {
	case size <= 64:
		buf := smallBytePool.Get().([]byte)
		if cap(buf) < size {
			return make([]byte, 0, size)
		}
		return buf[:0]
	case size <= 256:
		buf := mediumBytePool.Get().([]byte)
		if cap(buf) < size {
			return make([]byte, 0, size)
		}
		return buf[:0]
	case size <= 1024:
		buf := largeBytePool.Get().([]byte)
		if cap(buf) < size {
			return make([]byte, 0, size)
		}
		return buf[:0]
	default:
		return make([]byte, 0, size)
	}
}

// putByteBuffer returns a []byte buffer to the appropriate pool
func putByteBuffer(buf []byte) {
	capacity := cap(buf)
	switch {
	case capacity <= 64:
		buf = buf[:0]
		smallBytePool.Put(buf)
	case capacity <= 256:
		buf = buf[:0]
		mediumBytePool.Put(buf)
	case capacity <= 1024:
		buf = buf[:0]
		largeBytePool.Put(buf)
	}
	// Larger buffers are not pooled
}

// getWriterBuffer gets a []byte buffer for writers from pool
func getWriterBuffer() []byte {
	buf := writerBufferPool.Get().([]byte)
	return buf[:0] // Reset length but keep capacity
}

// putWriterBuffer returns a writer buffer to the pool
func putWriterBuffer(buf []byte) {
	if cap(buf) <= 8192 { // Only pool reasonably sized buffers
		buf = buf[:0]
		writerBufferPool.Put(buf)
	}
}

// getEncoder gets an encoder from the pool
func getEncoder(w Writer) *Encoder {
	e := encoderPool.Get().(*Encoder)
	e.writer = w
	return e
}

// putEncoder returns an encoder to the pool
func putEncoder(e *Encoder) {
	e.writer = nil // Clear reference
	encoderPool.Put(e)
}

// getDecoder gets a decoder from the pool  
func getDecoder(r Reader) *Decoder {
	d := decoderPool.Get().(*Decoder)
	d.reader = r
	return d
}

// putDecoder returns a decoder to the pool
func putDecoder(d *Decoder) {
	d.reader = nil // Clear reference
	decoderPool.Put(d)
}
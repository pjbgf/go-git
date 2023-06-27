package sync

import (
	"bytes"
	"sync"
)

var (
	byteSlice = sync.Pool{
		New: func() interface{} {
			b := make([]byte, 16*1024)
			return &b
		},
	}
	bytesBuffer = sync.Pool{
		New: func() interface{} {
			return bytes.NewBuffer(nil)
		},
	}
	smallByteSlice = sync.Pool{
		New: func() interface{} {
			b := make([]byte, 1024)
			return &b
		},
	}
)

// GetByteSlice returns a *[]byte that is managed by a sync.Pool.
// The initial slice length will be 16384 (16kb).
//
// After use, the *[]byte should be put back into the sync.Pool
// by calling PutByteSlice.
func GetByteSlice() *[]byte {
	buf := byteSlice.Get().(*[]byte)
	return buf
}

// PutByteSlice puts buf back into its sync.Pool.
func PutByteSlice(buf *[]byte) {
	byteSlice.Put(buf)
}

// SmallByteSlice returns a small (1024 bytes) slice
// which is managed by sync.Pool. The second return
// value is a func that puts the slice back into the
// pool after use.
func SmallByteSlice() (*[]byte, func()) {
	buf := smallByteSlice.Get().(*[]byte)
	return buf, func() {
		smallByteSlice.Put(buf)
	}
}

// GetBytesBuffer returns a *bytes.Buffer that is managed by a sync.Pool.
// Returns a buffer that is resetted and ready for use.
//
// After use, the *bytes.Buffer should be put back into the sync.Pool
// by calling PutBytesBuffer.
func GetBytesBuffer() *bytes.Buffer {
	buf := bytesBuffer.Get().(*bytes.Buffer)
	buf.Reset()
	return buf
}

// PutBytesBuffer puts buf back into its sync.Pool.
func PutBytesBuffer(buf *bytes.Buffer) {
	bytesBuffer.Put(buf)
}

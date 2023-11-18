package idxfile_test

// import (
// 	"bytes"
// 	"io"
// 	"testing"

// 	format "github.com/go-git/go-git/v5/plumbing/format/config"
// 	"github.com/go-git/go-git/v5/plumbing/format/idxfile"
// 	. "github.com/go-git/go-git/v5/plumbing/format/idxfile"
// 	"github.com/stretchr/testify/assert"

// 	fixtures "github.com/go-git/go-git-fixtures/v4"
// )

// func TestEncoderNilWriter(t *testing.T) {
// 	_, err := NewEncoder(nil)
// 	assert.ErrorContains(t, err, "nil writer")
// }

// func TestEncoderNilIndex(t *testing.T) {
// 	buf := bytes.NewBuffer(nil)
// 	d, err := NewEncoder(buf)
// 	assert.NoError(t, err)
// 	_, err = d.Encode(nil)
// 	assert.ErrorContains(t, err, "target index is nil")
// }

// func TestDecodeEncode(t *testing.T) {
// 	fixtures.ByTag("packfile").Run(t, func(t *testing.T, f *fixtures.Fixture) {
// 		expected, err := io.ReadAll(f.Idx())
// 		assert.NoError(t, err)

// 		idx := NewMemoryIndex(format.SHA1)
// 		d, err := NewDecoder(bytes.NewBuffer(expected))
// 		assert.NoError(t, err)

// 		err = d.Decode(idx)
// 		assert.NoError(t, err)

// 		result := bytes.NewBuffer(nil)
// 		e, err := NewEncoder(result)
// 		assert.NoError(t, err)
// 		size, err := e.Encode(idx)
// 		assert.NoError(t, err)
// 		assert.Len(t, expected, size)
// 		assert.Equal(t, expected, result.Bytes())
// 	})
// }

// func TestDecodeEncodeSHA256(t *testing.T) {
// 	fixtures.ByTag("packfile-sha256").Run(t, func(t *testing.T, f *fixtures.Fixture) {
// 		expected, err := io.ReadAll(f.Idx())
// 		assert.NoError(t, err)

// 		idx := NewMemoryIndex(format.SHA256)
// 		d, err := NewDecoderWithOptions(bytes.NewBuffer(expected), DecoderOptions{ObjectFormat: format.SHA256})
// 		assert.NoError(t, err)
// 		err = d.Decode(idx)
// 		assert.NoError(t, err)

// 		result := bytes.NewBuffer(nil)
// 		e, err := NewEncoderWithOptions(result, EncoderOptions{ObjectFormat: format.SHA256, Version: Version2})
// 		assert.NoError(t, err)
// 		size, err := e.Encode(idx)
// 		assert.NoError(t, err)
// 		assert.Len(t, expected, size)
// 		assert.Equal(t, expected, result.Bytes())
// 	})
// }

// func BenchmarkEncode(b *testing.B) {
// 	defer fixtures.Clean()

// 	b.Run("sha1-legacy", func(b *testing.B) {
// 		benchmarkEncodeLegacy(b, "packfile")
// 	})
// 	// Note that both uses the same source pack file so that
// 	// the benchmarks are comparing the same encoding operation.
// 	b.Run("sha1", func(b *testing.B) {
// 		benchmarkEncode(b, "packfile", format.SHA1)
// 	})
// 	b.Run("sha256", func(b *testing.B) {
// 		benchmarkEncode(b, "packfile", format.SHA256)
// 	})
// }

// func benchmarkEncode(b *testing.B, tag string, objectFormat format.ObjectFormat) {
// 	f := fixtures.ByTag(tag).One()
// 	d, err := NewDecoder(f.Idx())
// 	assert.NoError(b, err)

// 	idx := NewMemoryIndex(objectFormat)
// 	err = d.Decode(idx)
// 	assert.NoError(b, err)

// 	b.ResetTimer()
// 	for i := 0; i < b.N; i++ {
// 		d, _ := NewEncoderWithOptions(io.Discard, EncoderOptions{ObjectFormat: objectFormat})
// 		_, _ = d.Encode(idx)
// 	}
// }

// func benchmarkEncodeLegacy(b *testing.B, tag string) {
// 	f := fixtures.ByTag(tag).One()
// 	d := idxfile.NewDecoder(f.Idx())
// 	idx := new(idxfile.MemoryIndex)
// 	err := d.Decode(idx)
// 	assert.NoError(b, err)

// 	b.ResetTimer()
// 	for i := 0; i < b.N; i++ {
// 		d := idxfile.NewEncoder(io.Discard)
// 		_, _ = d.Encode(idx)
// 	}
// }

package idxfile_test

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"testing"

	"github.com/go-git/go-git/v5/plumbing/format/config"
	format "github.com/go-git/go-git/v5/plumbing/format/config"
	"github.com/go-git/go-git/v5/plumbing/format/idxfile"
	. "github.com/go-git/go-git/v5/plumbing/format/idxfile/v3"
	"github.com/go-git/go-git/v5/plumbing/hasher"
	"github.com/stretchr/testify/assert"

	fixtures "github.com/go-git/go-git-fixtures/v4"
)

func TestDecodeNilReader(t *testing.T) {
	_, err := NewDecoder(nil)
	assert.ErrorContains(t, err, "nil reader")
}

func TestDecodeNilIndex(t *testing.T) {
	f := fixtures.Basic().One()

	d, err := NewDecoder(f.Idx())
	assert.NoError(t, err)

	err = d.Decode(nil)
	assert.ErrorContains(t, err, "target index is nil")
}

func TestDecode(t *testing.T) {
	f := fixtures.Basic().One()

	d, err := NewDecoder(f.Idx())
	assert.NoError(t, err)

	idx := NewMemoryIndex(format.SHA1)
	err = d.Decode(idx)
	assert.NoError(t, err)

	count, err := idx.Count()
	assert.NoError(t, err)
	assert.Equal(t, int64(31), count)

	hash, ok := hasher.Parse("1669dce138d9b841a518c64b10914d88f5e488ea")
	assert.True(t, ok)
	ok, err = idx.Contains(hash)
	assert.NoError(t, err)
	assert.True(t, ok)

	offset, err := idx.FindOffset(hash)
	assert.NoError(t, err)
	assert.Equal(t, int64(615), offset)

	crc32, err := idx.FindCRC32(hash)
	assert.NoError(t, err)
	assert.Equal(t, uint32(3645019190), crc32)

	assert.Equal(t, "fb794f1ec720b9bc8e43257451bd99c4be6fa1c9", fmt.Sprintf("%x", idx.IdxChecksum))
	assert.Equal(t, f.PackfileHash, fmt.Sprintf("%x", idx.PackfileChecksum))
}

func TestDecodeSHA256(t *testing.T) {
	f := fixtures.ByTag("packfile-sha256").One()

	d, err := NewDecoderWithOptions(f.Idx(), DecoderOptions{config.SHA256})
	assert.NoError(t, err)
	idx := NewMemoryIndex(format.SHA256)
	err = d.Decode(idx)
	assert.NoError(t, err)

	count, err := idx.Count()
	assert.NoError(t, err)
	assert.Equal(t, int64(6), count)

	hash, ok := hasher.Parse("0d8d657df872bef9d0684fe4bc4ee3a088b6f0f72d64f951daff9465068905ac")
	assert.True(t, ok)
	ok, err = idx.Contains(hash)
	assert.NoError(t, err)
	assert.True(t, ok)

	offset, err := idx.FindOffset(hash)
	assert.NoError(t, err)
	assert.Equal(t, int64(459), offset)

	crc32, err := idx.FindCRC32(hash)
	assert.NoError(t, err)
	assert.Equal(t, uint32(2212914800), crc32)

	assert.Equal(t, "c5bc22dd894603a3a0d498bb072f1391fd0c0452a18290cc8caa32b9842d876b", fmt.Sprintf("%x", idx.IdxChecksum))
	assert.Equal(t, f.PackfileHash, fmt.Sprintf("%x", idx.PackfileChecksum))
}

func TestDecode64bitsOffsets(t *testing.T) {
	f := bytes.NewBufferString(fixtureLarge4GB)

	idx := NewMemoryIndex(format.SHA1)

	d, err := NewDecoder(base64.NewDecoder(base64.StdEncoding, f))
	assert.NoError(t, err)
	err = d.Decode(idx)
	assert.NoError(t, err)

	expected := map[string]uint64{
		"303953e5aa461c203a324821bc1717f9b4fff895": 12,
		"5296768e3d9f661387ccbff18c4dea6c997fd78c": 142,
		"03fc8d58d44267274edef4585eaeeb445879d33f": 1601322837,
		"8f3ceb4ea4cb9e4a0f751795eb41c9a4f07be772": 2646996529,
		"e0d1d625010087f79c9e01ad9d8f95e1628dda02": 3452385606,
		"90eba326cdc4d1d61c5ad25224ccbf08731dd041": 3707047470,
		"bab53055add7bc35882758a922c54a874d6b1272": 5323223332,
		"1b8995f51987d8a449ca5ea4356595102dc2fbd4": 5894072943,
		"35858be9c6f5914cbe6768489c41eb6809a2bceb": 5924278919,
	}

	iter, err := idx.Entries()
	assert.NoError(t, err)

	var entries int
	for {
		e, err := iter.Next()
		if err == io.EOF {
			break
		}
		assert.NoError(t, err)
		entries++

		assert.Equal(t, e.Offset, expected[e.Hash.String()])
	}

	assert.Len(t, expected, entries)
}

const fixtureLarge4GB = `/3RPYwAAAAIAAAAAAAAAAAAAAAAAAAABAAAAAQAAAAEAAAABAAAAAQAAAAEAAAABAAAAAQAAAAEA
AAABAAAAAQAAAAEAAAABAAAAAQAAAAEAAAABAAAAAQAAAAEAAAABAAAAAQAAAAEAAAABAAAAAQAA
AAEAAAACAAAAAgAAAAIAAAACAAAAAgAAAAIAAAACAAAAAgAAAAIAAAACAAAAAgAAAAIAAAACAAAA
AgAAAAIAAAACAAAAAgAAAAIAAAACAAAAAgAAAAIAAAADAAAAAwAAAAMAAAADAAAAAwAAAAQAAAAE
AAAABAAAAAQAAAAEAAAABAAAAAQAAAAEAAAABAAAAAQAAAAEAAAABAAAAAQAAAAEAAAABAAAAAQA
AAAEAAAABAAAAAQAAAAEAAAABAAAAAQAAAAEAAAABAAAAAQAAAAEAAAABAAAAAQAAAAEAAAABQAA
AAUAAAAFAAAABQAAAAUAAAAFAAAABQAAAAUAAAAFAAAABQAAAAUAAAAFAAAABQAAAAUAAAAFAAAA
BQAAAAUAAAAFAAAABQAAAAUAAAAFAAAABQAAAAUAAAAFAAAABQAAAAUAAAAFAAAABQAAAAUAAAAF
AAAABQAAAAUAAAAFAAAABQAAAAUAAAAFAAAABQAAAAUAAAAFAAAABQAAAAUAAAAFAAAABQAAAAUA
AAAFAAAABQAAAAUAAAAFAAAABQAAAAUAAAAFAAAABQAAAAUAAAAFAAAABQAAAAUAAAAFAAAABQAA
AAUAAAAFAAAABQAAAAYAAAAHAAAABwAAAAcAAAAHAAAABwAAAAcAAAAHAAAABwAAAAcAAAAHAAAA
BwAAAAcAAAAHAAAABwAAAAcAAAAHAAAABwAAAAcAAAAHAAAABwAAAAcAAAAHAAAABwAAAAcAAAAH
AAAABwAAAAcAAAAHAAAABwAAAAcAAAAHAAAABwAAAAcAAAAHAAAABwAAAAcAAAAHAAAABwAAAAcA
AAAHAAAABwAAAAcAAAAIAAAACAAAAAgAAAAIAAAACAAAAAgAAAAIAAAACAAAAAgAAAAIAAAACAAA
AAgAAAAIAAAACAAAAAgAAAAIAAAACAAAAAgAAAAIAAAACAAAAAgAAAAIAAAACAAAAAgAAAAIAAAA
CAAAAAgAAAAIAAAACAAAAAgAAAAIAAAACAAAAAgAAAAIAAAACAAAAAgAAAAIAAAACAAAAAkAAAAJ
AAAACQAAAAkAAAAJAAAACQAAAAkAAAAJAAAACQAAAAkAAAAJAAAACQAAAAkAAAAJAAAACQAAAAkA
AAAJAAAACQAAAAkAAAAJAAAACQAAAAkAAAAJAAAACQAAAAkAAAAJAAAACQAAAAkAAAAJAAAACQAA
AAkAAAAJA/yNWNRCZydO3vRYXq7rRFh50z8biZX1GYfYpEnKXqQ1ZZUQLcL71DA5U+WqRhwgOjJI
IbwXF/m0//iVNYWL6cb1kUy+Z2hInEHraAmivOtSlnaOPZ9mE4fMv/GMTepsmX/XjI88606ky55K
D3UXletByaTwe+dykOujJs3E0dYcWtJSJMy/CHMd0EG6tTBVrde8NYgnWKkixUqHTWsScuDR1iUB
AIf3nJ4BrZ2PleFijdoCkp36qiGHwFa8NHxMnInZ0s3CKEKmHe+KcZPzuqwmm44GvqGAX3I/VYAA
AAAAAAAMgAAAAQAAAI6AAAACgAAAA4AAAASAAAAFAAAAAV9Qam8AAAABYR1ShwAAAACdxfYxAAAA
ANz1Di4AAAABPUnxJAAAAADNxzlGr6vCJpIFz4XaG/fi/f9C9zgQ8ptKSQpfQ1NMJBGTDTxxYGGp
ch2xUA==
`

func BenchmarkDecode(b *testing.B) {
	defer fixtures.Clean()

	fixture := fixtures.Basic().One().Idx()
	b.Run("sha1-legacy", func(b *testing.B) {
		benchmarkDecodeLegacy(b, fixture)
	})
	b.Run("sha1", func(b *testing.B) {
		benchmarkDecode(b, fixture, format.SHA1)
	})
	fixture256 := fixtures.ByTag("packfile-sha256").One().Idx()
	b.Run("sha256", func(b *testing.B) {
		benchmarkDecode(b, fixture256, format.SHA256)
	})

	largeBytes := make([]byte, 1372)
	base64.RawStdEncoding.Decode(largeBytes, []byte(fixtureLarge4GB))

	largeFixture := bytes.NewReader(largeBytes)
	b.Run("large-sha1-legacy", func(b *testing.B) {
		benchmarkDecodeLegacy(b, largeFixture)
	})
	b.Run("large-sha1", func(b *testing.B) {
		benchmarkDecode(b, largeFixture, format.SHA1)
	})
}

func benchmarkDecode(b *testing.B, fixture io.ReadSeeker, f format.ObjectFormat) {
	d, _ := NewDecoderWithOptions(fixture, DecoderOptions{ObjectFormat: f})
	idx := NewMemoryIndex(f)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := fixture.Seek(0, 0)
		if err != nil {
			b.Fatalf("failed seeking: %v", err)
		}

		err = d.Decode(idx)
		if err != nil {
			b.Fatalf("failed to decode: %v", err)
		}
	}
}

func benchmarkDecodeLegacy(b *testing.B, fixture io.ReadSeeker) {
	d := idxfile.NewDecoder(fixture)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Legacy idxfile cannot reuse memory indexes.
		idx := idxfile.NewMemoryIndex()

		_, err := fixture.Seek(0, 0)
		if err != nil {
			b.Fatalf("failed seeking: %v", err)
		}

		err = d.Decode(idx)
		if err != nil {
			b.Fatalf("failed to decode: %v", err)
		}
	}
}

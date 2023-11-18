package idxfile_test

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"testing"

	. "github.com/go-git/go-git/v5/plumbing/format/idxfile"
	"github.com/go-git/go-git/v5/plumbing/hash/common"
	"github.com/go-git/go-git/v5/plumbing/hash/sha1"
	"github.com/stretchr/testify/assert"
)

func BenchmarkFindOffset(b *testing.B) {
	idx, err := fixtureIndex()
	if err != nil {
		b.Fatalf(err.Error())
	}

	for i := 0; i < b.N; i++ {
		for _, h := range fixtureHashes {
			_, err := idx.FindOffset(h)
			if err != nil {
				b.Fatalf("error getting offset: %s", err)
			}
		}
	}
}

func BenchmarkFindCRC32(b *testing.B) {
	idx, err := fixtureIndex()
	if err != nil {
		b.Fatalf(err.Error())
	}

	for i := 0; i < b.N; i++ {
		for _, h := range fixtureHashes {
			_, err := idx.FindCRC32(h)
			if err != nil {
				b.Fatalf("error getting crc32: %s", err)
			}
		}
	}
}

func BenchmarkContains(b *testing.B) {
	idx, err := fixtureIndex()
	if err != nil {
		b.Fatalf(err.Error())
	}

	for i := 0; i < b.N; i++ {
		for _, h := range fixtureHashes {
			ok, err := idx.Contains(h)
			if err != nil {
				b.Fatalf("error checking if hash is in index: %s", err)
			}

			if !ok {
				b.Error("expected hash to be in index")
			}
		}
	}
}

func BenchmarkEntries(b *testing.B) {
	idx, err := fixtureIndex()
	if err != nil {
		b.Fatalf(err.Error())
	}

	for i := 0; i < b.N; i++ {
		iter, err := idx.Entries()
		if err != nil {
			b.Fatalf("unexpected error getting entries: %s", err)
		}

		var entries int
		for {
			_, err := iter.Next()
			if err != nil {
				if err == io.EOF {
					break
				}

				b.Errorf("unexpected error getting entry: %s", err)
			}

			entries++
		}

		if entries != len(fixtureHashes) {
			b.Errorf("expecting entries to be %d, got %d", len(fixtureHashes), entries)
		}
	}
}

func TestFindHash(t *testing.T) {
	idx, err := fixtureIndex()
	assert.NoError(t, err)

	for i, pos := range fixtureOffsets {
		hash, err := idx.FindHash(pos)
		assert.NoError(t, err)
		assert.Equal(t, fixtureHashes[i], hash)
	}
}

func TestEntriesByOffset(t *testing.T) {
	idx, err := fixtureIndex()
	assert.NoError(t, err)

	entries, err := idx.EntriesByOffset()
	assert.NoError(t, err)

	for _, pos := range fixtureOffsets {
		e, err := entries.Next()
		assert.NoError(t, err)

		assert.Equal(t, uint64(pos), e.Offset)
	}
}

var x = func(h common.ObjectHash, _ bool) common.ObjectHash {
	return h
}

var fixtureHashes = []common.ObjectHash{
	x(sha1.FromHex("303953e5aa461c203a324821bc1717f9b4fff895")),
	x(sha1.FromHex("5296768e3d9f661387ccbff18c4dea6c997fd78c")),
	x(sha1.FromHex("03fc8d58d44267274edef4585eaeeb445879d33f")),
	x(sha1.FromHex("8f3ceb4ea4cb9e4a0f751795eb41c9a4f07be772")),
	x(sha1.FromHex("e0d1d625010087f79c9e01ad9d8f95e1628dda02")),
	x(sha1.FromHex("90eba326cdc4d1d61c5ad25224ccbf08731dd041")),
	x(sha1.FromHex("bab53055add7bc35882758a922c54a874d6b1272")),
	x(sha1.FromHex("1b8995f51987d8a449ca5ea4356595102dc2fbd4")),
	x(sha1.FromHex("35858be9c6f5914cbe6768489c41eb6809a2bceb")),
}

var fixtureOffsets = []int64{
	12,
	142,
	1601322837,
	2646996529,
	3452385606,
	3707047470,
	5323223332,
	5894072943,
	5924278919,
}

func fixtureIndex() (*MemoryIndex, error) {
	f := bytes.NewBufferString(fixtureLarge4GB)

	idx := NewMemoryIndex(sha1.Factory)

	d, err := NewDecoder(base64.NewDecoder(base64.StdEncoding, f), sha1.Factory)
	if err != nil {
		return nil, fmt.Errorf("failed to create decoder: %s", err)
	}

	err = d.Decode(idx)
	if err != nil {
		return nil, fmt.Errorf("unexpected error decoding index: %s", err)
	}

	return idx, nil
}

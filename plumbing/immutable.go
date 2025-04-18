package plumbing

import (
	"bytes"
	"encoding/hex"

	format "github.com/go-git/go-git/v6/plumbing/format/config"
	"github.com/go-git/go-git/v6/plumbing/hash"
)

// ImmutableHash represents a calculated hash.
type ImmutableHash interface {
	// Size returns the length of the resulting hash.
	Size() int
	// Empty returns true if the hash is zero.
	Empty() bool
	// Compare compares the hash's sum with a slice of bytes.
	Compare([]byte) int
	// String returns the hexadecimal representation of the hash's sum.
	String() string
	// Sum returns the slice of bytes containing the hash.
	Bytes() []byte
	HasPrefix([]byte) bool

	IsZero() bool
}

// FromHex parses a hexadecimal string and returns an ImmutableHash
// and a boolean confirming whether the operation was successful.
// The hash (and object format) is inferred from the length of the
// input.
//
// If the operation was not successful, the resulting hash is nil
// instead of a zeroed hash.
func FromHex(in string) (ImmutableHash, bool) {
	if len(in) < hash.SHA1HexSize ||
		len(in) > hash.SHA256HexSize {
		return nil, false
	}

	h := StringHash{}
	h.WriteHex(in)
	return h, true

	// switch len(in) {
	// case hash.SHA1HexSize:
	// 	h := SHA1Hash{}
	// 	copy(h[:], b)
	// 	return h, true

	// case hash.SHA256HexSize:
	// 	h := SHA256Hash{}
	// 	copy(h[:], b)
	// 	return h, true

	// default:
	// 	return nil, false
	// }
}

// FromBytes creates an ImmutableHash object based on the value its
// Sum() should return.
// The hash type (and object format) is inferred from the length of
// the input.
//
// If the operation was not successful, the resulting hash is nil
// instead of a zeroed hash.
func FromBytes(in []byte) (ImmutableHash, bool) {
	if len(in) < hash.SHA1Size ||
		len(in) > hash.SHA256Size {
		return nil, false
	}

	switch len(in) {
	case hash.SHA1Size:
		h := SHA1Hash{}
		copy(h[:], in)
		return h, true

	case hash.SHA256Size:
		h := SHA256Hash{}
		copy(h[:], in)
		return h, true

	default:
		return nil, false
	}
}

// ZeroFromHash returns a zeroed hash based on the given hash.Hash.
//
// Defaults to SHA1-sized hash if the provided hash is not supported.
func ZeroFromHash(h hash.Hash) ImmutableHash {
	switch h.Size() {
	case hash.SHA256Size:
		return SHA256Hash{}
	default:
		return SHA1Hash{}
	}
}

// ZeroFromHash returns a zeroed hash based on the given ObjectFormat.
//
// Defaults to SHA1-sized hash if the provided format is not supported.
func ZeroFromObjectFormat(f format.ObjectFormat) ImmutableHash {
	switch f {
	case format.SHA256:
		return SHA256Hash{}
	default:
		return SHA1Hash{}
	}
}

type SHA1Hash [hash.SHA1Size]byte

func (ih SHA1Hash) Size() int {
	return len(ih)
}

func (ih SHA1Hash) Empty() bool {
	var empty SHA1Hash
	return ih == empty
}

func (ih SHA1Hash) IsZero() bool {
	return ih.Empty()
}

func (ih SHA1Hash) String() string {
	return hex.EncodeToString(ih[:])
}

func (ih SHA1Hash) Bytes() []byte {
	return ih[:]
}

func (ih SHA1Hash) Compare(in []byte) int {
	return bytes.Compare(ih[:], in)
}

func (ih SHA1Hash) HasPrefix(prefix []byte) bool {
	return bytes.HasPrefix(ih[:], prefix)
}

type SHA256Hash [hash.SHA256Size]byte

func (ih SHA256Hash) Size() int {
	return len(ih)
}

func (ih SHA256Hash) Empty() bool {
	var empty SHA256Hash
	return ih == empty
}
func (ih SHA256Hash) IsZero() bool {
	return ih.Empty()
}

func (ih SHA256Hash) String() string {
	return hex.EncodeToString(ih[:])
}

func (ih SHA256Hash) Bytes() []byte {
	return ih[:]
}

func (ih SHA256Hash) Compare(in []byte) int {
	return bytes.Compare(ih[:], in)
}

func (ih SHA256Hash) HasPrefix(prefix []byte) bool {
	return bytes.HasPrefix(ih[:], prefix)
}

// ImmutableHashesSort sorts a slice of ImmutableHashes in increasing order.
// func ImmutableHashesSort(a []ImmutableHash) {
// 	sort.Sort(HashSlice(a))
// }

// // HashSlice attaches the methods of sort.Interface to []Hash, sorting in
// // increasing order.
// type HashSlice []ImmutableHash

// func (p HashSlice) Len() int           { return len(p) }
// func (p HashSlice) Less(i, j int) bool { return p[i].Compare(p[j].Sum()) <= 0 }
// func (p HashSlice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

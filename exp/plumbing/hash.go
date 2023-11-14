package plumbing

import (
	"bytes"
	"encoding/hex"

	format "github.com/go-git/go-git/v5/plumbing/format/config"
	"github.com/go-git/go-git/v5/plumbing/hash"
)

// Hash represents a calculated hash.
type Hash interface {
	// Size returns the length of the resulting hash.
	Size() int
	// IsZero returns true if the hash is zero.
	IsZero() bool
	// Compare compares the hash's sum with a slice of bytes.
	Compare([]byte) int
	// String returns the hexadecimal representation of the hash's sum.
	String() string
	// Sum returns the slice of bytes containing the hash.
	Sum() []byte
}

// Deprecated: Use FromHex instead.
// Note that the majority of the calls to NewHash are from test code.
func NewHash(in string) Hash {
	h, _ := FromHex(in)
	return h
}

// FromHex parses a hexadecimal string and returns an Hash
// and a boolean confirming whether the operation was successful.
// The hash (and object format) is inferred from the length of the
// input.
//
// Partial hashes will be handled as SHA1.
func FromHex(in string) (Hash, bool) {
	switch len(in) {
	case hash.SHA256HexSize:
		return SHA256HashFromHex(in)
	default:
		return SHA1HashFromHex(in)
	}
}

// FromBytes creates an Hash object based on the value its
// Sum() should return.
// The hash type (and object format) is inferred from the length of
// the input.
//
// Partial hashes will be handled as SHA1, and only the initial 20
// bytes will be copied into the resulting hash.
func FromBytes(in []byte) Hash {
	switch len(in) {
	case hash.SHA256Size:
		return SHA256HashFromBytes(in)
	default:
		return SHA1HashFromBytes(in)
	}
}

// ZeroFromHash returns a zeroed hash based on the given hash.Hash.
//
// Defaults to SHA1-sized hash if the provided hash is not supported.
func ZeroFromHash(h hash.Hash) Hash {
	switch h.Size() {
	case hash.SHA256Size:
		return ZeroHashSHA256
	default:
		return ZeroHashSHA1
	}
}

// ZeroFromObjectFormat returns a zeroed hash based on the given ObjectFormat.
//
// Defaults to SHA1-sized hash if the provided format is not supported.
func ZeroFromObjectFormat(f format.ObjectFormat) Hash {
	switch f {
	case format.SHA256:
		return ZeroHashSHA256
	default:
		return ZeroHashSHA1
	}
}

var (
	ZeroHashSHA1   = SHA1Hash{}
	ZeroHashSHA256 = SHA256Hash{}
)

type SHA1Hash [hash.SHA1Size]byte

func SHA1HashFromHex(in string) (SHA1Hash, bool) {
	if len(in) > hash.SHA1HexSize {
		return ZeroHashSHA1, false
	}

	b, err := hex.DecodeString(in)
	if err != nil {
		return ZeroHashSHA1, false
	}

	h := SHA1Hash{}
	copy(h[:], b)
	return h, true
}

func SHA1HashFromBytes(b []byte) SHA1Hash {
	l := len(b)
	if l > hash.SHA1Size {
		l = hash.SHA1Size
	}

	h := SHA1Hash{}
	copy(h[:], b[:l])
	return h
}

func (ih SHA1Hash) Size() int {
	return len(ih)
}

func (ih SHA1Hash) IsZero() bool {
	var empty SHA1Hash
	return ih == empty
}

func (ih SHA1Hash) String() string {
	return hex.EncodeToString(ih[:])
}

func (ih SHA1Hash) Sum() []byte {
	return ih[:]
}

func (ih SHA1Hash) Compare(in []byte) int {
	return bytes.Compare(ih[:], in)
}

type SHA256Hash [hash.SHA256Size]byte

func SHA256HashFromHex(in string) (SHA256Hash, bool) {
	if len(in) > hash.SHA256HexSize {
		return ZeroHashSHA256, false
	}

	b, err := hex.DecodeString(in)
	if err != nil {
		return ZeroHashSHA256, false
	}

	h := SHA256Hash{}
	copy(h[:], b)
	return h, true
}

func SHA256HashFromBytes(b []byte) SHA256Hash {
	l := len(b)
	if len(b) > hash.SHA256Size {
		l = hash.SHA256Size
	}

	h := SHA256Hash{}
	copy(h[:], b[:l])
	return h
}

func (ih SHA256Hash) Size() int {
	return len(ih)
}

func (ih SHA256Hash) IsZero() bool {
	var empty SHA256Hash
	return ih == empty
}

func (ih SHA256Hash) String() string {
	return hex.EncodeToString(ih[:])
}

func (ih SHA256Hash) Sum() []byte {
	return ih[:]
}

func (ih SHA256Hash) Compare(in []byte) int {
	return bytes.Compare(ih[:], in)
}

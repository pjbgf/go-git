package common

import (
	"hash"

	format "github.com/go-git/go-git/v5/plumbing/format/config"
)

// TODO: Move to internal?
type SupportedHash struct {
	format.ObjectFormat
	HashFactory
	NewHasher func() hash.Hash
}

type HashFactory interface {
	ZeroHash() ObjectHash
	FromHex(in string) (ObjectHash, bool)
	FromBytes(b []byte) ObjectHash
	Size() int
	HexSize() int
}

// ObjectHash represents a calculated hash.
type ObjectHash interface {
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

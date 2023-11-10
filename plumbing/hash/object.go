package hash

import "io"

// ObjectID represents a calculated hash.
type ObjectID interface {
	// Size returns the length of the resulting hash.
	Size() int
	// Empty returns true if the hash is zero.
	Empty() bool
	// Compare compares the hash's sum with a slice of bytes.
	Compare([]byte) int
	// String returns the hexadecimal representation of the hash's sum.
	String() string
	// Sum returns the slice of bytes containing the hash.
	Sum() []byte

	HasPrefix([]byte) bool

	// deprecated use Empty instead.
	IsZero() bool
}

// Dynamic
// -> List/Array of objects for to be calculated hash
// Static/Readonly

type StaticHash interface {
	ObjectID
}

type DynamicHash interface {
	StaticHash
	io.Writer
}

package plumbing

import (
	"encoding/hex"
	"strconv"

	"github.com/go-git/go-git/v5/plumbing/hash"
)

// Hash SHA1 hashed content
type Hash [hash.Size]byte

// ZeroHash is Hash with value zero
var ZeroHash Hash

func (h Hash) IsZero() bool {
	var empty Hash
	return h == empty
}

func (h Hash) String() string {
	return hex.EncodeToString(h[:])
}

func (h Hash) Zero() []byte {
	return ZeroHash.Bytes()
}

func (h Hash) Bytes() []byte {
	return h[:]
}

func (h Hash) Size() int {
	return hash.Size
}

func (h Hash) Write(data []byte) (n int, err error) {
	n = copy(h[:], data)
	return
}

// TODO: maybe not
func (_ Hash) Parse(s string) Hash {
	b, _ := hex.DecodeString(s)

	var h Hash
	copy(h[:], b)

	return h
}

// ComputeHash compute the hash for a given ObjectType and content
func ComputeHash(t ObjectType, content []byte) Hash {
	h := NewHasher(t, int64(len(content)))
	h.Write(content)
	return h.Sum()
}

// NewHash return a new Hash from a hexadecimal hash representation
func NewHash(s string) Hash {
	b, _ := hex.DecodeString(s)

	var h Hash
	copy(h[:], b)

	return h
}

type Hasher struct {
	hash.Hash
}

func NewHasher(t ObjectType, size int64) Hasher {
	h := Hasher{hash.New(hash.CryptoType)}
	h.Write(t.Bytes())
	h.Write([]byte(" "))
	h.Write([]byte(strconv.FormatInt(size, 10)))
	h.Write([]byte{0})
	return h
}

func (h Hasher) Sum() (hash Hash) {
	copy(hash[:], h.Hash.Sum(nil))
	return
}

// IsHash returns true if the given string is a valid hash.
func IsHash(s string) bool {
	switch len(s) {
	case hash.HexSize:
		_, err := hex.DecodeString(s)
		return err == nil
	default:
		return false
	}
}

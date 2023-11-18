package plumbing

import (
	"encoding/hex"
	"sort"
	"strconv"

	"github.com/go-git/go-git/v5/plumbing/format/config"
	"github.com/go-git/go-git/v5/plumbing/hash"
	"github.com/go-git/go-git/v5/plumbing/hash/common"
	"github.com/go-git/go-git/v5/plumbing/hash/sha1"
)

// Hash SHA1 hashed content
// Deprecated use common.ObjectHash instead.
type Hash sha1.SHA1Hash

// ZeroHash is Hash with value zero
// Deprecated use sha1.ZeroHash() or sha256.ZeroHash() instead.
var ZeroHash Hash

func (ih Hash) Size() int {
	return len(ih)
}

func (ih Hash) IsZero() bool {
	return sha1.SHA1Hash(ih).IsZero()
}

func (ih Hash) String() string {
	return sha1.SHA1Hash(ih).String()
}

func (ih Hash) Sum() []byte {
	return sha1.SHA1Hash(ih).Sum()
}

func (ih Hash) Compare(in []byte) int {
	return sha1.SHA1Hash(ih).Compare(in)
}

// ComputeHash compute the hash for a given ObjectType and content
func ComputeHash(t ObjectType, content []byte) common.ObjectHash {
	hasher, _ := FromObjectFormat(config.SHA1)
	h, _ := hasher.Compute(t, content)
	return h
}

// NewHash return a new Hash from a hexadecimal hash representation
// Deprecated. Use sha1.FromHex or sha256.FromHex instead.
func NewHash(s string) common.ObjectHash {
	h, _ := sha1.FromHex(s)
	return h
}

// Deprecated
type Hasher struct {
	hash.Hash
}

// Deprecated
func NewHasher(t ObjectType, size int64) Hasher {
	h := Hasher{hash.NewHasher(config.SHA1)}
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

// HashesSort sorts a slice of Hashes in increasing order.
func HashesSort(a []common.ObjectHash) {
	sort.Sort(HashSlice(a))
}

// HashSlice attaches the methods of sort.Interface to []Hash, sorting in
// increasing order.
type HashSlice []common.ObjectHash

func (p HashSlice) Len() int           { return len(p) }
func (p HashSlice) Less(i, j int) bool { return p[i].Compare(p[j].Sum()) < 0 }
func (p HashSlice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

// IsHash returns true if the given string is a valid hash.
// Deprecated
func IsHash(s string) bool {
	switch len(s) {
	case hash.HexSize:
		_, err := hex.DecodeString(s)
		return err == nil
	default:
		return false
	}
}

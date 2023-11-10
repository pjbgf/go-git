package hash

// import (
// 	"crypto"
// 	"fmt"
// 	"strings"
// 	"sync"

// 	"github.com/go-git/go-git/v5/plumbing/format/config"
// )

// func NewStringHashFromBytes(in []byte) StringHash {
// 	return StringHash{
// 		hash:   string(in),
// 		format: config.SHA1,
// 	}
// }

// // ImmutableHash represents a calculated hash.
// type StringHash struct {
// 	format config.ObjectFormat
// 	hash   string
// }

// // Size returns the length of the resulting hash.
// func (s StringHash) Size() int {
// 	if s.format == config.SHA256 {
// 		return SHA256Size
// 	}
// 	return SHA1Size
// }

// // Size returns the length of the resulting hash.
// func (s StringHash) HexSize() int {
// 	if s.format == config.SHA256 {
// 		return SHA256HexSize
// 	}
// 	return SHA1HexSize
// }

// // Empty returns true if the hash is zero.
// func (s StringHash) Empty() bool {
// 	return s.hash == ""
// }

// // TODO: Compare and CompareBytes
// // Compare compares the hash's sum with a slice of bytes.
// func (s StringHash) Compare(b []byte) int {
// 	return strings.Compare(s.hash, string(b))
// }

// // String returns the hexadecimal representation of the hash's sum.
// func (s StringHash) String() string {
// 	return string(s.hash)
// }

// // Sum returns the slice of bytes containing the hash.
// func (s StringHash) Sum() []byte {
// 	return []byte(s.hash)
// }

// // Sum returns the slice of bytes containing the hash.
// func (s StringHash) Bytes() []byte {
// 	return []byte(s.hash)
// }

// func (s StringHash) HasPrefix(prefix []byte) bool {
// 	return strings.HasPrefix(string(s.hash), string(prefix))
// }

// func (s StringHash) IsZero() bool {
// 	return s.Empty()
// }

// func (s *StringHash) Write(in []byte) (int, error) {
// 	s.hash = string(in)
// 	return len(in), nil
// }

// func (s *StringHash) WriteHex(in string) error {
// 	s.hash = in
// 	return nil
// }

// // ImmutableHashesSort sorts a slice of ImmutableHashes in increasing order.
// // func ImmutableHashesSort(a []ImmutableHash) {
// // 	sort.Sort(HashSlice(a))
// // }

// // // HashSlice attaches the methods of sort.Interface to []Hash, sorting in
// // // increasing order.
// // type HashSlice []ImmutableHash

// // func (p HashSlice) Len() int           { return len(p) }
// // func (p HashSlice) Less(i, j int) bool { return p[i].Compare(p[j].Sum()) <= 0 }
// // func (p HashSlice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

// func newStringHasher() *stringHasher {
// 	return &stringHasher{
// 		hasher: hash.New(crypto.SHA1),
// 	}
// }

// type stringHasher struct {
// 	hasher hash.Hash
// 	m      sync.Mutex
// }

// func (h *stringHasher) Compute(ot ObjectType, d []byte) (ImmutableHash, error) {
// 	h.m.Lock()
// 	h.hasher.Reset()

// 	writeHeader(h.hasher, ot, int64(len(d)))
// 	_, err := h.hasher.Write(d)
// 	if err != nil {
// 		h.m.Unlock()
// 		return nil, fmt.Errorf("failed to compute hash: %w", err)
// 	}

// 	out := StringHash{}
// 	out.Write(h.hasher.Sum(nil))
// 	h.m.Unlock()
// 	return out, nil
// }

// func (h *stringHasher) Size() int {
// 	return h.hasher.Size()
// }

// func (h *stringHasher) Write(p []byte) (int, error) {
// 	return h.hasher.Write(p)
// }

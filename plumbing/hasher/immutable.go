package hasher

import (
	"bytes"
	"encoding/hex"

	format "github.com/go-git/go-git/v5/plumbing/format/config"
	"github.com/go-git/go-git/v5/plumbing/hash"
)

type ImmutableHash interface {
	Size() int
	Empty() bool
	Compare([]byte) int
	String() string
	Sum() []byte
}

func Parse(in string) (ImmutableHash, bool) {
	if len(in) < hash.SHA1_HexSize ||
		len(in) > hash.SHA256_HexSize {
		return nil, false
	}

	b, err := hex.DecodeString(in)
	if err != nil {
		return nil, false
	}

	switch len(in) {
	case hash.SHA1_HexSize:
		h := immutableHashSHA1{}
		copy(h[:], b)
		return h, true

	case hash.SHA256_HexSize:
		h := immutableHashSHA256{}
		copy(h[:], b)
		return h, true

	default:
		return nil, false
	}
}

func FromBytes(in []byte) (ImmutableHash, bool) {
	if len(in) < hash.SHA1_Size ||
		len(in) > hash.SHA256_Size {
		return nil, false
	}

	switch len(in) {
	case hash.SHA1_Size:
		h := immutableHashSHA1{}
		copy(h[:], in)
		return h, true

	case hash.SHA256_Size:
		h := immutableHashSHA256{}
		copy(h[:], in)
		return h, true

	default:
		return nil, false
	}
}

func ZeroFromHash(h hash.Hash) ImmutableHash {
	switch h.Size() {
	case hash.SHA1_Size:
		return immutableHashSHA1{}
	case hash.SHA256_Size:
		return immutableHashSHA1{}
	default:
		return nil
	}
}

func ZeroFromObjectFormat(f format.ObjectFormat) ImmutableHash {
	switch f {
	case format.SHA1:
		return immutableHashSHA1{}
	case format.SHA256:
		return immutableHashSHA1{}
	default:
		return nil
	}
}

type immutableHashSHA1 [hash.SHA1_Size]byte

func (ih immutableHashSHA1) Size() int {
	return len(ih)
}

func (ih immutableHashSHA1) Empty() bool {
	var empty immutableHashSHA1
	return ih == empty
}

func (ih immutableHashSHA1) String() string {
	return hex.EncodeToString(ih[:])
}

func (ih immutableHashSHA1) Sum() []byte {
	return ih[:]
}

func (ih immutableHashSHA1) Compare(in []byte) int {
	return bytes.Compare(ih[:], in)
}

type immutableHashSHA256 [hash.SHA256_Size]byte

func (ih immutableHashSHA256) Size() int {
	return len(ih)
}

func (ih immutableHashSHA256) Empty() bool {
	var empty immutableHashSHA256
	return ih == empty
}

func (ih immutableHashSHA256) String() string {
	return hex.EncodeToString(ih[:])
}

func (ih immutableHashSHA256) Sum() []byte {
	return ih[:]
}

func (ih immutableHashSHA256) Compare(in []byte) int {
	return bytes.Compare(ih[:], in)
}

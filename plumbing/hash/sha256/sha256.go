package sha256

import (
	"bytes"
	"encoding/hex"
)

const (
	Size    = 32
	HexSize = Size * 2
)

var (
	zeroHash = SHA256Hash{}
)

type SHA256Hash [Size]byte

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

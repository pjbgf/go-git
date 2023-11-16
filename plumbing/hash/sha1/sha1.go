package sha1

import (
	"bytes"
	"encoding/hex"
)

const (
	Size    = 20
	HexSize = Size * 2
)

var zeroHash = SHA1Hash{}

type SHA1Hash [Size]byte

func (ih SHA1Hash) Size() int {
	return len(ih)
}

func (ih SHA1Hash) IsZero() bool {
	return ih == zeroHash
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

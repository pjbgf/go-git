//go:build !sha256
// +build !sha256

package hash

import (
	"crypto"

	"github.com/go-git/go-git/v5/plumbing/hash/sha1"
)

const (
	// CryptoType defines what hash algorithm is being used.
	CryptoType = crypto.SHA1
	// Size defines the amount of bytes the hash yields.
	Size = sha1.Size
	// HexSize defines the strings size of the hash when represented in hexadecimal.
	HexSize = sha1.HexSize
)

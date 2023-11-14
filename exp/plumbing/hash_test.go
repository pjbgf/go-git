package plumbing_test

import (
	"crypto"
	"fmt"
	"hash"
	"strings"
	"testing"

	. "github.com/go-git/go-git/v5/exp/plumbing"
	"github.com/go-git/go-git/v5/plumbing"
	format "github.com/go-git/go-git/v5/plumbing/format/config"
	"github.com/pjbgf/sha1cd"
	"github.com/stretchr/testify/assert"
)

func TestFromHex(t *testing.T) {
	tests := []struct {
		name string
		in   string
		ok   bool
		zero bool
	}{
		{"valid sha1", "8ab686eafeb1f44702738c8b0f24f2567c36da6d", true, false},
		{"valid sha256", "edeaaff3f1774ad2888673770c6d64097e391bc362d7d6fb34982ddf0efd18cb", true, false},
		{"zero sha1", "0000000000000000000000000000000000000000", true, true},
		{"zero sha256", "0000000000000000000000000000000000000000000000000000000000000000", true, true},
		{"partial sha1", "8ab686eafeb1f44702738", false, true},
		{"partial sha256", "edaaff3f17aaff3f17eaaff3f1774ad288867776666", false, true},
		{"invalid sha1", "8ab686eafeb1f44702738x", false, true},
		{"invalid sha256", "edeaaff3f1774ad28886x", false, true},
	}

	for _, tc := range tests {
		t.Run(fmt.Sprintf("%s:%q", tc.name, tc.in), func(t *testing.T) {
			h, ok := FromHex(tc.in)

			assert.Equal(t, tc.ok, ok, "OK did not match")
			assert.Equal(t, tc.zero, h.IsZero(), "Zero did not match expectations")
		})
	}
}

func TestFromBytes(t *testing.T) {
	tests := []struct {
		name string
		in   []byte
		ok   bool
		zero bool
	}{
		{"valid sha1", NewHash("8ab686eafeb1f44702738c8b0f24f2567c36da6d").Sum(), true, false},
		{"valid sha256", NewHash("edeaaff3f1774ad2888673770c6d64097e391bc362d7d6fb34982ddf0efd18cb").Sum(), true, false},
		{"zero sha1", NewHash("0000000000000000000000000000000000000000").Sum(), true, true},
		{"zero sha256", NewHash("0000000000000000000000000000000000000000000000000000000000000000").Sum(), true, true},
		{"partial sha1", []byte{
			0x8a, 0xb6, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
			0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}, false, false},
	}

	for _, tc := range tests {
		t.Run(fmt.Sprintf("%s:%q", tc.name, tc.in), func(t *testing.T) {
			h := FromBytes(tc.in)

			assert.Equal(t, tc.in, h.Sum(), "in and Sum() did not match")
			assert.Equal(t, tc.zero, h.IsZero(), "Zero did not match expectations")
		})
	}
}

func TestZeroFromHash(t *testing.T) {
	tests := []struct {
		name string
		h    hash.Hash
		want string
	}{
		{"valid sha1", crypto.SHA1.New(), strings.Repeat("0", 40)},
		{"valid sha1cd", sha1cd.New(), strings.Repeat("0", 40)},
		{"valid sha256", crypto.SHA256.New(), strings.Repeat("0", 64)},
		{"unsupported hash", crypto.SHA384.New(), strings.Repeat("0", 40)},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := ZeroFromHash(tc.h)
			assert.Equal(t, tc.want, got.String())
			assert.True(t, got.IsZero(), "should be zero")
		})
	}
}

func TestZeroFromObjectFormat(t *testing.T) {
	tests := []struct {
		name string
		of   format.ObjectFormat
		want string
	}{
		{"valid sha1", format.SHA1, strings.Repeat("0", 40)},
		{"valid sha256", format.SHA256, strings.Repeat("0", 64)},
		{"invalid format", "invalid", strings.Repeat("0", 40)},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := ZeroFromObjectFormat(tc.of)
			assert.Equal(t, tc.want, got.String())
			assert.True(t, got.IsZero(), "should be zero")
		})
	}
}

func BenchmarkHashFromHex(b *testing.B) {
	tests := []struct {
		name   string
		sha1   string
		sha256 string
	}{
		{
			name:   "valid",
			sha1:   "9f361d484fcebb869e1919dc7467b82ac6ca5fad",
			sha256: "2c07a4773e3a957c77810e8cc5deb52cd70493803c048e48dcc0e01f94cbe677",
		},
		{
			name:   "invalid",
			sha1:   "9f361d484fcebb869e1919dc7467b82ac6ca5fxf",
			sha256: "2c07a4773e3a957c77810e8cc5deb52cd70493803c048e48dcc0e01f94cbe6xd",
		},
		{
			name:   "zero",
			sha1:   "0000000000000000000000000000000000000000",
			sha256: "0000000000000000000000000000000000000000000000000000000000000000",
		},
	}

	for _, tc := range tests {
		b.Run(fmt.Sprintf("hasher-parse-sha1-%s", tc.name), func(b *testing.B) {
			benchmarkHashParse(b, tc.sha1)
		})
		b.Run(fmt.Sprintf("objecthash-fromhex-sha1-%s", tc.name), func(b *testing.B) {
			benchmarkObjectHashParse(b, tc.sha1, func(in string) {
				FromHex(in)
			})
		})
		b.Run(fmt.Sprintf("objecthash-fromhex-sha256-%s", tc.name), func(b *testing.B) {
			benchmarkObjectHashParse(b, tc.sha256, func(in string) {
				FromHex(in)
			})
		})
		b.Run(fmt.Sprintf("objecthash-sha1fromhex-%s", tc.name), func(b *testing.B) {
			benchmarkObjectHashParse(b, tc.sha1, func(in string) {
				SHA1HashFromHex(in)
			})
		})
		b.Run(fmt.Sprintf("objecthash-sha256fromhex-%s", tc.name), func(b *testing.B) {
			benchmarkObjectHashParse(b, tc.sha256, func(in string) {
				SHA256HashFromHex(in)
			})
		})
	}
}

func benchmarkHashParse(b *testing.B, in string) {
	for i := 0; i < b.N; i++ {
		_ = plumbing.NewHash(in)
		b.SetBytes(int64(len(in)))
	}
}

func benchmarkObjectHashParse(b *testing.B, in string, fromHex func(in string)) {
	for i := 0; i < b.N; i++ {
		fromHex(in)
		b.SetBytes(int64(len(in)))
	}
}

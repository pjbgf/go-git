package sha256_test

import (
	"fmt"
	"testing"

	. "github.com/go-git/go-git/v5/plumbing/hash/sha256"
	"github.com/stretchr/testify/assert"
)

func TestFromHex(t *testing.T) {
	tests := []struct {
		name string
		in   string
		ok   bool
		zero bool
	}{
		{"valid sha256", "edeaaff3f1774ad2888673770c6d64097e391bc362d7d6fb34982ddf0efd18cb", true, false},
		{"zero sha256", "0000000000000000000000000000000000000000000000000000000000000000", true, true},
		{"partial sha256", "edaaff3f17aaff3f17eaaff3f1774ad288867776666", false, true},
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
	v1, _ := FromHex("edeaaff3f1774ad2888673770c6d64097e391bc362d7d6fb34982ddf0efd18cb")
	v2, _ := FromHex("0000000000000000000000000000000000000000000000000000000000000000")
	tests := []struct {
		name string
		in   []byte
		ok   bool
		zero bool
	}{
		{"valid sha256", v1.Sum(), true, false},
		{"zero sha256", v2.Sum(), true, true},
		{"partial sha1", []byte{
			0x8a, 0xb6, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
			0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
			0x0, 0x0, 0x0, 0x0}, false, false},
	}

	for _, tc := range tests {
		t.Run(fmt.Sprintf("%s:%q", tc.name, tc.in), func(t *testing.T) {
			h := FromBytes(tc.in)

			assert.Equal(t, tc.in, h.Sum(), "in and Sum() did not match")
			assert.Equal(t, tc.zero, h.IsZero(), "Zero did not match expectations")
		})
	}
}

func BenchmarkHashFromHex(b *testing.B) {
	tests := []struct {
		name   string
		sha256 string
	}{
		{
			name:   "valid",
			sha256: "2c07a4773e3a957c77810e8cc5deb52cd70493803c048e48dcc0e01f94cbe677",
		},
		{
			name:   "invalid",
			sha256: "2c07a4773e3a957c77810e8cc5deb52cd70493803c048e48dcc0e01f94cbe6xd",
		},
		{
			name:   "zero",
			sha256: "0000000000000000000000000000000000000000000000000000000000000000",
		},
	}

	for _, tc := range tests {
		b.Run(fmt.Sprintf("objecthash-fromhex-%s", tc.name), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				FromHex(tc.sha256)
				b.SetBytes(int64(len(tc.sha256)))
			}
		})
	}
}

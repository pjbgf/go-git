package sha1_test

import (
	"fmt"
	"testing"

	. "github.com/go-git/go-git/v5/plumbing/hash/sha1"
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
		{"zero sha1", "0000000000000000000000000000000000000000", true, true},
		{"partial sha1", "8ab686eafeb1f44702738", false, true},
		{"invalid sha1", "8ab686eafeb1f44702738x", false, true},
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
	v1, _ := FromHex("8ab686eafeb1f44702738c8b0f24f2567c36da6d")
	v2, _ := FromHex("0000000000000000000000000000000000000000")
	tests := []struct {
		name string
		in   []byte
		ok   bool
		zero bool
	}{
		{"valid sha1", v1.Sum(), true, false},
		{"zero sha1", v2.Sum(), true, true},
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

func BenchmarkHashFromHex(b *testing.B) {
	tests := []struct {
		name string
		sha1 string
	}{
		{
			name: "valid",
			sha1: "9f361d484fcebb869e1919dc7467b82ac6ca5fad",
		},
		{
			name: "invalid",
			sha1: "9f361d484fcebb869e1919dc7467b82ac6ca5fxf",
		},
		{
			name: "zero",
			sha1: "0000000000000000000000000000000000000000",
		},
	}

	for _, tc := range tests {
		b.Run(fmt.Sprintf("objecthash-fromhex-sha1-%s", tc.name), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				FromHex(tc.sha1)
				b.SetBytes(int64(len(tc.sha1)))
			}
		})
	}
}

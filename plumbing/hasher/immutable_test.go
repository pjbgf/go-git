package hasher_test

import (
	"crypto"
	"fmt"
	"testing"

	"github.com/go-git/go-git/v5/plumbing"
	. "github.com/go-git/go-git/v5/plumbing/hasher"
	"github.com/pjbgf/sha1cd"
	"github.com/stretchr/testify/assert"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name  string
		in    string
		ok    bool
		empty bool
	}{
		{"valid sha1", "8ab686eafeb1f44702738c8b0f24f2567c36da6d", true, false},
		{"valid sha256", "edeaaff3f1774ad2888673770c6d64097e391bc362d7d6fb34982ddf0efd18cb", true, false},
		{"empty sha1", "0000000000000000000000000000000000000000", true, true},
		{"empty sha256", "0000000000000000000000000000000000000000000000000000000000000000", true, true},
		{"partial sha1", "8ab686eafeb1f44702738", false, true},
		{"partial sha256", "edeaaff3f1774ad28886", false, true},
		{"invalid sha1", "8ab686eafeb1f44702738x", false, true},
		{"invalid sha256", "edeaaff3f1774ad28886x", false, true},
	}

	for _, tc := range tests {
		t.Run(fmt.Sprintf("%s:%q", tc.name, tc.in), func(t *testing.T) {
			h, ok := Parse(tc.in)

			assert.Equal(t, tc.ok, ok, "OK did not match")
			if tc.ok {
				assert.Equal(t, tc.empty, h.Empty(), "Empty did not match expectations")
			} else {
				assert.Nil(t, h)
			}
		})
	}
}

func TestZeroFromHash(t *testing.T) {
	assert.True(t, ZeroFromHash(crypto.SHA256.New()).Empty(), "NewZero(sha256) should be empty")
	assert.True(t, ZeroFromHash(crypto.SHA1.New()).Empty(), "NewZero(sha1) should be empty")
	assert.True(t, ZeroFromHash(sha1cd.New()).Empty(), "NewZero(sha1cd) should be empty")
}

func BenchmarkHashParse(b *testing.B) {
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
		b.Run(fmt.Sprintf("objecthash-parse-sha1-%s", tc.name), func(b *testing.B) {
			benchmarkObjectHashParse(b, tc.sha1)
		})
		b.Run(fmt.Sprintf("objecthash-parse-sha256-%s", tc.name), func(b *testing.B) {
			benchmarkObjectHashParse(b, tc.sha256)
		})
	}
}

func benchmarkHashParse(b *testing.B, in string) {
	for i := 0; i < b.N; i++ {
		_ = plumbing.NewHash(in)
		b.SetBytes(int64(len(in)))
	}
}

func benchmarkObjectHashParse(b *testing.B, in string) {
	for i := 0; i < b.N; i++ {
		_, _ = Parse(in)
		b.SetBytes(int64(len(in)))
	}
}

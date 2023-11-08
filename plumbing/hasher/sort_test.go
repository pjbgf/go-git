package hasher_test

import (
	"testing"

	"github.com/go-git/go-git/v5/plumbing/hasher"
	"github.com/stretchr/testify/assert"
)

func TestHashesSort(t *testing.T) {
	h1, _ := hasher.Parse("2222222222222222222222222222222222222222")
	h2, _ := hasher.Parse("1111111111111111111111111111111111111111")

	i := []hasher.ImmutableHash{
		h1,
		h2,
	}

	hasher.HashesSort(i)

	assert.Equal(t, h2, i[0])
	assert.Equal(t, h1, i[1])
}

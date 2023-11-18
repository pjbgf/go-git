package packp

import (
	"bytes"

	. "github.com/go-git/go-git/v5/internal/test"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/hash/common"
	"github.com/go-git/go-git/v5/plumbing/hash/sha1"

	. "gopkg.in/check.v1"
)

type ShallowUpdateSuite struct{}

var _ = Suite(&ShallowUpdateSuite{})

func (s *ShallowUpdateSuite) TestDecodeWithLF(c *C) {
	raw := "" +
		"0035shallow aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa\n" +
		"0035shallow bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb\n" +
		"0000"

	su := &ShallowUpdate{}
	err := su.Decode(bytes.NewBufferString(raw))
	c.Assert(err, IsNil)

	plumbing.HashesSort(su.Shallows)

	c.Assert(su.Unshallows, HasLen, 0)
	c.Assert(su.Shallows, HasLen, 2)
	c.Assert(su.Shallows, DeepEquals, []common.ObjectHash{
		X(sha1.FromHex("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")),
		X(sha1.FromHex("bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb")),
	})
}

func (s *ShallowUpdateSuite) TestDecode(c *C) {
	raw := "" +
		"0034shallow aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" +
		"0034shallow bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb" +
		"0000"

	su := &ShallowUpdate{}
	err := su.Decode(bytes.NewBufferString(raw))
	c.Assert(err, IsNil)

	plumbing.HashesSort(su.Shallows)

	c.Assert(su.Unshallows, HasLen, 0)
	c.Assert(su.Shallows, HasLen, 2)
	c.Assert(su.Shallows, DeepEquals, []common.ObjectHash{
		X(sha1.FromHex("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")),
		X(sha1.FromHex("bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb")),
	})
}

func (s *ShallowUpdateSuite) TestDecodeUnshallow(c *C) {
	raw := "" +
		"0036unshallow aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" +
		"0036unshallow bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb" +
		"0000"

	su := &ShallowUpdate{}
	err := su.Decode(bytes.NewBufferString(raw))
	c.Assert(err, IsNil)

	plumbing.HashesSort(su.Unshallows)

	c.Assert(su.Shallows, HasLen, 0)
	c.Assert(su.Unshallows, HasLen, 2)
	c.Assert(su.Unshallows, DeepEquals, []common.ObjectHash{
		X(sha1.FromHex("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")),
		X(sha1.FromHex("bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb")),
	})
}

func (s *ShallowUpdateSuite) TestDecodeMalformed(c *C) {
	raw := "" +
		"0035unshallow aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" +
		"0000"

	su := &ShallowUpdate{}
	err := su.Decode(bytes.NewBufferString(raw))
	c.Assert(err, NotNil)
}

func (s *ShallowUpdateSuite) TestEncodeEmpty(c *C) {
	su := &ShallowUpdate{}
	buf := bytes.NewBuffer(nil)
	c.Assert(su.Encode(buf), IsNil)
	c.Assert(buf.String(), Equals, "0000")
}

func (s *ShallowUpdateSuite) TestEncode(c *C) {
	su := &ShallowUpdate{
		Shallows: []common.ObjectHash{
			X(sha1.FromHex("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")),
			X(sha1.FromHex("bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb")),
		},
		Unshallows: []common.ObjectHash{
			X(sha1.FromHex("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")),
			X(sha1.FromHex("bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb")),
		},
	}
	buf := bytes.NewBuffer(nil)
	c.Assert(su.Encode(buf), IsNil)

	expected := "" +
		"0035shallow aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa\n" +
		"0035shallow bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb\n" +
		"0037unshallow aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa\n" +
		"0037unshallow bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb\n" +
		"0000"

	c.Assert(buf.String(), Equals, expected)
}

func (s *ShallowUpdateSuite) TestEncodeShallow(c *C) {
	su := &ShallowUpdate{
		Shallows: []common.ObjectHash{
			X(sha1.FromHex("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")),
			X(sha1.FromHex("bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb")),
		},
	}
	buf := bytes.NewBuffer(nil)
	c.Assert(su.Encode(buf), IsNil)

	expected := "" +
		"0035shallow aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa\n" +
		"0035shallow bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb\n" +
		"0000"

	c.Assert(buf.String(), Equals, expected)
}

func (s *ShallowUpdateSuite) TestEncodeUnshallow(c *C) {
	su := &ShallowUpdate{
		Unshallows: []common.ObjectHash{
			X(sha1.FromHex("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")),
			X(sha1.FromHex("bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb")),
		},
	}
	buf := bytes.NewBuffer(nil)
	c.Assert(su.Encode(buf), IsNil)

	expected := "" +
		"0037unshallow aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa\n" +
		"0037unshallow bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb\n" +
		"0000"

	c.Assert(buf.String(), Equals, expected)
}

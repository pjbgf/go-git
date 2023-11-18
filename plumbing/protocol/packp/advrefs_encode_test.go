package packp

import (
	"bytes"
	"strings"

	. "github.com/go-git/go-git/v5/internal/test"
	"github.com/go-git/go-git/v5/plumbing/format/pktline"
	"github.com/go-git/go-git/v5/plumbing/hash/common"
	"github.com/go-git/go-git/v5/plumbing/hash/sha1"
	"github.com/go-git/go-git/v5/plumbing/protocol/packp/capability"

	. "gopkg.in/check.v1"
)

type AdvRefsEncodeSuite struct{}

var _ = Suite(&AdvRefsEncodeSuite{})

func testEncode(c *C, input *AdvRefs, expected []byte) {
	var buf bytes.Buffer
	c.Assert(input.Encode(&buf), IsNil)
	obtained := buf.Bytes()

	comment := Commentf("\nobtained = %s\nexpected = %s\n", string(obtained), string(expected))

	c.Assert(obtained, DeepEquals, expected, comment)
}

func (s *AdvRefsEncodeSuite) TestZeroValue(c *C) {
	ar := &AdvRefs{}

	expected := pktlines(c,
		"0000000000000000000000000000000000000000 capabilities^{}\x00\n",
		pktline.FlushString,
	)

	testEncode(c, ar, expected)
}

func (s *AdvRefsEncodeSuite) TestHead(c *C) {
	hash := X(sha1.FromHex("6ecf0ef2c2dffb796033e5a02219af86ec6584e5"))
	ar := &AdvRefs{
		Head: hash,
	}

	expected := pktlines(c,
		"6ecf0ef2c2dffb796033e5a02219af86ec6584e5 HEAD\x00\n",
		pktline.FlushString,
	)

	testEncode(c, ar, expected)
}

func (s *AdvRefsEncodeSuite) TestCapsNoHead(c *C) {
	capabilities := capability.NewList()
	capabilities.Add(capability.MultiACK)
	capabilities.Add(capability.OFSDelta)
	capabilities.Add(capability.SymRef, "HEAD:/refs/heads/master")
	ar := &AdvRefs{
		Capabilities: capabilities,
	}

	expected := pktlines(c,
		"0000000000000000000000000000000000000000 capabilities^{}\x00multi_ack ofs-delta symref=HEAD:/refs/heads/master\n",
		pktline.FlushString,
	)

	testEncode(c, ar, expected)
}

func (s *AdvRefsEncodeSuite) TestCapsWithHead(c *C) {
	hash := X(sha1.FromHex("6ecf0ef2c2dffb796033e5a02219af86ec6584e5"))
	capabilities := capability.NewList()
	capabilities.Add(capability.MultiACK)
	capabilities.Add(capability.OFSDelta)
	capabilities.Add(capability.SymRef, "HEAD:/refs/heads/master")
	ar := &AdvRefs{
		Head:         hash,
		Capabilities: capabilities,
	}

	expected := pktlines(c,
		"6ecf0ef2c2dffb796033e5a02219af86ec6584e5 HEAD\x00multi_ack ofs-delta symref=HEAD:/refs/heads/master\n",
		pktline.FlushString,
	)

	testEncode(c, ar, expected)
}

func (s *AdvRefsEncodeSuite) TestRefs(c *C) {
	references := map[string]common.ObjectHash{
		"refs/heads/master":      X(sha1.FromHex("a6930aaee06755d1bdcfd943fbf614e4d92bb0c7")),
		"refs/tags/v2.6.12-tree": X(sha1.FromHex("1111111111111111111111111111111111111111")),
		"refs/tags/v2.7.13-tree": X(sha1.FromHex("3333333333333333333333333333333333333333")),
		"refs/tags/v2.6.13-tree": X(sha1.FromHex("2222222222222222222222222222222222222222")),
		"refs/tags/v2.6.11-tree": X(sha1.FromHex("5dc01c595e6c6ec9ccda4f6f69c131c0dd945f8c")),
	}
	ar := &AdvRefs{
		References: references,
	}

	expected := pktlines(c,
		"a6930aaee06755d1bdcfd943fbf614e4d92bb0c7 refs/heads/master\x00\n",
		"5dc01c595e6c6ec9ccda4f6f69c131c0dd945f8c refs/tags/v2.6.11-tree\n",
		"1111111111111111111111111111111111111111 refs/tags/v2.6.12-tree\n",
		"2222222222222222222222222222222222222222 refs/tags/v2.6.13-tree\n",
		"3333333333333333333333333333333333333333 refs/tags/v2.7.13-tree\n",
		pktline.FlushString,
	)

	testEncode(c, ar, expected)
}

func (s *AdvRefsEncodeSuite) TestPeeled(c *C) {
	references := map[string]common.ObjectHash{
		"refs/heads/master":      X(sha1.FromHex("a6930aaee06755d1bdcfd943fbf614e4d92bb0c7")),
		"refs/tags/v2.6.12-tree": X(sha1.FromHex("1111111111111111111111111111111111111111")),
		"refs/tags/v2.7.13-tree": X(sha1.FromHex("3333333333333333333333333333333333333333")),
		"refs/tags/v2.6.13-tree": X(sha1.FromHex("2222222222222222222222222222222222222222")),
		"refs/tags/v2.6.11-tree": X(sha1.FromHex("5dc01c595e6c6ec9ccda4f6f69c131c0dd945f8c")),
	}
	peeled := map[string]common.ObjectHash{
		"refs/tags/v2.7.13-tree": X(sha1.FromHex("4444444444444444444444444444444444444444")),
		"refs/tags/v2.6.12-tree": X(sha1.FromHex("5555555555555555555555555555555555555555")),
	}
	ar := &AdvRefs{
		References: references,
		Peeled:     peeled,
	}

	expected := pktlines(c,
		"a6930aaee06755d1bdcfd943fbf614e4d92bb0c7 refs/heads/master\x00\n",
		"5dc01c595e6c6ec9ccda4f6f69c131c0dd945f8c refs/tags/v2.6.11-tree\n",
		"1111111111111111111111111111111111111111 refs/tags/v2.6.12-tree\n",
		"5555555555555555555555555555555555555555 refs/tags/v2.6.12-tree^{}\n",
		"2222222222222222222222222222222222222222 refs/tags/v2.6.13-tree\n",
		"3333333333333333333333333333333333333333 refs/tags/v2.7.13-tree\n",
		"4444444444444444444444444444444444444444 refs/tags/v2.7.13-tree^{}\n",
		pktline.FlushString,
	)

	testEncode(c, ar, expected)
}

func (s *AdvRefsEncodeSuite) TestShallow(c *C) {
	shallows := []common.ObjectHash{
		X(sha1.FromHex("1111111111111111111111111111111111111111")),
		X(sha1.FromHex("4444444444444444444444444444444444444444")),
		X(sha1.FromHex("3333333333333333333333333333333333333333")),
		X(sha1.FromHex("2222222222222222222222222222222222222222")),
	}
	ar := &AdvRefs{
		Shallows: shallows,
	}

	expected := pktlines(c,
		"0000000000000000000000000000000000000000 capabilities^{}\x00\n",
		"shallow 1111111111111111111111111111111111111111\n",
		"shallow 2222222222222222222222222222222222222222\n",
		"shallow 3333333333333333333333333333333333333333\n",
		"shallow 4444444444444444444444444444444444444444\n",
		pktline.FlushString,
	)

	testEncode(c, ar, expected)
}

func (s *AdvRefsEncodeSuite) TestAll(c *C) {
	hash := X(sha1.FromHex("6ecf0ef2c2dffb796033e5a02219af86ec6584e5"))

	capabilities := capability.NewList()
	capabilities.Add(capability.MultiACK)
	capabilities.Add(capability.OFSDelta)
	capabilities.Add(capability.SymRef, "HEAD:/refs/heads/master")

	references := map[string]common.ObjectHash{
		"refs/heads/master":      X(sha1.FromHex("a6930aaee06755d1bdcfd943fbf614e4d92bb0c7")),
		"refs/tags/v2.6.12-tree": X(sha1.FromHex("1111111111111111111111111111111111111111")),
		"refs/tags/v2.7.13-tree": X(sha1.FromHex("3333333333333333333333333333333333333333")),
		"refs/tags/v2.6.13-tree": X(sha1.FromHex("2222222222222222222222222222222222222222")),
		"refs/tags/v2.6.11-tree": X(sha1.FromHex("5dc01c595e6c6ec9ccda4f6f69c131c0dd945f8c")),
	}

	peeled := map[string]common.ObjectHash{
		"refs/tags/v2.7.13-tree": X(sha1.FromHex("4444444444444444444444444444444444444444")),
		"refs/tags/v2.6.12-tree": X(sha1.FromHex("5555555555555555555555555555555555555555")),
	}

	shallows := []common.ObjectHash{
		X(sha1.FromHex("1111111111111111111111111111111111111111")),
		X(sha1.FromHex("4444444444444444444444444444444444444444")),
		X(sha1.FromHex("3333333333333333333333333333333333333333")),
		X(sha1.FromHex("2222222222222222222222222222222222222222")),
	}

	ar := &AdvRefs{
		Head:         hash,
		Capabilities: capabilities,
		References:   references,
		Peeled:       peeled,
		Shallows:     shallows,
	}

	expected := pktlines(c,
		"6ecf0ef2c2dffb796033e5a02219af86ec6584e5 HEAD\x00multi_ack ofs-delta symref=HEAD:/refs/heads/master\n",
		"a6930aaee06755d1bdcfd943fbf614e4d92bb0c7 refs/heads/master\n",
		"5dc01c595e6c6ec9ccda4f6f69c131c0dd945f8c refs/tags/v2.6.11-tree\n",
		"1111111111111111111111111111111111111111 refs/tags/v2.6.12-tree\n",
		"5555555555555555555555555555555555555555 refs/tags/v2.6.12-tree^{}\n",
		"2222222222222222222222222222222222222222 refs/tags/v2.6.13-tree\n",
		"3333333333333333333333333333333333333333 refs/tags/v2.7.13-tree\n",
		"4444444444444444444444444444444444444444 refs/tags/v2.7.13-tree^{}\n",
		"shallow 1111111111111111111111111111111111111111\n",
		"shallow 2222222222222222222222222222222222222222\n",
		"shallow 3333333333333333333333333333333333333333\n",
		"shallow 4444444444444444444444444444444444444444\n",
		pktline.FlushString,
	)

	testEncode(c, ar, expected)
}

func (s *AdvRefsEncodeSuite) TestErrorTooLong(c *C) {
	references := map[string]common.ObjectHash{
		strings.Repeat("a", pktline.MaxPayloadSize): X(sha1.FromHex("a6930aaee06755d1bdcfd943fbf614e4d92bb0c7")),
	}
	ar := &AdvRefs{
		References: references,
	}

	var buf bytes.Buffer
	err := ar.Encode(&buf)
	c.Assert(err, ErrorMatches, ".*payload is too long.*")
}

package packp

import (
	"bytes"

	. "github.com/go-git/go-git/v5/internal/test"
	"github.com/go-git/go-git/v5/plumbing/hash/sha1"
	"github.com/go-git/go-git/v5/plumbing/protocol/packp/capability"

	. "gopkg.in/check.v1"
)

type UploadPackRequestSuite struct{}

var _ = Suite(&UploadPackRequestSuite{})

func (s *UploadPackRequestSuite) TestNewUploadPackRequestFromCapabilities(c *C) {
	cap := capability.NewList()
	cap.Set(capability.Agent, "foo")

	r := NewUploadPackRequestFromCapabilities(cap)
	c.Assert(r.Capabilities.String(), Equals, "agent=go-git/5.x")
}

func (s *UploadPackRequestSuite) TestIsEmpty(c *C) {
	r := NewUploadPackRequest()
	r.Wants = append(r.Wants, X(sha1.FromHex("d82f291cde9987322c8a0c81a325e1ba6159684c")))
	r.Wants = append(r.Wants, X(sha1.FromHex("2b41ef280fdb67a9b250678686a0c3e03b0a9989")))
	r.Haves = append(r.Haves, X(sha1.FromHex("6ecf0ef2c2dffb796033e5a02219af86ec6584e5")))

	c.Assert(r.IsEmpty(), Equals, false)

	r = NewUploadPackRequest()
	r.Wants = append(r.Wants, X(sha1.FromHex("d82f291cde9987322c8a0c81a325e1ba6159684c")))
	r.Wants = append(r.Wants, X(sha1.FromHex("2b41ef280fdb67a9b250678686a0c3e03b0a9989")))
	r.Haves = append(r.Haves, X(sha1.FromHex("d82f291cde9987322c8a0c81a325e1ba6159684c")))

	c.Assert(r.IsEmpty(), Equals, false)

	r = NewUploadPackRequest()
	r.Wants = append(r.Wants, X(sha1.FromHex("d82f291cde9987322c8a0c81a325e1ba6159684c")))
	r.Haves = append(r.Haves, X(sha1.FromHex("d82f291cde9987322c8a0c81a325e1ba6159684c")))

	c.Assert(r.IsEmpty(), Equals, true)

	r = NewUploadPackRequest()
	r.Wants = append(r.Wants, X(sha1.FromHex("d82f291cde9987322c8a0c81a325e1ba6159684c")))
	r.Haves = append(r.Haves, X(sha1.FromHex("d82f291cde9987322c8a0c81a325e1ba6159684c")))
	r.Shallows = append(r.Shallows, X(sha1.FromHex("2b41ef280fdb67a9b250678686a0c3e03b0a9989")))

	c.Assert(r.IsEmpty(), Equals, false)
}

type UploadHavesSuite struct{}

var _ = Suite(&UploadHavesSuite{})

func (s *UploadHavesSuite) TestEncode(c *C) {
	uh := &UploadHaves{}
	uh.Haves = append(uh.Haves,
		X(sha1.FromHex("1111111111111111111111111111111111111111")),
		X(sha1.FromHex("3333333333333333333333333333333333333333")),
		X(sha1.FromHex("1111111111111111111111111111111111111111")),
		X(sha1.FromHex("2222222222222222222222222222222222222222")),
		X(sha1.FromHex("1111111111111111111111111111111111111111")),
	)

	buf := bytes.NewBuffer(nil)
	err := uh.Encode(buf, true)
	c.Assert(err, IsNil)
	c.Assert(buf.String(), Equals, ""+
		"0032have 1111111111111111111111111111111111111111\n"+
		"0032have 2222222222222222222222222222222222222222\n"+
		"0032have 3333333333333333333333333333333333333333\n"+
		"0000",
	)
}

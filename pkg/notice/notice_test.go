package notice

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestNotice(t *testing.T) {
	suite.Run(t, new(NoticeTestSuite))
}

type DummyNotice struct {
	msg string
}

func (n *DummyNotice) Notice(msg Msg, srv any) error {
	n.msg = msg.String()
	return nil
}

type NoticeTestSuite struct {
	suite.Suite
}

func (s *NoticeTestSuite) SetupTest() {
	notices = []Notice{}
	RegisterNotice(new(DummyNotice))
}

func (s *NoticeTestSuite) TestBroadCast() {
	msg := "Test msg"
	BroadCast(TextMsg(msg), nil)

	n, ok := notices[0].(*DummyNotice)
	s.True(ok)
	s.Equal(msg, n.msg)
}

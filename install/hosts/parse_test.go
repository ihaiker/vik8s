package hosts

import (
	"github.com/stretchr/testify/suite"
	"testing"
)

type ParseAddressSuite struct {
	suite.Suite
	opt *Option
}

func (t *ParseAddressSuite) SetupTest() {
	t.opt = &Option{
		User: "root", Password: "test", Port: "22",
	}
}

func (t ParseAddressSuite) TestMerge() {
	ip, err := merge_end_ip("10.24.0.10", "20")
	t.Nil(err)
	t.Equal("10.24.0.20", ip)

	ip, err = merge_end_ip("10.24.0.10", "1.20")
	t.Nil(err)
	t.Equal("10.24.1.20", ip)

	ip, err = merge_end_ip("10.24.0.10", "25.1.20")
	t.Nil(err)
	t.Equal("10.25.1.20", ip)
}

func (t ParseAddressSuite) TestBase() {
	nodes, err := parse_addr(*t.opt, "10.24.0.10")
	t.Nil(err, "parse error")
	t.Equal(1, len(nodes), "node size not equal 1")

	nodes, err = parse_addr(*t.opt, "10.24.0.10-10.24.0.11")
	t.Nil(err, "parse error")
	t.Equal(2, len(nodes), "node size not equal 1")
	t.Equal("10.24.0.10", nodes[0].Host)
	t.Equal("10.24.0.11", nodes[1].Host)

	nodes, err = parse_addr(*t.opt, "test@10.24.0.10-10.24.0.11")
	t.Nil(err, "parse error")
	t.Equal(2, len(nodes), "node size not equal 2")
	t.Equal("10.24.0.10", nodes[0].Host)
	t.Equal("test", nodes[1].User)

	nodes, err = parse_addr(*t.opt, "test@10.24.0.10-10.24.0.11:234")
	t.Nil(err, "parse error")
	t.Equal(2, len(nodes), "node size not equal 2")
	t.Equal("10.24.0.10", nodes[0].Host)
	t.Equal("test", nodes[1].User)
	t.Equal("234", nodes[1].Port)

	nodes, err = parse_addr(*t.opt, "test:123@10.24.0.10-11:234")
	t.Nil(err, "parse error")
	t.Equal(2, len(nodes), "node size not equal 2")
	t.Equal("10.24.0.10", nodes[0].Host)
	t.Equal("test", nodes[1].User)
	t.Equal("234", nodes[1].Port)
	t.Equal("123", nodes[1].Password)

	nodes, err = parse_addr(*t.opt, "test:123@10.24.0.10-15:234")
	t.Nil(err, "parse error")
	t.Equal(6, len(nodes), "node size not equal 26")
	t.Equal("10.24.0.10", nodes[0].Host)
}

func (t ParseAddressSuite) TestArgs() {
	nodes, err := parse_addrs(*t.opt, "10.24.0.10")
	t.Nil(err, "parse error")
	t.Equal(1, len(nodes), "node size not equal 1")

	nodes, err = parse_addrs(*t.opt, "10.24.0.10", "10.24.0.11")
	t.Nil(err, "parse error")
	t.Equal(2, len(nodes), "node size not equal 1")
	t.Equal("10.24.0.11", nodes[1].Host)

	nodes, err = parse_addrs(*t.opt, "10.24.0.10-11", "10.24.0.20-23")
	t.Nil(err, "parse error")
	t.Equal(6, len(nodes), "node size not equal 1")
	t.Equal("10.24.0.21", nodes[3].Host)
}

func TestParseAddress(t *testing.T) {
	suite.Run(t, new(ParseAddressSuite))
}

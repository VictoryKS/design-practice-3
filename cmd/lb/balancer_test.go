package main

import (
  "testing"
  . "gopkg.in/check.v1"
)

func Test(t *testing.T) { TestingT(t) }

type MySuite struct{}

var _ = Suite(&MySuite{})

var tests = [][3]struct {
  health bool
  connections int
}{
  {{true, 1}, {true, 2}, {true, 3}},
  {{true, 2}, {true, 3}, {true, 1}},
  {{true, 3}, {true, 1}, {true, 2}},
  {{false, 1}, {true, 2}, {true, 3}},
  {{false, 1}, {false, 2}, {true, 3}},
  {{false, 1}, {false, 1}, {false, 1}},
}

var expIndex = []int{0, 2, 1, 1, 2, -1}

func (s *MySuite) TestBalancer(c *C) {
  for j, test := range tests {
    for i, inServer := range test {
      server := serversPool[i]
      (*server).isHealthy = inServer.health
      (*server).connCnt = inServer.connections
    }
    c.Assert(expIndex[j], Equals, findMin(serversPool))
	}
}

package main

import (
  "testing"
  "container/heap"
  . "gopkg.in/check.v1"
)

func Test(t *testing.T) { TestingT(t) }

type MySuite struct{}

var _ = Suite(&MySuite{})

var tests = [][3]struct {
  health bool
  connections int
}{
  {{true, 3}, {true, 2}, {true, 1}},
  {{true, 2}, {true, 3}, {true, 1}},
  {{false, 1}, {true, 3}, {true, 2}},
  {{false, 1}, {false, 2}, {true, 3}},
}

var expIndex = []int{2, 2, 2, 2}

func (s *MySuite) TestBalancer(c *C) {
  for _, server := range serversPool {
    pq[(*server).index] = server
  }
  heap.Init(&pq)

  for j, test := range tests {
    for i, inServer := range test {
      server := serversPool[i]
      (*server).isHealthy = inServer.health
      (*server).connCnt = inServer.connections

      pq.update(server, (*server).connCnt, (*server).isHealthy)
    }

    c.Assert(findMin(&pq), Equals, expIndex[j])
	}
}

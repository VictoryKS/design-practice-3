package integration

import (
	"fmt"
	"net/http"
	"testing"
	"time"
	// "log"
	"sync"
	"math/rand"
	"strconv"

	. "gopkg.in/check.v1"
)

const baseAddress = "http://localhost:8090"

type server struct {
	name string
	connCnt int
	mutex sync.Mutex
}

var (
	serversPool = []*server {
		&server {
			name: "server1:8080",
			connCnt: 0,
		},
		&server {
			name: "server2:8080",
			connCnt: 0,
		},
		&server {
			name: "server3:8080",
			connCnt: 0,
		},
	}
)

var client = http.Client{
	Timeout: 3 * time.Second,
}

func Test(t *testing.T) { TestingT(t) }

type MySuite struct{}

var _ = Suite(&MySuite{})

func (s *MySuite) TestBalancer(c *C) {
	var wg sync.WaitGroup

	n := 30
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(wg *sync.WaitGroup){
			defer wg.Done()

			time.Sleep(time.Duration(rand.Intn(10)) * time.Second)

			resp, err := client.Get(fmt.Sprintf("%s/api/v1/some-data", baseAddress))
			if err != nil {
				c.Logf("Response error%s", err)
			}

			server := resp.Header.Get("lb-from")
			// log.Printf("response from [%s]", server)

			sum := 0
			for i := 0; i < len(serversPool); i++ {
				svr := serversPool[i]
				if (*svr).name == server {
					(*svr).mutex.Lock()
					(*svr).connCnt++
					(*svr).mutex.Unlock()
				}
				sum += (*svr).connCnt
			}

			sum = sum/len(serversPool) + 1
			for i := 0; i < len(serversPool); i++ {
				test := false
				if (*serversPool[i]).connCnt-1 <= sum {
					test = true
				}
				c.Assert(test, Equals, true)
			}

		}(&wg)
	}
	wg.Wait()

}

func BenchmarkBalancer(b *testing.B) {
	var wg sync.WaitGroup

	for i := 0; i < b.N; i++ {
		wg.Add(1)
		go func(group sync.WaitGroup) {
			defer wg.Done()

			_, err := client.Get(fmt.Sprintf("%s/api/v1/some-data", baseAddress))

			if err != nil {
				b.Logf("Response error%s", err)
			}
		}(wg)
	}
	wg.Wait()
}

func BenchmarkServers(b *testing.B) {
	var wg sync.WaitGroup

	for i := 0; i < b.N; i++ {
		wg.Add(1)
		go func(group sync.WaitGroup) {
			defer wg.Done()

			svrAddress := "http://localhost:808" + strconv.Itoa(i % 3)
			_, err := client.Get(fmt.Sprintf("%s/api/v1/some-data", svrAddress))

			if err != nil {
				b.Error(err)
			}
		}(wg)
	}
	wg.Wait()
}

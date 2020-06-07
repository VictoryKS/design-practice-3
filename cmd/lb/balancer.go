package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
	"sync"
	"container/heap"

	"github.com/VictoryKS/design-practice-3/httptools"
	"github.com/VictoryKS/design-practice-3/signal"
)

type server struct {
	name string
	isHealthy bool
	connCnt int
	mutex sync.Mutex
	index int
}

//------------------
type PriorityQueue []*server


func (pq *PriorityQueue) Push(x interface{}) {
	n := len(*pq)
	item := x.(*server)
	item.index = n
	*pq = append(*pq, item)
}

func (pq *PriorityQueue) update(item *server, connCnt int, isHealthy bool) {
	item.connCnt = connCnt
	item.isHealthy = isHealthy
	heap.Fix(pq, item.index)
}

func (pq PriorityQueue) Len() int { return len(pq) }

func (pq PriorityQueue) Less(i, j int) bool {
	// We want Pop to give us the highest, not lowest, priority so we use greater than here.
	return pq[i].connCnt > pq[j].connCnt
}

func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}


func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil  // avoid memory leak
	item.index = -1 // for safety
	*pq = old[0 : n-1]
	return item
}
//------------------

var (
	port = flag.Int("port", 8090, "load balancer port")
	timeoutSec = flag.Int("timeout-sec", 3, "request timeout time in seconds")
	https = flag.Bool("https", false, "whether backends support HTTPs")

	traceEnabled = flag.Bool("trace", true, "whether to include tracing information into responses")
)

var (
	timeout = time.Duration(*timeoutSec) * time.Second
	serversPool = []*server {
		&server {
			name: "server1:8080",
			isHealthy: false,
			connCnt: 0,
			index: 0,
		},
		&server {
			name: "server2:8080",
			isHealthy: false,
			connCnt: 0,
			index: 1,
		},
		&server {
			name: "server3:8080",
			isHealthy: false,
			connCnt: 0,
			index: 2,
		},
	}
	pq = make(PriorityQueue, len(serversPool))
)

func findMin(pq *PriorityQueue) int {
	old := *pq
	n := len(old)
	item := old[n-1]
	return item.index
}

func scheme() string {
	if *https {
		return "https"
	}
	return "http"
}

func health(dst string) bool {
	ctx, _ := context.WithTimeout(context.Background(), timeout)
	req, _ := http.NewRequestWithContext(ctx, "GET",
		fmt.Sprintf("%s://%s/health", scheme(), dst), nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false
	}
	if resp.StatusCode != http.StatusOK {
		return false
	}
	return true
}

func forward(server *server, rw http.ResponseWriter, r *http.Request) error {
	server.mutex.Lock()
	(*server).connCnt++
	server.mutex.Unlock()

	dst := (*server).name
	ctx, _ := context.WithTimeout(r.Context(), timeout)
	fwdRequest := r.Clone(ctx)
	fwdRequest.RequestURI = ""
	fwdRequest.URL.Host = dst
	fwdRequest.URL.Scheme = scheme()
	fwdRequest.Host = dst

	resp, err := http.DefaultClient.Do(fwdRequest)
	if err == nil {
		for k, values := range resp.Header {
			for _, value := range values {
				rw.Header().Add(k, value)
			}
		}
		if *traceEnabled {
			rw.Header().Set("lb-from", dst)
		}
		log.Println("fwd", resp.StatusCode, resp.Request.URL)
		rw.WriteHeader(resp.StatusCode)
		defer resp.Body.Close()
		_, err := io.Copy(rw, resp.Body)
		if err != nil {
			log.Printf("Failed to write response: %s", err)
		}

		server.mutex.Lock()
		(*server).connCnt--
		server.mutex.Unlock()

		return nil
	} else {
		log.Printf("Failed to get response from %s: %s", dst, err)
		rw.WriteHeader(http.StatusServiceUnavailable)

		server.mutex.Lock()
		(*server).connCnt--
		server.mutex.Unlock()

		return err
	}
}

func main() {
	flag.Parse()

	for _, server := range serversPool {
		pq[(*server).index] = server
	}
	heap.Init(&pq)

  for _, server := range serversPool {
		server := server
		(*server).isHealthy = health((*server).name) // immediately after Start
		go func() {
			for range time.Tick(10 * time.Second) {
				(*server).isHealthy = health((*server).name)
			}
		}()
	}

	frontend := httptools.CreateServer(*port, http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		serverIndex := findMin(&pq)

		if serverIndex >= 0 {
			server := serversPool[serverIndex]
			forward(server, rw, r)
		}
	}))

	log.Println("Starting load balancer...")
	log.Printf("Tracing support enabled: %t", *traceEnabled)
	frontend.Start()
	signal.WaitForTerminationSignal()
}

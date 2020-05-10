package integration

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	// "gopkg.in/check.v1"
)

const baseAddress = "http://balancer:8090"

var client = http.Client{
	Timeout: 3 * time.Second,
}

func TestBalancer(t *testing.T) {
	// TODO: Реалізуйте інтеграційний тест для балансувальникка.
	resp, err := client.Get(fmt.Sprintf("%s/api/v1/some-data", baseAddress))
	if err != nil {
		t.Error(err)
	}
	t.Logf("response from [%s]", resp.Header.Get("lb-from"))

	t.Skip("...skipping test...")
	fmt.Sprintf("------> TestBalancer <--------")
}

func BenchmarkBalancer(b *testing.B) {
	// TODO: Реалізуйте інтеграційний бенчмарк для балансувальникка.

	for i := 0; i < b.N; i++ {
		b.Skip("...skipping BenchmarkBalancer test...")
		fmt.Sprintf("------> BenchmarkBalancer <--------")
  }

}

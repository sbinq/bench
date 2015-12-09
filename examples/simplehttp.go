package main

import (
	"fmt"
	"github.com/chirayu/bench"
	"net/http"
	"time"
)

func f(id int, iter int) time.Duration {
	start := time.Now()
	resp, err := http.Get("http://www.flipkart.com")
	if err != nil {
		fmt.Printf("Error %s", err)
	}
	defer resp.Body.Close()
	return time.Since(start)
}

func main() {
	b := bench.NewBench(10, time.Second*10, f)
	b.Run()
	fmt.Printf("%s", b)
}

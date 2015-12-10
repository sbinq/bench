package main

import (
	"fmt"
	"github.com/chirayu/bench"
	"net/http"
	"time"
)

func f(c *bench.Context) {
	start := time.Now()
	resp, err := http.Get("http://www.flipkart.com")
	if err != nil {
		fmt.Printf("Error %s", err)
	}
	defer resp.Body.Close()
	c.RecordTime(time.Since(start))
}

func main() {
	b := bench.NewBench(10, time.Second*10, f)
	b.Run()
	fmt.Printf("%s", b)
}

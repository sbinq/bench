package main

import (
	"fmt"
	"github.com/chirayu/bench"
	"net/http"
	"time"
)

func f(c *bench.Context) {
	var client *http.Client 
	v, ok := c.Values["client"]
	if !ok {
		client = &http.Client{}
		c.Incr("Counter", 1)
		c.Values["client"] = client
	} else {
		client = v.(*http.Client)
	}
	
	start := time.Now()
	resp, err := client.Get("http://localhost:8000")
	if err != nil {
		fmt.Printf("Error :  %s\n", err)
		//c.Incr("Errors", 1)
		return
	}
	defer resp.Body.Close()
	c.Timer("Latency", time.Since(start))
	// c.Incr("Counter", 1)
}

func main() {
	b := bench.NewBench(10, time.Second*10, f)
	b.UniformRun(100)
	fmt.Printf("%s", b)
}

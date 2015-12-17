// A simple benchmarking example
package main

import (
	"fmt"
	"github.com/chirayu/bench"
	"time"
)

func delayedFn(c *bench.Context) {
	t := time.Now()
	time.Sleep(100 * time.Millisecond)
	c.Timer("Latency", time.Since(t))
	c.Incr("Call Counter", 1)
}

func quickFn(c *bench.Context) {
	t := time.Now()
	for i := 0; i < 1000*1000; i++ {
	}
	c.Timer("Latency", time.Since(t))
	c.Incr("Call Counter", 1)
	//c.Incr("Another Call Counter", 10)
}

func main() {
 	// run at the provided rps
	b1 := bench.NewBench(2, time.Second*5, 10, quickFn)
	b1.Run()
	fmt.Printf("%s", b1)
}

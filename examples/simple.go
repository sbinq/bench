// A simple benchmarking example
package main

import (
	"fmt"
	"github.com/chirayu/bench"
	"time"
)

func f(c *bench.Context) {
	time.Sleep(100 * time.Millisecond)
	c.Timer("Latency", time.Duration(c.Iteration) * time.Millisecond)
	c.Incr("Call Counter", 1)
}

func f1(c *bench.Context) {
	c.Timer("Latency", time.Duration(c.Iteration) * time.Millisecond)
	c.Incr("Call Counter", 1)
}

func main() {
	// run at full steam
	// b := bench.NewBench(1, time.Second*2, f)
	// b.Run()
	// fmt.Printf("%s", b)
	
	// run at the provided rps
	b1 := bench.NewBench(2, time.Second*5, f)
	b1.UniformRun(10)
	fmt.Printf("%s", b1)
}

// A simple benchmarking example
package main

import (
	"fmt"
	"github.com/chirayu/bench"
	"time"
)

func f(c *bench.Context) {
	time.Sleep(100 * time.Millisecond)
	c.RecordTime(time.Duration(c.Iteration) * time.Millisecond)
}

func main() {
	b := bench.NewBench(10, time.Second*2, f)
	b.Run()
	fmt.Printf("%s", b)
}

//Package bench makes it easy to write benchamrking apps.
package bench

import (
	"bytes"
	"fmt"
	"sync"
	"time"

	"github.com/codahale/hdrhistogram"
)

const (
	min        = 0
	max        = 10000000
	resolution = 3
)

// Bench holds data to be used for benchmarking
type Bench struct {
	concurrentTasks int           // number of concurrent tasks
	duration        time.Duration // benchmark duration
	toBenchmark     func(int, int) time.Duration

	// histogram for the benchmarking run
	histogram *hdrhistogram.Histogram
}

// NewBench creates a new instance of Bench
func NewBench(concurrentTasks int, duration time.Duration, toBenchmark func(int, int) (time.Duration)) *Bench {
	return &Bench{
		concurrentTasks: concurrentTasks,
		duration:        duration,
		toBenchmark:     toBenchmark,
		histogram:       hdrhistogram.New(min, max, resolution),
	}
}

// Run the benchmark
func (b *Bench) Run() {
	c := make(chan *hdrhistogram.Histogram, b.concurrentTasks)
	var wg sync.WaitGroup

	for i := 1; i <= b.concurrentTasks; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			histogram := hdrhistogram.New(min, max, resolution)
			for j := 1; ; j++ {
				timeTaken := b.toBenchmark(i, j)
				histogram.RecordValue(timeTaken.Nanoseconds())
			}
		}()
	}

	wg.Wait()
	close(c)

	for h := range c {
		h.Merge(h)
	}
}

// String converts the output of the bench into a printable form
func (b *Bench) String() string {
	var buf bytes.Buffer
	percentiles := []float64{99, 99.9, 100}

	for _, p := range percentiles {
		fmt.Fprintf(&buf, "%2.1fth percentile: %dns\n", p, b.histogram.ValueAtQuantile(p))
	}
	return buf.String()
}

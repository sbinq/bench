//Package bench makes it easy to write benchamrking apps.
package bench

import (
	"bytes"
	"fmt"
	"sync"
	"time"

	"github.com/codahale/hdrhistogram"
	"golang.org/x/net/context"
)

const (
	min        = 0
	max        = 100 * 1000 * 1000 * 1000 // upto 10 seconds
	resolution = 3
)

// Bench holds data to be used for benchmarking
type Bench struct {
	// values configured by the user
	concurrentTasks int
	duration        time.Duration
	toBenchmark     func(int, int) time.Duration

	// histogram for the benchmarking run
	histogram *hdrhistogram.Histogram
	// histograms for individual tasks
	taskHistograms []*hdrhistogram.Histogram
	// total calls made to each task
	calls int
	// time taken for the full run
	timeTaken time.Duration
}

// NewBench creates a new instance of Bench
func NewBench(concurrentTasks int, duration time.Duration, toBenchmark func(int, int) time.Duration) *Bench {
	b := &Bench{
		concurrentTasks: concurrentTasks,
		duration:        duration,
		toBenchmark:     toBenchmark,
		histogram:       hdrhistogram.New(min, max, resolution),
	}

	for i := 0; i < concurrentTasks; i++ {
		b.taskHistograms = append(b.taskHistograms, hdrhistogram.New(min, max, resolution))
	}

	return b
}

// Run the benchmark
func (b *Bench) Run() {
	var wg sync.WaitGroup
	start := time.Now()
	c := make(chan int, b.concurrentTasks)
	ctx, _ := context.WithTimeout(context.Background(), b.duration)

	for i := 1; i <= b.concurrentTasks; i++ {
		wg.Add(1)

		go func(ctx context.Context, c chan int, histogram *hdrhistogram.Histogram) {
			defer wg.Done()

			for j := 1; ; j++ {
				select {
				case <-ctx.Done():
					c <- j
					return
				default:
					timeTaken := b.toBenchmark(i, j)
					histogram.RecordValue(timeTaken.Nanoseconds())
				}
			}
		}(ctx, c, b.taskHistograms[i-1])
	}

	wg.Wait()
	close(c)

	// merge task specific data
	for i := 0; i < b.concurrentTasks; i++ {
		b.histogram.Merge(b.taskHistograms[i])
		b.calls += <-c
	}

	b.timeTaken = time.Since(start)
}

// String converts the output of the bench into a printable form
func (b *Bench) String() string {
	prefix := "  "
	var buf bytes.Buffer
	percentiles := []float64{50, 99.9, 100}

	fmt.Fprintf(&buf, "Duration: %2.2f, Concurrency: %d, Total runs: %d\n", b.timeTaken.Seconds(), b.concurrentTasks, b.calls)
	for _, p := range percentiles {
		fmt.Fprintf(&buf, "%s%2.1fth percentile: %.2fms\n", prefix, p, float32(b.histogram.ValueAtQuantile(p))/1000000.0)
	}
	fmt.Fprintf(&buf, "%s%d calls in %.2fs\n", prefix, b.calls, b.duration.Seconds())
	return buf.String()
}

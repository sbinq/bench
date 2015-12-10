# What is Bench?

Bench is a golang package to help speed up benchmarking projects. You can use it to do a quick benchmark of your database, queue or http server.

## What does it do?

Bench repeatedly calls a function with the specificied concurrency. 

# How do I use it? 

Bench is easy to integrate with your code. This code benchmarks a http server with a concurrency of 10 for 10 seconds.

```
func benchmarkMe(c *bench.Context) {
	start := time.Now()
	resp, err := http.Get("http://www.flipkart.com")
	if err != nil {
		fmt.Printf("Error %s", err)
	}
	defer resp.Body.Close()
	c.RecordTime(time.Since(start))
}

func main() {
	b := bench.NewBench(10, time.Second*10, benchmarkMe)
	b.Run()
	fmt.Printf("%s", b)
```

# What's next?

Two features are in the offing.

1. Ability to capture multiple metrics during runtime. Apart from latency, a http client may want to record data size and error codes. 
2. Make requests at a uniform rate. This is useful when you want to see if you server can sustain a load of X per second. 
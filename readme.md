
# Bench [![Build Status](https://travis-ci.org/chirayu/bench.svg?branch=master)](https://travis-ci.org/chirayu/bench) [![GoDoc](http://godoc.org/github.com/chirayu/bench?status.png)](http://godoc.org/github.com/chirayu/bench)

**Bench** is a golang package to help speed up benchmarking projects. It can be used to do benchmark your database, queue or rest server. 

## Features

Bench can call a function **repeatedly** or at a **uniform rate**  with the specificied concurrency. For each call, the called function can captures metrics, which are summarized at the end of the run.

Captured metrics include:
* Timers. Use it to capture latency of functions, I/O bound calls.
* Counters. 

A sample run will spew similar output

```console
$ go run simple.go
Rate: 10 calls/sec, Duration: 5.00s, Concurrency: 2, Total runs: 50
  >>Timer: Latency
    5.0th percentile: 0.59ms
    50.0th percentile: 0.67ms
    99.9th percentile: 0.82ms
    100.0th percentile: 0.82ms
  >>Counter: Call Counter
    Value: 50
```

Bench provides a Context in each function call. Use it to

* Pick different between urls for each call
* Add it as a custom header for your Rest call
* Generate different DB data
* Store run specific values. Example connection pools.

## Usage

Bench is easy to integrate with your code. This code benchmarks a http server with a concurrency of 10 for 10 seconds by calling the `benchmarkMe` function continuously.

```go
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
	b := bench.NewBench(10, time.Second*10, 0, benchmarkMe)
	b.Run()
	fmt.Printf("%s", b)
```

Here, the same function is called at a rate of 100 per second.

```go
func main() {
	b := bench.NewBench(10, time.Second*10, 100, benchmarkMe)
	b.Run()
	fmt.Printf("%s", b)
```


## Upcoming features

Two features are in the offing.

1. Ability to capture multiple metrics during runtime. Apart from latency, a http client may want to record data size and error codes. 
2. Make requests at a uniform rate. This is useful when you want to see if you server can sustain a load of X per second. 
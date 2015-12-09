Bench is a golang package to speed up benchmarking projects.

Integration with existing code is easy.

```
func f(id int, iter int) time.Duration {
	time.Sleep(100 * time.Millisecond)
	return time.Duration(iter) * time.Millisecond
}

func main() {
	b := bench.NewBench(10, time.Second*2, f)
	b.Run()
	fmt.Printf("%s", b)
}
```
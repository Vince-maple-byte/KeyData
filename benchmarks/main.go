package main

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"slices"
	"time"

	database "github.com/Vince-maple-byte/KeyData/internals/db"
)

func main() {
	fmt.Println("Hello from the benchmark tests.")

	dir := os.TempDir()
	storageDir := filepath.Join(dir, "storage")
	walDir := filepath.Join(dir, "wal")
	os.Mkdir(storageDir, 0700)
	os.Mkdir(walDir, 0700)
	walFile := filepath.Join(walDir, "mem1.wal")

	//defer os.RemoveAll(dir)

	db, err := database.CreateDatabase(storageDir, walFile)

	if err != nil {
		fmt.Printf("not able to create the database:%v", err)
		return
	}

	durations, err := FullWriteBenchmark(10_000, 1, db)

	if err != nil {
		fmt.Printf("not able to commit the full write benchmark:%v", err)
		return
	}

	slices.Sort(durations)
	p50Index := int(math.Ceil(0.50*float64(len(durations)))) - 1
	p95Index := int(math.Ceil(0.95*float64(len(durations)))) - 1
	p99Index := int(math.Ceil(0.99*float64(len(durations)))) - 1

	var sum time.Duration
	for _, i := range durations {
		sum += i
	}
	average := (sum) / time.Duration(len(durations))

	fmt.Printf("For 10,000 write operations:\nAverage latency: %v\nP50 latency: %v\nP95 latency: %v\nP99 latency: %v",
		average, durations[p50Index], durations[p95Index], durations[p99Index])

	db.Close()
	os.RemoveAll(storageDir)
	os.RemoveAll(walDir)

	// (keys := make([]string, 10_000);

	// for i

	// durations, err = FullReadBenchmark(10_000, 10, db)

	// if err != nil {
	// 	fmt.Printf("not able to commit the full write benchmark:%v", err)
	// 	return
	// }

	// slices.Sort(durations)
	// p99Index = int(math.Ceil(0.99*float64(len(durations)))) - 1

	// fmt.Printf("For 10,000 write operations, run 10 times:\nAverage time of completion: %v\nFastest time: %v\nP99 time: %v",
	// 	durations[len(durations)/2], durations[0], durations[p99Index]))
}

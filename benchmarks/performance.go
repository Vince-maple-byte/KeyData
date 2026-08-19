package main

import (
	"fmt"
	"time"

	database "github.com/Vince-maple-byte/KeyData/internals/db"
)

func FullWriteBenchmark(numOfOperations, numOfRuns int, db *database.Database) ([]time.Duration, error) {
	durations := make([]time.Duration, numOfOperations)
	for range numOfRuns {

		for i := range numOfOperations {
			start := time.Now()
			key := fmt.Sprintf("%d", i)
			val := fmt.Sprintf("val_%d", i)

			ok, err := db.Put(key, val)

			if !ok {
				return durations, err
			}
			duration := time.Since(start)
			durations[i] = duration
		}

	}

	return durations, nil
}

// We are going to expect that the db is already populated with values.
func FullReadBenchmark(numOfOperations, numOfRuns int, db *database.Database, keysToRead []string) ([]time.Duration, error) {
	durations := make([]time.Duration, 0)
	for range numOfRuns {
		start := time.Now()
		for i := range numOfOperations {
			db.Get(keysToRead[i])
		}

		duration := time.Since(start)
		durations = append(durations, duration)
	}

	return durations, nil
}

func MixedOperationsBenchmark(numOfWriteOperations, numOfReadOperations, numOfRuns int, db *database.Database, keysToRead []string) ([]time.Duration, error) {
	durations := make([]time.Duration, 0)
	for range numOfRuns {
		start := time.Now()
		for i := range numOfWriteOperations {
			key := fmt.Sprintf("%d", i)
			val := fmt.Sprintf("val_%d", i)

			ok, err := db.Put(key, val)

			if !ok {
				return durations, err
			}
		}

		for i := range numOfReadOperations {
			_, err := db.Get(keysToRead[i])

			if err != nil {
				return durations, err
			}
		}

		duration := time.Since(start)
		durations = append(durations, duration)
	}

	return durations, nil
}

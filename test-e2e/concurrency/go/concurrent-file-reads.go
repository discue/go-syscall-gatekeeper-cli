package main

import (
	"fmt"
	"os"
	"sync"
)

func main() {
	var wg sync.WaitGroup
	numGoroutines := 2000
	results := make(chan int, numGoroutines) // Channel to receive goroutine IDs on success

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			f, err := os.OpenFile("/etc/resolv.conf", os.O_RDONLY, 0644)
			if err != nil {
				fmt.Printf("Goroutine %d failed: %s\n", i+1, err) // Report individual failure
				return
			}
			defer func() { _ = f.Close() }()

			// ... process file ...

			fmt.Printf("Goroutine %d finished successfully.\n", i+1)
			results <- i + 1 // Send goroutine ID on success
		}(i)
	}

	wg.Wait() // Wait for all goroutines to complete

	close(results) // Close the results channel to signal no more results

	fmt.Println("All goroutines launched. Processing results:")

	successCount := 0
	for id := range results { // Iterate over successful goroutine IDs
		fmt.Printf("Goroutine %d reported success.\n", id)
		successCount++
	}

	fmt.Printf("%d goroutines succeeded.\n", successCount) // Summarize successes

	if numGoroutines == successCount {
		fmt.Println("Good times. Goodbye.")
		os.Exit(0)
	} else {
		fmt.Println("Not all goroutines finished successfully. Thus, returning exit code 1")
		os.Exit(1)
	}
}

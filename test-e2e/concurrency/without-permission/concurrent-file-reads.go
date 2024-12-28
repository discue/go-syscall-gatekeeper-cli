package main

import (
	"fmt"
	"os"
	"sync"
)

func main() {
	var wg sync.WaitGroup
	numGoroutines := 2000 // Number of goroutines to launch
	numSuccessfulGoroutines := 25

	expectedFailures := make(chan int, numGoroutines) // Channel to receive goroutine IDs on success
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			f, err := os.OpenFile("/etc/resolv.conf", os.O_RDWR, 0644)
			if err != nil {
				fmt.Printf("Goroutine %d failed: %s\n", i+1, err) // Report individual failure
				return
			}
			defer f.Close()

			// ... process file ...

			fmt.Printf("Goroutine %d finished successfully.\n", i+1)
			expectedFailures <- i + 1 // Send goroutine ID on success
		}(i)
	}

	expectedSuccesses := make(chan int, numSuccessfulGoroutines) // Channel to receive goroutine IDs on success
	for i := 0; i < numSuccessfulGoroutines; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			f, err := os.OpenFile("/etc/resolv.conf", os.O_RDONLY, 0644)
			if err != nil {
				fmt.Printf("Goroutine %d failed: %s\n", i+1, err) // Report individual failure
				return
			}
			defer f.Close()

			bytes := make([]byte, 1024)
			_, err = f.Read(bytes)
			if err != nil {
				fmt.Printf("Goroutine %d failed: %s\n", i+1, err) // Report individual failure
				return
			}

			fmt.Printf("Goroutine %d read %s .\n", i+1, string(bytes))

			// ... process file ...
			fmt.Printf("Goroutine %d finished successfully.\n", i+1)
			expectedSuccesses <- i + 1 // Send goroutine ID on success
		}(i)
	}

	wg.Wait() // Wait for all goroutines to complete

	close(expectedFailures)  // Close the results channel to signal no more results
	close(expectedSuccesses) // Close the results channel to signal no more results

	fmt.Println("All goroutines launched. Processing results:")

	failures := 0
	for id := range expectedFailures { // Iterate over successful goroutine IDs
		fmt.Printf("Goroutine %d reported success.\n", id)
		failures++
	}

	successes := 0
	for id := range expectedSuccesses { // Iterate over successful goroutine IDs
		fmt.Printf("Goroutine %d reported success.\n", id)
		successes++
	}

	fmt.Printf("%d of %d goroutines we expected to fail succeeded.\n", failures, numGoroutines)               // Summarize successes
	fmt.Printf("%d of %d goroutines we expected to succeed succeeded.\n", successes, numSuccessfulGoroutines) // Summarize successes

	if numSuccessfulGoroutines == successes && failures == 0 {
		fmt.Println("Good times. Goodbye.")
		os.Exit(0)
	} else {
		fmt.Println("Returning exit code 1")
		os.Exit(1)
	}
}

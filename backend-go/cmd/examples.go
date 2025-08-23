package main

import (
	"fmt"
	"sync"
	"time"
)

//concurrency

func worker(id int) {
	fmt.Printf("Worker %d starting\n", id)

	time.Sleep(time.Second)
	fmt.Printf("Worker %d done\n", id)
}

func WorkGroupConcurrency() {
	var wg sync.WaitGroup

	for i := 1; i <= 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			worker(id)
		}(i)
	}
	wg.Wait()
	fmt.Println("Done")
}

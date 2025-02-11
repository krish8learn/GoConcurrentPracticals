// Write your answer here, and then test your code.
package malurlparse

// Your job is to convert MultiURLTime to run concurrently.

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"
)

// MutliURLTimes calls URLTime for every URL in URLs.
func MultiURLTime(urls []string) {
	var wg sync.WaitGroup
	wg.Add(len(urls))
	for _, url := range urls {
		go URLTime(url, &wg)
	}
	wg.Wait()
	fmt.Println("all done")
}

// URLTime checks how much time it takes url to respond.
func URLTime(url string, wg *sync.WaitGroup) {
	defer wg.Done()
	start := time.Now()

	resp, err := http.Get(url)
	if err != nil {
		log.Printf("error: %q - %s", url, err)
		return
	}
	if resp.StatusCode != http.StatusOK {
		log.Printf("error: %q - bad status - %s", url, resp.Status)
		return
	}
	// Read body
	_, err = io.Copy(io.Discard, resp.Body)
	if err != nil {
		log.Printf("error: %q - %s", url, err)
		return
	}

	duration := time.Since(start)
	log.Printf("info: %q - %v", url, duration)
}

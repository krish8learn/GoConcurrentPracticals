package main

import (
	"GoConcurrentPracticals/cache"
	"GoConcurrentPracticals/digitalsig"
	"GoConcurrentPracticals/fixedpool"
	"GoConcurrentPracticals/malurlparse"
	"GoConcurrentPracticals/movierecom"
	"context"
	"fmt"
	"log"
	"runtime"
	"sync"
	"time"
)

func main() {
	problem3()
}

func problem1() {
	start := time.Now()

	urls := []string{
		"http://localhost:8080/200",
		"http://localhost:8080/100",
		"http://localhost:8080/50",
	}

	malurlparse.MultiURLTime(urls)

	duration := time.Since(start)
	log.Printf("%d URLs in %v", len(urls), duration)
}

func problem2() {
	start := time.Now()

	inputs := []digitalsig.File{
		{
			Name:      "test1",
			Content:   []byte("krishendu.karmakar"),
			Signature: "signature",
		},
	}

	ok, bad, err := digitalsig.ValidateSigs(inputs)
	if err != nil {
		log.Fatal(err)
	}

	duration := time.Since(start)
	log.Printf("info: %d files in %v\n", len(ok)+len(bad), duration)
	log.Printf("ok: %v", ok)
	log.Printf("bad: %v", bad)
}

func problem3() {
	log.Printf("info: checking finish in time")
	ctx, cancel := context.WithTimeout(context.Background(), movierecom.BmvTime*2)
	defer cancel()

	mOK := movierecom.NextMovie(ctx, "ridley")
	log.Printf("info: got %+v", mOK)

	log.Printf("info: checking timeout")
	ctx, cancel = context.WithTimeout(context.Background(), movierecom.BmvTime/2)
	defer cancel()

	mTimeout := movierecom.NextMovie(ctx, "ridley")
	log.Printf("info: got %+v", mTimeout)
}

func problem4() {
	start := time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	n := runtime.GOMAXPROCS(0) // number of cores
	srcDir, destDir := "", ""
	err := fixedpool.CenterDir(ctx, srcDir, destDir, n)

	duration := time.Since(start)
	log.Printf("info: finished in %v (err=%v)", duration, err)
}

func problem5() {
	keyFmt := "key-%02d"
	keyName := func(i int) string { return fmt.Sprintf(keyFmt, i) }

	size := 5
	ttl := 10 * time.Millisecond
	log.Printf("info: creating cache: size=%d, ttl=%v", size, ttl)
	c, err := cache.New(size, ttl)
	if err != nil {
		log.Printf("error: can't create - %s", err)
		return
	}
	log.Printf("info: OK")

	log.Printf("info: checking TTL")
	key, val := keyName(1), 3
	c.Set(key, val)
	v, ok := c.Get(key)
	if !ok || v != val {
		log.Printf("error: %q: got %v (ok=%v)", key, v, ok)
		return
	}

	// Let key expire
	time.Sleep(2 * ttl)
	_, ok = c.Get(key)
	if ok {
		log.Printf("error: %q: got value after TTL", key)
		return
	}
	log.Printf("info: OK")

	log.Printf("info: checking overflow")
	n := size * 2
	for i := 0; i < n; i++ {
		c.Set(keyName(i), i)
	}
	_, ok = c.Get(keyName(1))
	if ok {
		log.Printf("error: %q: got value after overflow", key)
		return
	}
	_, ok = c.Get(keyName(n - 1))
	if !ok {
		log.Printf("error: %q: not found", key)
		return
	}
	log.Printf("info: OK")

	numGr := size * 3
	count := 1000
	log.Printf("info: checking concurrency (%d goroutines, %d loops each)", numGr, count)

	var wg sync.WaitGroup
	wg.Add(numGr)
	for i := 0; i < numGr; i++ {
		key := keyName(i)
		go func() {
			defer wg.Done()
			for i := 0; i < count; i++ {
				time.Sleep(time.Microsecond)
				c.Set(key, i)
			}
		}()
	}
	wg.Wait()
	log.Printf("info: OK")
}

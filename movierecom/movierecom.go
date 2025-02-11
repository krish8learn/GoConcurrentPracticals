package movierecom

import (
	"context"
	"time"
)

var (
	// Everybody loves "The Princess Bride"
	defaultMovie = Movie{
		ID:    "tt0093779",
		Title: "The Princess Bride",
	}
	// Time it takes for BestNextMovie to finish
	BmvTime = 50 * time.Millisecond
)

// Movie is a movie recommendation
type Movie struct {
	ID    string
	Title string
}

// BestNextMovie return the best move recommendation for a user
func BestNextMovie(user string) Movie {
	time.Sleep(BmvTime) // Simulate work

	// Don't change this, otherwise the test will fail
	return Movie{
		ID:    "tt0083658",
		Title: "Blade Runner",
	}
}

// NextMovie return BestNextMovie result if it finished before ctx expires, otherwise defaultMovie
func NextMovie(ctx context.Context, user string) Movie {
	// FIXME: You code goes here

	resChan := make(chan Movie, 1)

	go func() {
		resChan <- BestNextMovie(user)
	}()

	select {
	case m := <-resChan:
		return m
	case <-ctx.Done():
		return defaultMovie
	}
}

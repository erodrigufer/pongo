package limiter

import (
	"fmt"
	"net/http"
	"time"

	limiter "github.com/ulule/limiter/v3"
	mhttp "github.com/ulule/limiter/v3/drivers/middleware/stdlib"

	"github.com/ulule/limiter/v3/drivers/store/memory"
)

// TODO: finish
// NewRateLimiterMiddleware, creates a rate limiter middleware with
func NewRateLimiterMiddleware(limit int64, period string, extHandler http.Handler) (http.Handler, error) {
	// Define the period/time.Duration for the rate limiter.
	per, err := time.ParseDuration(period)
	if err != nil {
		return nil, fmt.Errorf("could not parse time duration correctly: %w", err)
	}
	// Define rate limiter.
	rate := limiter.Rate{
		Period: per,
		Limit:  limit,
	}

	// New in-memory store for the keys and values of the rate limiter.
	store := memory.NewStore()

	// Rate limiter instance with a given rate limit and in-memory store.
	rateLimiter := limiter.New(store, rate)

	// Create a new middleware with the limiter instance.
	// middleware := mhttp.NewMiddleware(rateLimiter)
	middleware := mhttp.Middleware{
		Limiter: rateLimiter,
		OnError: mhttp.DefaultErrorHandler,
		OnLimitReached: 
	}


	// TODO: remove this comments when the application works, they are here just
	// as a reference on how to use the Handler method within an HTTP server.
	// Launch a simple server.
	// http.Handle("/", middleware.Handler(http.HandlerFunc(index)))

	return middleware.Handler(extHandler), nil

}

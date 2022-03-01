package ratelimiter

// RateLimiter exposes an Acquire() method for obtaining a Rate Limit Token
type RateLimiter interface {
	Acquire() (*Token, error)
	Release(*Token)
}

// Config represents a rate limiter config object
type Config struct {
	// Limit determines how many rate limit tokens can be active at a time
	Limit int
}

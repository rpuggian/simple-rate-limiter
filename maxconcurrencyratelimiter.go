package ratelimiter

// NewMaxConcurrencyRateLimiter returns a max concurrency rate limiter
func NewMaxConcurrencyRateLimiter(conf *Config) (RateLimiter, error) {
	if conf.Limit <= 0 {
		return nil, ErrInvalidLimit
	}

	m := NewManager(conf)
	// max concurrency await function
	go func() {
		for {
			select {
			case <-m.inChan:
				m.tryGenerateToken()
			case t := <-m.releaseChan:
				m.releaseToken(t)
			}
		}
	}()
	return m, nil
}

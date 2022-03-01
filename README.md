<h2>Simple Rate Limiter</h2>


Simple Rate Limiter limit concurrency operations using a pool of tokens. 

Sample of usage: 
```go
  //Create ratelimiter instance, should be used as singleton to all operations
  r, err := ratelimiter.NewMaxConcurrencyRateLimiter(&ratelimiter.Config{
    //Limit of pool of tokens
		Limit: 3,
	})
  
  // Acquire a rate limit token, if the limit of operations was achieved the operation will be sleeping here.
  token, err := c.r.Acquire()
  
  //Perform Operation
  
  //Set the usage token free to go back to pool
  defer c.r.Release(token)
```

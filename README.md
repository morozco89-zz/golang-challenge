# First Todo: Enabling cache

The first thought that crossed my mind when I looked at the code was that there was no way to keep track of each price "age".

Then I realize maybe map[string]float64 was not the right data structure to represent the cache.

So I created a type (struct CachedPrice) to model the price and the time at which that price was retrieved from actual service.

This time will help me to compare against TransparentCache configured maxAge and determinate weather or not that price was still valid. 

# Second Todo: Parallelize GetPricesFor

Since GetPricesFor return type is []float64, this (for me) means that I should expect for each item price (passed as argument), a corresponding price at the same index.

So I thought about using a buffer with fixed size in order to expect each item price at the same index.

Then, in order to parallelize, I use go routines and a waiting group in order to wait for all routines. This means []float64 result slice is populated.

Finally, I check if there was an error in any of the request. If at least one failed, then it returns an empty []float64 with the first error.

When I ran the tests at this point, I notice this error `Golang fatal error: concurrent map read and map write` so I use a mutex in order to protect from concurrent writing the prices map from TransparentCache.

I wrote a test to probe that GetPricesFor returns error when any of the item codes service request return an error.

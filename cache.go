package sample1

import (
	"fmt"
	"sync"
	"time"
)

// PriceService is a service that we can use to get prices for the items
// Calls to this service are expensive (they take time)
type PriceService interface {
	GetPriceFor(itemCode string) (float64, error)
}

// TransparentCache is a cache that wraps the actual service
// The cache will remember prices we ask for, so that we don't have to wait on every call
// Cache should only return a price if it is not older than "maxAge", so that we don't get stale prices
type TransparentCache struct {
	actualPriceService PriceService
	maxAge             time.Duration
	prices             map[string]CachedPrice
}

type CachedPrice struct {
	time  time.Time
	price float64
}

func NewTransparentCache(actualPriceService PriceService, maxAge time.Duration) *TransparentCache {
	return &TransparentCache{
		actualPriceService: actualPriceService,
		maxAge:             maxAge,
		prices:             map[string]CachedPrice{},
	}
}

// GetPriceFor gets the price for the item, either from the cache or the actual service if it was not cached or too old
func (c *TransparentCache) GetPriceFor(itemCode string) (float64, error) {
	cachedPrice, ok := c.prices[itemCode]
	if ok {
		if time.Now().Sub(cachedPrice.time) < c.maxAge {
			return cachedPrice.price, nil
		}
	}
	price, err := c.actualPriceService.GetPriceFor(itemCode)
	if err != nil {
		return 0, fmt.Errorf("getting price from service : %v", err.Error())
	}

	mutex := sync.RWMutex{}

	mutex.Lock()
	c.prices[itemCode] = CachedPrice{
		time:  time.Now(),
		price: price,
	}
	mutex.Unlock()
	return price, nil
}

// GetPricesFor gets the prices for several items at once, some might be found in the cache, others might not
// If any of the operations returns an error, it should return an error as well
func (c *TransparentCache) GetPricesFor(itemCodes ...string) ([]float64, error) {
	// Use a waiting group to wait for each result
	var wg sync.WaitGroup
	var err error

	// Create a buffer with fixed length to accumulate results
	results := make([]float64, len(itemCodes))

	for i, itemCode := range itemCodes {
		wg.Add(1)
		go func(i int, code string) {
			var price float64
			var innerErr error

			defer wg.Done()

			price, innerErr = c.GetPriceFor(code)
			if err == nil {
				err = innerErr
			}

			// Since I'm writing to different memory addresses, I don't need to worry about race conditions
			results[i] = price
		}(i, itemCode)
	}

	wg.Wait()

	// Ignore other successful response if there was an error
	if err != nil {
		return []float64{}, err
	}

	return results, nil
}

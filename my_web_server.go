package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	libredis "github.com/go-redis/redis/v8"

	limiter "github.com/panii/limiter/v3"
	mhttp "github.com/panii/limiter/v3/drivers/middleware/stdlib"
	sredis "github.com/panii/limiter/v3/drivers/store/redis"
)

var START_TIME string

func main() {
	START_TIME = time.Now().Add(time.Hour * 8).Format("2006-01-02 15:04:05")
	
	indexHandler := indexLimiterHandler()
	http.Handle("/rate_check/do", indexHandler)
	http.HandleFunc("/rate_check/version", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("START_TIME", START_TIME)
		w.Write([]byte("0.1"))
	})

	fmt.Println("Server is running on port 9000...")
	log.Fatal(http.ListenAndServe(":9000", nil))
}

func index(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, err := w.Write([]byte(`allow`))
	if err != nil {
		log.Fatal(err)
	}
}

func indexLimiterHandler() http.Handler {

	secondRate := limiter.Rate{Id: "second"}
	minuteRate := limiter.Rate{Id: "minute"}
	hourRate := limiter.Rate{Id: "hour"}
	dayRate := limiter.Rate{Id: "day"}

	client := libredis.NewClient(&libredis.Options{
		Addr: "redis.rds.aliyuncs.com:6379",

		Password: "password",
		DB:       2,
	})

	// Create a store with the redis client.
	storeDay, err := sredis.NewStoreWithOptions(client, limiter.StoreOptions{
		Prefix:   "rate-limit-server-day",
		MaxRetry: 3,
	})
	if err != nil {
		log.Fatal(err)
	}

	storeSecond, err := sredis.NewStoreWithOptions(client, limiter.StoreOptions{
		Prefix:   "rate-limit-server-second",
		MaxRetry: 3,
	})
	if err != nil {
		log.Fatal(err)
	}

	storeMinute, err := sredis.NewStoreWithOptions(client, limiter.StoreOptions{
		Prefix:   "rate-limit-server-minute",
		MaxRetry: 3,
	})
	if err != nil {
		log.Fatal(err)
	}

	storeHour, err := sredis.NewStoreWithOptions(client, limiter.StoreOptions{
		Prefix:   "rate-limit-server-hour",
		MaxRetry: 3,
	})
	if err != nil {
		log.Fatal(err)
	}

	secondLimit := limiter.New(storeSecond, secondRate)
	hourLimit := limiter.New(storeHour, hourRate)
	dayLimit := limiter.New(storeDay, dayRate)
	minuteLimit := limiter.New(storeMinute, minuteRate)

	var handler http.Handler = http.HandlerFunc(index)

	mhttp.Secret = "-xxxxxxx"

	handler = mhttp.NewMiddleware(dayLimit).Handler(handler)
	handler = mhttp.NewMiddleware(hourLimit).Handler(handler)
	handler = mhttp.NewMiddleware(minuteLimit).Handler(handler)
	handler = mhttp.NewMiddleware(secondLimit).Handler(handler)

	return handler
}

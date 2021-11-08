package stdlib

import (
	bigcontext "context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/panii/limiter/v3"
)

var Secret string

// Middleware is the middleware for basic http.Handler.
type Middleware struct {
	Limiter        *limiter.Limiter
	OnError        ErrorHandler
	OnLimitReached LimitReachedHandler
	ExcludedKey    func(string) bool
}

// NewMiddleware return a new instance of a basic HTTP middleware.
func NewMiddleware(limiter *limiter.Limiter, options ...Option) *Middleware {
	middleware := &Middleware{
		Limiter:        limiter,
		OnError:        DefaultErrorHandler,
		OnLimitReached: DefaultLimitReachedHandler,
		ExcludedKey:    nil,
	}

	for _, option := range options {
		option.apply(middleware)
	}

	return middleware
}

// Handler handles a HTTP request.
func (middleware *Middleware) Handler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		keys, ok := query["key"]
		var key string
		if !ok || len(keys[0]) < 1 {
			middleware.OnLimitReached(w, r)
			return
		}
		hash := md5.Sum([]byte(keys[0]))
		key = hex.EncodeToString(hash[:])

		signs, ok := query["sign"]
		var sign string
		if !ok || len(signs[0]) < 1 {
			middleware.OnLimitReached(w, r)
			return
		}
		hash2 := md5.Sum([]byte(keys[0] + Secret))
		sign = hex.EncodeToString(hash2[:])

		if sign != signs[0] {
			middleware.OnLimitReached(w, r)
			return
		}

		//key := middleware.Limiter.GetIPKey(r)
		if middleware.ExcludedKey != nil && middleware.ExcludedKey(key) {
			h.ServeHTTP(w, r)
			return
		}
		ctx := r.Context()
		var err error

		rateId := middleware.Limiter.Rate.Id

		rateTemp := limiter.Rate{}

		// Add limit to context
		if rateId == "second" {
			limit1, ok := query["limitSecond"] // second limit
			var limit1Int int64
			if !ok || len(limit1[0]) < 1 {
				limit1Int = 0
			} else {
				limit1Int, err = strconv.ParseInt(limit1[0], 10, 64)
				if err != nil {
					limit1Int = 0
				}
			}
			rateTemp.Limit = limit1Int
		}

		if rateId == "minute" {
			limit2, ok := query["limitMinute"] // minute limit
			var limit2Int int64
			if !ok || len(limit2[0]) < 1 {
				limit2Int = 0
			} else {
				limit2Int, err = strconv.ParseInt(limit2[0], 10, 64)
				if err != nil {
					limit2Int = 0
				}
			}
			rateTemp.Limit = limit2Int
		}

		if rateId == "hour" {
			limit3, ok := query["limitHour"] // hour limit
			var limit3Int int64
			if !ok || len(limit3[0]) < 1 {
				limit3Int = 0
			} else {
				limit3Int, err = strconv.ParseInt(limit3[0], 10, 64)
				if err != nil {
					limit3Int = 0
				}
			}
			rateTemp.Limit = limit3Int
		}

		if rateId == "day" {
			limit4, ok := query["limitDay"] // day limit
			var limit4Int int64
			if !ok || len(limit4[0]) < 1 {
				limit4Int = 0
			} else {
				limit4Int, err = strconv.ParseInt(limit4[0], 10, 64)
				if err != nil {
					limit4Int = 0
				}
			}
			rateTemp.Limit = limit4Int
		}

		// Add time to context
		if rateId == "second" {
			time1, ok := query["periodSecond"] // second
			var time1Int int64
			if !ok || len(time1[0]) < 1 {
				rateTemp.Period = time.Second * 1
			} else {
				time1Int, err = strconv.ParseInt(time1[0], 10, 64)
				if err != nil {
					time1Int = 0
				}
				rateTemp.Period = time.Second * time.Duration(time1Int)
			}
		}

		if rateId == "minute" {
			time2, ok := query["periodMinute"] // minute
			var time2Int int64
			if !ok || len(time2[0]) < 1 {
				rateTemp.Period = time.Minute * 1
			} else {
				time2Int, err = strconv.ParseInt(time2[0], 10, 64)
				if err != nil {
					time2Int = 0
				}
				rateTemp.Period = time.Minute * time.Duration(time2Int)
			}
		}

		if rateId == "hour" {
			time3, ok := query["periodHour"] // hour
			var time3Int int64
			if !ok || len(time3[0]) < 1 {
				rateTemp.Period = time.Hour * 1
			} else {
				time3Int, err = strconv.ParseInt(time3[0], 10, 64)
				if err != nil {
					time3Int = 0
				}
				rateTemp.Period = time.Hour * time.Duration(time3Int)
			}
		}

		if rateId == "day" {
			time4, ok := query["periodDay"] // day
			var time4Int int64
			if !ok || len(time4[0]) < 1 {
				rateTemp.Period = time.Hour * 24 * 1
			} else {
				time4Int, err = strconv.ParseInt(time4[0], 10, 64)
				if err != nil {
					time4Int = 0
				}
				rateTemp.Period = time.Hour * 24 * time.Duration(time4Int)
			}
		}
		// do not check
		if rateTemp.Limit == 0 || rateTemp.Period == 0 {
			h.ServeHTTP(w, r)
			return
		}

		fmt.Println("rateId", rateId)
		fmt.Println("rateTemp.Limit", rateTemp.Limit)
		fmt.Println("rateTemp.Period", rateTemp.Period)
		ctx = bigcontext.WithValue(ctx, "rateTemp", rateTemp)

		context, err := middleware.Limiter.Get(ctx, key)
		if err != nil {
			middleware.OnError(w, r, err)
			return
		}

		if rateId == "second" {
			w.Header().Add("X-RateLimit-Limit-Second", strconv.FormatInt(context.Limit, 10))
			w.Header().Add("X-RateLimit-Remaining-Second", strconv.FormatInt(context.Remaining, 10))
			w.Header().Add("X-RateLimit-Reset-Second", strconv.FormatInt(context.Reset, 10))
		} else if rateId == "minute" {
			w.Header().Add("X-RateLimit-Limit-Minute", strconv.FormatInt(context.Limit, 10))
			w.Header().Add("X-RateLimit-Remaining-Minute", strconv.FormatInt(context.Remaining, 10))
			w.Header().Add("X-RateLimit-Reset-Minute", strconv.FormatInt(context.Reset, 10))
		} else if rateId == "hour" {
			w.Header().Add("X-RateLimit-Limit-Hour", strconv.FormatInt(context.Limit, 10))
			w.Header().Add("X-RateLimit-Remaining-Hour", strconv.FormatInt(context.Remaining, 10))
			w.Header().Add("X-RateLimit-Reset-Hour", strconv.FormatInt(context.Reset, 10))
		} else if rateId == "day" {
			w.Header().Add("X-RateLimit-Limit-Day", strconv.FormatInt(context.Limit, 10))
			w.Header().Add("X-RateLimit-Remaining-Day", strconv.FormatInt(context.Remaining, 10))
			w.Header().Add("X-RateLimit-Reset-Day", strconv.FormatInt(context.Reset, 10))
		}

		if context.Reached {
			middleware.OnLimitReached(w, r)
			return
		}

		h.ServeHTTP(w, r)
	})
}

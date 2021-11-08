package stdlib

import (
	bigcontext "context"
	"crypto/md5"
	"encoding/hex"
	"net/http"
	"strconv"

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
        
        oldLimit := middleware.Limiter.Rate.Limit

		// Add limit to context
        if middleware.Limiter.Rate.Limit == 1 {
            limit1, ok := query["limitSecond"] // second limit
            var limit1Int int64
            if !ok || len(limit1[0]) < 1 {
                middleware.Limiter.Rate.Limit = 0
            }
            limit1Int, err = strconv.ParseInt(limit1[0], 10, 64)
            if err != nil {
                limit1Int = 0
            }
            middleware.Limiter.Rate.Limit = limit1Int
            // ctx = bigcontext.WithValue(ctx, "limit1Int", limit1Int)
        }

        if middleware.Limiter.Rate.Limit == 2 {
            limit2, ok := query["limitMinute"] // minute limit
            var limit2Int int64
            if !ok || len(limit2[0]) < 1 {
                middleware.Limiter.Rate.Limit = 0
            }
            limit2Int, err = strconv.ParseInt(limit2[0], 10, 64)
            if err != nil {
                limit2Int = 0
            }
            middleware.Limiter.Rate.Limit = limit2Int
        }

        if middleware.Limiter.Rate.Limit == 3 {
            limit3, ok := query["limitHour"] // hour limit
            var limit3Int int64
            if !ok || len(limit3[0]) < 1 {
                middleware.Limiter.Rate.Limit = 0
            }
            limit3Int, err = strconv.ParseInt(limit3[0], 10, 64)
            if err != nil {
                limit3Int = 0
            }
            middleware.Limiter.Rate.Limit = limit3Int
        }

        if middleware.Limiter.Rate.Limit == 4 {
            limit4, ok := query["limitDay"] // day limit
            var limit4Int int64
            if !ok || len(limit4[0]) < 1 {
                middleware.Limiter.Rate.Limit = 0
            } else {
                limit4Int, err = strconv.ParseInt(limit4[0], 10, 64)
                if err != nil {
                    limit4Int = 0
                }
                middleware.Limiter.Rate.Limit = limit4Int
            }
        }
        
        // Add time to context
        if middleware.Limiter.Rate.Limit == 1 {
            time1, ok := query["periodSecond"] // second
            var time1Int int64
            if !ok || len(time1[0]) < 1 {
                
            } else {
                time1Int, err = strconv.ParseInt(time1[0], 10, 64)
                if err != nil {
                    time1Int = 0
                }
                middleware.Limiter.Rate.Period = middleware.Limiter.Rate.Period * time1Int
            }
        }

        if middleware.Limiter.Rate.Limit == 2 {
            time2, ok := query["periodMinute"] // minute
            var time2Int int64
            if !ok || len(time2[0]) < 1 {
                
            } else {
                time2Int, err = strconv.ParseInt(time2[0], 10, 64)
                if err != nil {
                    time2Int = 0
                }
                middleware.Limiter.Rate.Period = middleware.Limiter.Rate.Period * time2Int
            }
        }

        if middleware.Limiter.Rate.Limit == 3 {
            time3, ok := query["periodHour"] // hour
            var time3Int int64
            if !ok || len(time3[0]) < 1 {
                
            } else {
                time3Int, err = strconv.ParseInt(time3[0], 10, 64)
                if err != nil {
                    time3Int = 0
                }
                middleware.Limiter.Rate.Period = middleware.Limiter.Rate.Period * time3Int
            }
        }

        if middleware.Limiter.Rate.Limit == 4 {
            time4, ok := query["periodDay"] // day
            var time4Int int64
            if !ok || len(time4[0]) < 1 {
                
            } else {
                time4Int, err = strconv.ParseInt(time4[0], 10, 64)
                if err != nil {
                    time4Int = 0
                }
                middleware.Limiter.Rate.Period = middleware.Limiter.Rate.Period * time4Int
            }
        }
        
        // do not check
        if middleware.Limiter.Rate.Limit == 0 || middleware.Limiter.Rate.Period == 0 {
			h.ServeHTTP(w, r)
			return
		}

		context, err := middleware.Limiter.Get(ctx, key)
		if err != nil {
			middleware.OnError(w, r, err)
			return
		}

		if oldLimit == 1 {
			w.Header().Add("X-RateLimit-LimitSecond", strconv.FormatInt(context.Limit, 10))
			w.Header().Add("X-RateLimit-RemainingSecond", strconv.FormatInt(context.Remaining, 10))
			w.Header().Add("X-RateLimit-ResetSecond", strconv.FormatInt(context.Reset, 10))
		} else if oldLimit == 2 {
			w.Header().Add("X-RateLimit-LimitMinute", strconv.FormatInt(context.Limit, 10))
			w.Header().Add("X-RateLimit-RemainingMinute", strconv.FormatInt(context.Remaining, 10))
			w.Header().Add("X-RateLimit-ResetMinute", strconv.FormatInt(context.Reset, 10))
		} else if oldLimit == 3 {
			w.Header().Add("X-RateLimit-LimitHour", strconv.FormatInt(context.Limit, 10))
			w.Header().Add("X-RateLimit-RemainingHour", strconv.FormatInt(context.Remaining, 10))
			w.Header().Add("X-RateLimit-ResetHour", strconv.FormatInt(context.Reset, 10))
		} else if oldLimit == 4 {
			w.Header().Add("X-RateLimit-LimitDay", strconv.FormatInt(context.Limit, 10))
			w.Header().Add("X-RateLimit-RemainingDay", strconv.FormatInt(context.Remaining, 10))
			w.Header().Add("X-RateLimit-ResetDay", strconv.FormatInt(context.Reset, 10))
		}

		if context.Reached {
			middleware.OnLimitReached(w, r)
			return
		}

		h.ServeHTTP(w, r)
	})
}

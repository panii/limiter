package stdlib

import (
	bigcontext "context"
	"crypto/md5"
	"encoding/hex"
	"net/http"
	"strconv"

	"github.com/panii/limiter/v3"
)

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
		hash2 := md5.Sum([]byte(keys[0] + "-asdfjlijlxxxlkjbbbgyugiaaaoiuoccchiu...yuiyuifkfcom"))
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

		// Add data to context
		limit1, ok := query["limit1"] // second limit
		var limit1Int int64
		var err error
		if !ok || len(limit1[0]) < 1 {
			middleware.OnLimitReached(w, r)
			return
		}
		limit1Int, err = strconv.ParseInt(limit1[0], 10, 64)
		if err != nil {
			limit1Int = 0
		}
		ctx := bigcontext.WithValue(r.Context(), "limit1Int", limit1Int)

		limit2, ok := query["limit2"] // minute limit
		var limit2Int int64
		if !ok || len(limit2[0]) < 1 {
			middleware.OnLimitReached(w, r)
			return
		}
		limit2Int, err = strconv.ParseInt(limit2[0], 10, 64)
		if err != nil {
			limit2Int = 0
		}
		ctx = bigcontext.WithValue(ctx, "limit2Int", limit2Int)

		limit3, ok := query["limit3"] // hour limit
		var limit3Int int64
		if !ok || len(limit3[0]) < 1 {
			middleware.OnLimitReached(w, r)
			return
		}
		limit3Int, err = strconv.ParseInt(limit3[0], 10, 64)
		if err != nil {
			limit3Int = 0
		}
		ctx = bigcontext.WithValue(ctx, "limit3Int", limit3Int)

		limit4, ok := query["limit4"] // day limit
		var limit4Int int64
		if !ok || len(limit4[0]) < 1 {
			middleware.OnLimitReached(w, r)
			return
		}
		limit4Int, err = strconv.ParseInt(limit4[0], 10, 64)
		if err != nil {
			limit4Int = 0
		}
		ctx = bigcontext.WithValue(ctx, "limit4Int", limit4Int)

		context, err := middleware.Limiter.Get(ctx, key)
		if err != nil {
			middleware.OnError(w, r, err)
			return
		}

		if limit1Int == context.Limit {
			w.Header().Add("X-RateLimit-Limit1", strconv.FormatInt(context.Limit, 10))
			w.Header().Add("X-RateLimit-Remaining1", strconv.FormatInt(context.Remaining, 10))
			w.Header().Add("X-RateLimit-Reset1", strconv.FormatInt(context.Reset, 10))
		} else if limit2Int == context.Limit {
			w.Header().Add("X-RateLimit-Limit2", strconv.FormatInt(context.Limit, 10))
			w.Header().Add("X-RateLimit-Remaining2", strconv.FormatInt(context.Remaining, 10))
			w.Header().Add("X-RateLimit-Reset2", strconv.FormatInt(context.Reset, 10))
		} else if limit3Int == context.Limit {
			w.Header().Add("X-RateLimit-Limit3", strconv.FormatInt(context.Limit, 10))
			w.Header().Add("X-RateLimit-Remaining3", strconv.FormatInt(context.Remaining, 10))
			w.Header().Add("X-RateLimit-Reset3", strconv.FormatInt(context.Reset, 10))
		} else if limit4Int == context.Limit {
			w.Header().Add("X-RateLimit-Limit4", strconv.FormatInt(context.Limit, 10))
			w.Header().Add("X-RateLimit-Remaining4", strconv.FormatInt(context.Remaining, 10))
			w.Header().Add("X-RateLimit-Reset4", strconv.FormatInt(context.Reset, 10))
		}

		if context.Reached {
			middleware.OnLimitReached(w, r)
			return
		}

		h.ServeHTTP(w, r)
	})
}

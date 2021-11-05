package stdlib

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
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
		keys, ok := r.URL.Query()["key"]
		var key string
		if !ok || len(keys[0]) < 1 {
			middleware.OnError(w, r, errors.New("empty key"))
		}
		hash := md5.Sum([]byte(keys[0]))
		key = hex.EncodeToString(hash[:])

		//key := middleware.Limiter.GetIPKey(r)
		if middleware.ExcludedKey != nil && middleware.ExcludedKey(key) {
			h.ServeHTTP(w, r)
			return
		}

		context, err := middleware.Limiter.Get(r.Context(), key)
		if err != nil {
			middleware.OnError(w, r, err)
			return
		}

		w.Header().Add("X-RateLimit-Limit", strconv.FormatInt(context.Limit, 10))
		w.Header().Add("X-RateLimit-Remaining", strconv.FormatInt(context.Remaining, 10))
		w.Header().Add("X-RateLimit-Reset", strconv.FormatInt(context.Reset, 10))

		if context.Reached {
			middleware.OnLimitReached(w, r)
			return
		}

		h.ServeHTTP(w, r)
	})
}

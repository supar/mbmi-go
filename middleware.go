package main

import (
	"context"
	"net/http"
	"net/http/httputil"
	"strings"
)

type WrapHandler func(http.Handler) http.Handler

// Wrap router handler with middlewares
func Middlewares(h http.Handler, m ...WrapHandler) http.Handler {
	for _, mw := range m {
		h = mw(h)
	}

	return h
}

// Take JWT from Authorization header
// Parse token string and validate it if path is not /login
func JWT(log LogIface) WrapHandler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var (
				ctx context.Context
				err error

				id = r.Context().Value("Id").(string)
				tk = NewToken([]byte("secret"))
			)

			if t := r.Header.Get("Authorization"); t != "" {
				tk.JWT = strings.TrimPrefix(t, "Bearer ")

				if err = tk.Parse(); err != nil {
					log.Error("%s: Can't parse token: %s", id, err.Error())
				}
			}

			ctx = context.WithValue(r.Context(), "Token", tk)
			r = r.WithContext(ctx)

			next.ServeHTTP(w, r)
		})
	}
}

//  Set request id
func RequestId() WrapHandler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var (
				id  = RandStringId(12)
				ctx = context.WithValue(r.Context(), "Id", id)
			)

			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		})
	}
}

// Print request data
func verbose(log LogIface) WrapHandler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var (
				id      = r.Context().Value("Id").(string)
				dump, _ = httputil.DumpRequest(r, true)
			)

			log.Notice("%s: %s, %s: %s %s", id, r.RemoteAddr, r.UserAgent(), r.Method, r.URL.RequestURI())
			log.Debug("%s: %s", id, dump)

			next.ServeHTTP(w, r)

			log.Debug("%s: Finished", id)
		})
	}
}

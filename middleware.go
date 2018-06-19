package main

import (
	"context"
	"mbmi-go/models"
	"net/http"
	"net/http/httputil"
	"strings"
)

// The key type is unexported to prevent collisions with context keys defined in
// other packages.
type key int

const (
	requestIDKey key = iota
	secretKey
)

// WrapHandler represents type http.Handler
type WrapHandler func(http.Handler) http.Handler

// Middlewares returns router handler with middlewares tree
func Middlewares(h http.Handler, m ...WrapHandler) http.Handler {
	for _, mw := range m {
		h = mw(h)
	}

	return h
}

// JWT parses http header Authorization
// Parse token string and validate it if path is not /login
func JWT(secret string, log LogIface) WrapHandler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var (
				ctx context.Context
				err error
				tk  *Token

				tSecret = secret
				id      = r.Context().Value("Id").(string)
			)

			if s := r.Context().Value(secretKey); s != nil {
				if s.(string) == "" {
					log.Error("Empty secret, using common secret")
				} else {
					tSecret = s.(string)
				}
			}

			tk = NewToken([]byte(tSecret))

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

// ApplicationToken parses http header Application-Token and identify user secret
// by token
func ApplicationToken(env Enviroment) WrapHandler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var (
				ctx   context.Context
				err   error
				model []*models.User
				flt   models.FilterIface

				id = r.Context().Value("Id").(string)
			)

			if t := r.Header.Get("Application-Token"); t != "" {
				env.Debug("Found header Application-Token: %s", t)

				flt = models.NewFilter().
					Where("token", t).
					Where("manager", 1)

				model, _, err = env.Users(flt, false)

				if l := len(model); err != nil || l != 1 {
					if err != nil {
						env.Error("%s: %s", id, err.Error())
					} else {
						env.Error("%s: Unknown user with token(%s)", id, t)
					}
				} else {
					ctx = context.WithValue(r.Context(), secretKey, model[0].Secret())
					r = r.WithContext(ctx)
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

//  RequestId adds unique string to the request
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

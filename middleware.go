package main

import (
	"context"
	"net/http"
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
				//claims TokenClaims
				ctx  context.Context
				data []byte
				err  error
				id   string
				res  ResponseIface
				t    string
				tk   *Token
			)

			id = r.Context().Value("Id").(string)
			tk = NewToken([]byte("secret"))

			if t = r.Header.Get("Authorization"); t != "" {
				tk.JWT = strings.TrimPrefix(t, "Bearer ")

				if err = tk.Parse(); err != nil {
					log.Error("%s: Can't parse token: %s", id, err.Error())
				}
			}

			if r.URL.Path != "/login" {
				if !tk.Valid() {
					log.Error("%s: Token is nil or not valid", id)

					res = NewResponseError(401, nil)
					http.Error(w, "", res.Status())
					data, _ = res.Get()
					w.Write(data)

					return
				}

				log.Debug("%s: Token is valid", id)
			}

			ctx = context.WithValue(r.Context(), "Token", tk)
			r = r.WithContext(ctx)

			next.ServeHTTP(w, r)
		})
	}
}

// Send static files to the client
func Assets(dir string) WrapHandler {
	return func(next http.Handler) http.Handler {
		var (
			h http.Handler
		)

		if dir != "" {
			h = http.FileServer(http.Dir(dir))
		}

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var (
				t string
			)

			if t = r.Header.Get("Accept"); t == "application/json" {
				next.ServeHTTP(w, r)
				return
			}

			if h != nil {
				h.ServeHTTP(w, r)
				return
			}

			http.Error(w, http.StatusText(406), 406)
		})
	}
}

//  Set request id
func RequestId(log LogIface) WrapHandler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var (
				id  = RandStringId(12)
				ctx = context.WithValue(context.Background(), "Id", id)
			)

			r = r.WithContext(ctx)

			log.Notice("%s: %s, %s: %s %s", id, r.RemoteAddr, r.UserAgent(), r.Method, r.URL.RequestURI())
			log.Debug("%s: #%v", id, r.Header)

			next.ServeHTTP(w, r)

			log.Debug("%s: Finished", id)
		})
	}
}

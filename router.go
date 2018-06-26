package main

import (
	"context"
	"encoding/json"
	"github.com/gorilla/schema"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"strings"
)

// Shortcut
type routerParams interface {
	ByName(string) string
}

// Router extends basic httprouter
type Router struct {
	// Наследуемый маршрутизатор
	*httprouter.Router
}

// NewRouter creates instance of the Router
func NewRouter() *Router {
	return &Router{
		Router: httprouter.New(),
	}
}

// Handle requests httprouter handle function to backward default Handle from net/http package
// Put context to the request
func (r *Router) Handle(method, path string, handle http.HandlerFunc) {
	this.Router.Handle(method, path, func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		var (
			ctx = context.WithValue(r.Context(), "Params", p)
		)

		r = r.WithContext(ctx)
		handle(w, r)
	})
}

func parseFormTo(r *http.Request, v interface{}) (err error) {
	var (
		jsonDecoder   *json.Decoder
		schemaDecoder *schema.Decoder
	)

	if err = r.ParseForm(); err != nil {
		return
	}

	if t := r.Header.Get("Content-Type"); strings.Contains(t, "application/json") {
		defer r.Body.Close()

		jsonDecoder = json.NewDecoder(r.Body)
		return jsonDecoder.Decode(&v)
	}

	schemaDecoder = schema.NewDecoder()
	return schemaDecoder.Decode(v, r.PostForm)
}

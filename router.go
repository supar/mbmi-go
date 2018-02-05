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

// Extend basic httprouter
type Router struct {
	// Наследуемый маршрутизатор
	*httprouter.Router
}

// Create new router
func NewRouter() *Router {
	return &Router{
		Router: httprouter.New(),
	}
}

// Override httprouter handle function to backward default Handle from bet/http package
// Put context to the request
func (this *Router) Handle(method, path string, handle http.HandlerFunc) {
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
		json_decoder   *json.Decoder
		schema_decoder *schema.Decoder
	)

	if err = r.ParseForm(); err != nil {
		return
	}

	if t := r.Header.Get("Content-Type"); strings.Contains(t, "application/json") {
		defer r.Body.Close()

		json_decoder = json.NewDecoder(r.Body)
		return json_decoder.Decode(&v)
	}

	schema_decoder = schema.NewDecoder()
	return schema_decoder.Decode(v, r.PostForm)
}

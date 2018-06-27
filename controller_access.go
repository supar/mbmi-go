package main

import (
	"mbmi-go/models"
	"net/http"
)

// Accesses returns the list of blocked domains or ip addresses
func Accesses(r *http.Request, env Enviroment) ResponseIface {
	var (
		count uint64
		err   error
		resp  *Response
		t     []*models.Access

		flt = models.NewFilter()
		id  = r.Context().Value("Id")
	)

	if err = r.ParseForm(); err != nil {
		env.Error("%s, %s", id, err.Error())

		return NewResponse(&Error{
			Code:    500,
			Message: "cannot parse form data",
			Title:   http.StatusText(500),
		})
	}

	if _, ok := r.Form["sort"]; ok {
		flt.Order(r.Form.Get("sort"), false)
	}

	// Apply page limitation
	helperLimit(r, flt)

	if t, count, err = env.Accesses(flt, true); err != nil {
		env.Error("%s: %s", id, err.Error())

		if err == models.ErrFilterArgument {
			return NewResponse(err)
		}

		return NewResponse(&Error{
			Code:    500,
			Message: http.StatusText(500),
			Title:   http.StatusText(500),
		})
	}

	resp = NewResponse(t)
	resp.Count = count

	return resp
}

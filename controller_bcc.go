package main

import (
	"mbmi-go/models"
	"net/http"
	"strconv"
)

// Get bcc record by id
func Bccs(r *http.Request, env Enviroment) ResponseIface {
	var (
		count  uint64
		err    error
		params routerParams
		bid    int64
		b      []*models.BccItem

		cnt = true
		flt = models.NewFilter()
		id  = r.Context().Value("Id")
	)

	params = r.Context().Value("Params").(routerParams)

	if params.IsSet("bid") {
		bid, err = strconv.ParseInt(params.ByName("bid"), 10, 32)

		if err != nil || bid < 1 {
			if err != nil {
				env.Error("%s: %s", id, err.Error())
			}

			return NewResponse(&Error{
				Code:    404,
				Message: "empty bcc id",
				Title:   http.StatusText(404),
			})
		}

		cnt = false
		flt.Where("id", bid)
	} else {
		if v := r.FormValue("query"); v != "" {
			flt.Where("search", "%"+v+"%")
		}
	}

	if srt := r.FormValue("sort"); srt != "" {
		dir := true
		if r.FormValue("dir") == "desc" {
			dir = false
		}

		flt.Order(srt, dir)
	}

	if cnt {
		// Apply page limitation
		helperLimit(r, flt)
	}

	if b, count, err = env.Bccs(flt, cnt); err != nil {
		env.Error("%s: %s", id, err.Error())

		return NewResponse(&Error{
			Code:    500,
			Message: "Cannot fetch bcc item from database",
			Title:   http.StatusText(500),
		})
	}

	if l := len(b); l > 0 {
		if l == 1 {
			return NewResponse(b[0])
		}

		resp := NewResponse(b)
		resp.Count = count
		return resp
	}

	env.Error("%s: Can't find bcc item with id=(%d)", id, bid)

	return NewResponse(&Error{
		Code:    404,
		Message: http.StatusText(404),
		Title:   http.StatusText(404),
	})
}

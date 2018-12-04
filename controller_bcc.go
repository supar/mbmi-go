package main

import (
	"errors"
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

func SetBcc(r *http.Request, env Enviroment) ResponseIface {
	var (
		err error

		form   = models.BccItem{}
		id     = r.Context().Value("Id")
		params = r.Context().Value("Params").(routerParams)
	)

	if err = parseFormTo(r, &form); err != nil {
		env.Error("%s, %#v, %s", id, r.PostForm, err.Error())

		return NewResponse(&Error{
			Code:    500,
			Message: "cannot parse form data",
			Title:   http.StatusText(500),
		})
	}

	// Validate Sender
	if form.Sender != "" {
		if _, _, err = form.Sender.Split(); err != nil {
			env.Error("%s: %s", id, err.Error())

			return NewResponse(&Error{
				Code:    500,
				Message: err.Error(),
				Title:   http.StatusText(500),
			})
		}
	}

	// Validate Recipient
	if form.Recipient != "" {
		if _, _, err = form.Recipient.Split(); err != nil {
			env.Error("%s: %s", id, err.Error())

			return NewResponse(&Error{
				Code:    500,
				Message: err.Error(),
				Title:   http.StatusText(500),
			})
		}
	}

	if form.Sender == "" && form.Recipient == "" {
		return NewResponse(&Error{
			Code:    500,
			Message: "Sender or Recipient required",
			Title:   http.StatusText(500),
		})
	}

	// Validate Copy
	if _, _, err = form.Copy.Split(); err != nil {
		env.Error("%s, %s", id, err.Error())

		return NewResponse(&Error{
			Code:    500,
			Message: err.Error(),
			Title:   http.StatusText(500),
		})
	}

	if r.Method == "PUT" {
		var bid int64

		// Check
		if bid, err = strconv.ParseInt(params.ByName("bid"), 10, 64); err != nil || bid < 1 {
			if err == nil {
				err = errors.New("Invalid record id")
			}

			env.Error("%s: %s (id=%d)", id, err.Error(), bid)

			return NewResponse(&Error{
				Code:    500,
				Message: err.Error(),
				Title:   http.StatusText(500),
			})
		}

		form.ID = bid
	} else {
		form.ID = 0
	}

	env.Debug("%s: Bcc item data is valid", id)

	if err = env.SetBcc(&form); err != nil {
		env.Error("%s: %s", id, err.Error())

		return NewResponse(&Error{
			Code:    500,
			Message: "Cannot save bcc data",
			Title:   http.StatusText(500),
		})
	}

	return NewResponse(nil)
}

func DelBcc(r *http.Request, env Enviroment) ResponseIface {
	var (
		bid int64
		err error

		id     = r.Context().Value("Id")
		params = r.Context().Value("Params").(routerParams)
	)

	// Check
	if bid, err = strconv.ParseInt(params.ByName("bid"), 10, 64); err != nil || bid < 1 {
		if err == nil {
			err = errors.New("Invalid record id")
		}

		env.Error("%s: %s (id=%d)", id, err.Error(), bid)

		return NewResponse(&Error{
			Code:    500,
			Message: err.Error(),
			Title:   http.StatusText(500),
		})
	}

	if err = env.DelBcc(bid); err != nil {
		env.Error("%s: %s", id, err.Error())

		return NewResponse(&Error{
			Code:    500,
			Message: "Cannot remove bcc item data",
			Title:   http.StatusText(500),
		})
	}

	return NewResponse(nil)
}

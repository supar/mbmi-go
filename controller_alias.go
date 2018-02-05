package main

import (
	"errors"
	"mbmi-go/models"
	"net/http"
	"strconv"
)

// Get Aliases list
func Aliases(r *http.Request, env Enviroment) ResponseIface {
	var (
		count uint64
		err   error
		resp  *Response
		a     []*models.Alias

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

	if _, ok := r.Form["alias"]; ok {
		flt.Where("alias", r.Form.Get("alias"))
	}

	if _, ok := r.Form["recipient"]; ok {
		flt.Where("recipient", r.Form.Get("recipient")+"%")
	}

	if g := r.Context().Value("Group"); g != nil {
		flt.Group(g.(string))
	} else {
		// Apply page limitation
		helperLimit(r, flt)
	}

	if a, count, err = env.Aliases(flt, true); err != nil {
		env.Error("%s: %s", id, err.Error())

		return NewResponse(&Error{
			Code:    500,
			Message: "Cannot fetch aliases from database",
			Title:   http.StatusText(500),
		})
	}

	resp = NewResponse(a)
	resp.Count = count

	return resp
}

// Get Alias by id
func Alias(r *http.Request, env Enviroment) ResponseIface {
	var (
		err    error
		params routerParams
		aid    int64
		m      []*models.Alias

		flt = models.NewFilter()
		id  = r.Context().Value("Id")
	)

	params = r.Context().Value("Params").(routerParams)

	if str := params.ByName("aid"); str != "" {
		if aid, err = strconv.ParseInt(str, 10, 32); err != nil {
			env.Error("%s: %s", id, err.Error())
		}
	}

	if aid < 1 {
		return NewResponse(&Error{
			Code:    404,
			Message: "empty transport id",
			Title:   http.StatusText(404),
		})
	}

	flt.Where("id", aid)

	if m, _, err = env.Aliases(flt, false); err != nil {
		env.Error("%s: %s", id, err.Error())

		return NewResponse(&Error{
			Code:    500,
			Message: "Cannot fetch alias from database",
			Title:   http.StatusText(500),
		})
	}

	if l := len(m); l == 1 {
		return NewResponse(m[0])
	}

	env.Error("%s: Can't find alias with id=(%d)", id, aid)

	return NewResponse(&Error{
		Code:    404,
		Message: http.StatusText(404),
		Title:   http.StatusText(404),
	})
}

func SetAlias(r *http.Request, env Enviroment) ResponseIface {
	var (
		aid int64
		err error

		form   = models.Alias{}
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

	// Validate Recipient
	if _, _, err = form.Recipient.Split(); err != nil {
		env.Error("%s: %s", id, err.Error())

		return NewResponse(&Error{
			Code:    500,
			Message: err.Error(),
			Title:   http.StatusText(500),
		})
	}

	// Validate Alias
	if _, _, err = form.Alias.Split(); err != nil {
		env.Error("%s, %s", id, err.Error())

		return NewResponse(&Error{
			Code:    500,
			Message: err.Error(),
			Title:   http.StatusText(500),
		})
	}

	if r.Method == "PUT" {
		// Check
		if aid, err = strconv.ParseInt(params.ByName("aid"), 10, 64); err != nil || aid < 1 {
			if err == nil {
				err = errors.New("Invalid record id")
			}

			env.Error("%s: %s (id=%d)", id, err.Error(), aid)

			return NewResponse(&Error{
				Code:    500,
				Message: err.Error(),
				Title:   http.StatusText(500),
			})
		}

		form.Id = aid
	} else {
		form.Id = 0
	}

	env.Debug("%s: Alias data is valid", id)

	if err = env.SetAlias(&form); err != nil {
		env.Error("%s: %s", id, err.Error())

		return NewResponse(&Error{
			Code:    500,
			Message: "Cannot save alias data",
			Title:   http.StatusText(500),
		})
	}

	return NewResponse(nil)
}

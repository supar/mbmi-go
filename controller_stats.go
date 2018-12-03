package main

import (
	"mbmi-go/models"
	"net/http"
	"strconv"
)

func ServicesStat(r *http.Request, env Enviroment) ResponseIface {
	var (
		count uint64
		err   error
		resp  *Response
		m     []*models.Stat

		flt = models.NewFilter()
		id  = r.Context().Value("Id")
	)

	if uid, _ := strconv.ParseInt(r.FormValue("uid"), 10, 32); uid > 0 {
		flt.Where("uid", uid)
	}

	// Sort
	srt := r.FormValue("sort")
	if srt == "" {
		srt = "updated"
	}

	dir := false
	if r.FormValue("dir") == "asc" {
		dir = true
	}

	flt.Order(srt, dir)

	// Apply page limitation
	helperLimit(r, flt)

	if m, count, err = env.ServicesStat(flt, true); err != nil {
		env.Error("%s: %s", id, err.Error())

		return NewResponse(&Error{
			Code:    500,
			Message: "Cannot fetch services statistcs from database",
			Title:   http.StatusText(500),
		})
	}

	resp = NewResponse(m)
	resp.Count = count

	return resp
}

// StatImapLogin updates last imap login information
func StatImapLogin(r *http.Request, env Enviroment) ResponseIface {
	var (
		domain string
		email  models.Email
		err    error
		flt    models.FilterIface
		login  string
		params routerParams
		user   []*models.User

		form = models.Stat{}
		id   = r.Context().Value("Id")
	)

	params = r.Context().Value("Params").(routerParams)
	email = models.Email(params.ByName("uid"))

	if login, domain, err = email.Split(); err != nil {
		env.Error("%s: %s", id, err.Error())

		return NewResponse(err)
	}

	flt = models.NewFilter().
		Where("imap", 1).
		Where("login", login).
		Where("domain", domain)

	if user, _, err = env.Users(flt, false); err != nil || len(user) != 1 {
		env.Error("%s: %s", id, err.Error())

		return NewResponse(&Error{
			Code:    500,
			Message: "Cannot fetch user from database",
			Title:   http.StatusText(500),
		})
	}

	if err = parseFormTo(r, &form); err != nil {
		env.Error("%s, %#v, %s", id, r.PostForm, err.Error())

		return NewResponse(&Error{
			Code:    500,
			Message: "cannot parse form data",
			Title:   http.StatusText(500),
		})
	}

	form.UID = user[0].Id
	form.Service = "imap"

	if err = env.SetStatImapLogin(&form); err != nil {
		env.Error("%s: %s", id, err.Error())

		return NewResponse(&Error{
			Code:    500,
			Message: "Cannot fetch user from database",
			Title:   http.StatusText(500),
		})
	}

	return NewResponse(nil)
}

package main

import (
	"context"
	"mbmi-go/models"
	"net/http"
	"strconv"
)

type Controller func(*http.Request, Enviroment) ResponseIface

// Create http handler
func NewHandler(fn Controller, env Enviroment) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var (
			data     []byte
			response ResponseIface
		)

		if response = fn(r, env); response == nil {
			return
		}

		data, _ = response.Get()
		w.Header().Set("Content-Type", "application/json; charset=utf-8")

		if !response.Ok() {
			w.WriteHeader(response.Status())
		}

		w.Write(data)
	})
}

// Wrap Controller function to prevent unauthorized requests
func Protect(fn Controller) Controller {
	return func(r *http.Request, env Enviroment) ResponseIface {
		var (
			id = r.Context().Value("Id").(string)
			tk = r.Context().Value(tokenKey).(IdentityIface)
		)

		if tk.Subject() != "authentication" {
			env.Error("%s: Invalid token: subject(authentication)=%s", id, tk.Subject())

			return NewResponse(&Error{
				Code:    401,
				Message: http.StatusText(401),
				Title:   http.StatusText(401),
			})
		}

		if !tk.Valid() {
			env.Error("%s: Unauthorized, token is nil or not valid", id)

			return NewResponse(&Error{
				Code:    401,
				Message: http.StatusText(401),
				Title:   http.StatusText(401),
			})
		}

		env.Debug("%s: Token is valid", id)

		return fn(r, env)
	}
}

// Authorization
func Login(r *http.Request, env Enviroment) ResponseIface {
	var (
		err   error
		model []*models.User
		token *Token
		claim TokenClaims

		flt    = models.NewFilter()
		form   = models.User{}
		id     = r.Context().Value("Id")
		secret = r.Context().Value(secretKey).(string)
	)

	if err = parseFormTo(r, &form); err != nil {
		env.Error("%s, %s", id, err.Error())

		return NewResponse(&Error{
			Code:    500,
			Message: "cannot parse form data",
			Title:   http.StatusText(500),
		})
	}

	form.Login, form.DomainName, err = form.Email.Split()

	if err != nil {
		env.Error("%s, email %s", id, err.Error())

		return NewResponse(&Error{
			Code:    500,
			Message: "cannot parse form data",
			Title:   http.StatusText(500),
		})
	}

	env.Debug("%s, #%v", id, form)

	if form.Login == "" ||
		form.DomainName == "" ||
		form.Password == "" {

		return NewResponse(&Error{
			Code:    401,
			Message: http.StatusText(401),
			Title:   http.StatusText(401),
		})
	}

	flt.Where("login", form.Login).
		Where("domain", form.DomainName).
		Where("passwd", form.Password).
		Where("manager", 1)

	if model, _, err = env.Users(flt, false); err != nil || len(model) != 1 {
		if err != nil {
			env.Error("%s: %s", id, err.Error())
		} else {
			if len(model) != 1 {
				env.Error("%s: User=(%s) with password=(%s) not found", id, form.Email, form.Password)
			}
		}

		return NewResponse(&Error{
			Code:    401,
			Message: http.StatusText(401),
			Title:   http.StatusText(401),
		})
	}

	claim = NewClaims(model[0].Id, "authentication")
	claim.Issuer = string(form.Email)

	token = NewToken([]byte(secret)).Sign(claim)

	return NewResponse(token)
}

// Spamers statistics
func Spam(r *http.Request, env Enviroment) ResponseIface {
	var (
		count uint64
		err   error
		resp  *Response
		t     []*models.Spam

		interval uint64

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

	if _, ok := r.Form["interval"]; ok {
		interval, _ = strconv.ParseUint(r.Form.Get("interval"), 10, 64)
	}

	if interval < 1 {
		interval = 60
	}

	flt.Where("interval", interval)

	if srt := r.FormValue("sort"); srt != "" {
		dir := true
		if r.FormValue("dir") == "desc" {
			dir = false
		}

		flt.Order(srt, dir)
	}

	// Apply page limitation
	helperLimit(r, flt)

	if t, count, err = env.Spam(flt, true); err != nil {
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

// Get transport list
func Transports(r *http.Request, env Enviroment) ResponseIface {
	var (
		count uint64
		err   error
		resp  *Response
		m     []*models.Transport

		//tid int64 = -1
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

	if _, ok := r.Form["domain"]; ok {
		flt.Where("domain", r.Form.Get("domain")+"%")
	}

	// Apply page limitation
	helperLimit(r, flt)

	if m, count, err = env.Transports(flt, true); err != nil {
		env.Error("%s: %s", id, err.Error())

		return NewResponse(&Error{
			Code:    500,
			Message: "Cannot fetch transports from database",
			Title:   http.StatusText(500),
		})
	}

	resp = NewResponse(m)
	resp.Count = count

	return resp
}

// Get Single transport item
func Transport(r *http.Request, env Enviroment) ResponseIface {
	var (
		err    error
		params routerParams
		tid    int64
		t      []*models.Transport

		flt = models.NewFilter()
		id  = r.Context().Value("Id")
	)

	params = r.Context().Value("Params").(routerParams)

	if tid_str := params.ByName("tid"); tid_str != "" {
		if tid, err = strconv.ParseInt(tid_str, 10, 32); err != nil {
			env.Error("%s: %s", id, err.Error())
		}
	}

	if tid < 1 {
		return NewResponse(&Error{
			Code:    404,
			Message: "empty transport id",
			Title:   http.StatusText(404),
		})
	}

	flt.Where("id", tid)

	if t, _, err = env.Transports(flt, false); err != nil {
		env.Error("%s: %s", id, err.Error())

		return NewResponse(&Error{
			Code:    500,
			Message: "Cannot fetch user from database",
			Title:   http.StatusText(500),
		})
	}

	if l := len(t); l == 1 {
		return NewResponse(t[0])
	}

	env.Error("%s: Can't find transport with id=(%d)", id, tid)

	return NewResponse(&Error{
		Code:    404,
		Message: http.StatusText(404),
		Title:   http.StatusText(404),
	})
}

func MailSearch(r *http.Request, env Enviroment) ResponseIface {
	var (
		count uint64
		err   error
		resp  *Response
		m     []string

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

	if _, ok := r.Form["query"]; ok {
		flt.Where("mail", r.Form.Get("query")+"%")
	}

	if _, ok := r.Form["sort"]; ok {
		flt.Order(r.Form.Get("sort"), false)
	} else {
		flt.Order(r.Form.Get("sort"), false)
	}

	// Apply page limitation
	helperLimit(r, flt)

	if m, count, err = env.MailSearch(flt, true); err != nil {
		env.Error("%s: %s", id, err.Error())

		return NewResponse(&Error{
			Code:    500,
			Message: "Cannot fetch any mail from database",
			Title:   http.StatusText(500),
		})
	}

	resp = NewResponse(m)
	resp.Count = count

	return resp
}

// Get random passwords list
func Password(r *http.Request, env Enviroment) ResponseIface {
	var (
		count  uint64
		err    error
		resp   *Response
		pass_l int

		m  = make([]string, 3)
		id = r.Context().Value("Id")
	)

	if err = r.ParseForm(); err != nil {
		env.Error("%s, %s", id, err.Error())

		return NewResponse(&Error{
			Code:    500,
			Message: "cannot parse form data",
			Title:   http.StatusText(500),
		})
	}

	pass_l, _ = strconv.Atoi(r.Form.Get("length"))

	if pass_l <= 0 {
		pass_l = 16
	}

	for l, i := len(m), 0; i < l; i++ {
		var str string

		if str, err = createSecret(pass_l, false, false, false); err != nil {
			env.Error("%s: %s", id, err.Error())

			return NewResponse(&Error{
				Code:    500,
				Message: "Cannot fetch transports from database",
				Title:   http.StatusText(500),
			})
		}

		m[i] = str
	}

	resp = NewResponse(m)
	resp.Count = count

	return resp
}

// Wrap aliases and add grouping property
func aliasGroupWrap(fn Controller) Controller {
	return func(r *http.Request, env Enviroment) ResponseIface {
		return fn(
			r.WithContext(context.WithValue(r.Context(), "Group", "alias")),
			env,
		)
	}
}

func secretWrap(fn Controller, secret string) Controller {
	return func(r *http.Request, env Enviroment) ResponseIface {
		return fn(
			r.WithContext(context.WithValue(r.Context(), secretKey, secret)),
			env,
		)
	}
}

// Helper to parse query and paste page limitation to the filler object
func helperLimit(r *http.Request, flt models.FilterIface) {
	var (
		limit, offset uint64
	)

	limit = 10
	offset = 0

	if _, ok := r.Form["limit"]; ok {
		limit, _ = strconv.ParseUint(r.Form.Get("limit"), 10, 64)
	}

	if _, ok := r.Form["offset"]; ok {
		offset, _ = strconv.ParseUint(r.Form.Get("offset"), 10, 64)
	}

	flt.Limit(limit, offset)
}

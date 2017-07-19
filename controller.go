package main

import (
	"context"
	"github.com/supar/dsncfg"
	"mbmi-go/models"
	"net/http"
	"strconv"
	"strings"
)

type Controller struct {
	LogIface
	models models.Datastore
}

func NewController(log LogIface) (c *Controller, err error) {
	var (
		dsn = &dsncfg.Database{
			Host:     DBADDRESS,
			Name:     DBNAME,
			User:     DBUSER,
			Password: DBPASS,
			Type:     "mysql",
			Parameters: map[string]string{
				"charset":   "utf8",
				"parseTime": "True",
				"loc":       "Local",
			},
		}
		m models.Datastore
	)

	if m, err = models.Init(dsn); err != nil {
		return
	}

	c = &Controller{
		LogIface: log,
		models:   m,
	}

	return
}

func (s *Controller) Handle(fn func(*http.Request, context.Context) ResponseIface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var (
			err  error
			data []byte
			res  ResponseIface

			ctx = r.Context()
		)

		ctx = context.WithValue(ctx, "Query", r.URL.Query())

		res = fn(r, ctx)
		data, err = res.Get()

		if err != nil {
			s.Error("%s: %s", ctx.Value("Id"), err.Error())
		}

		if !res.Ok() {
			http.Error(w, "", res.Status())
		}

		w.Write(data)
	}
}

// Authorization
func (s *Controller) Login(req *http.Request, ctx context.Context) (resp ResponseIface) {
	var (
		claims        TokenClaims
		err           error
		login, domain string
		m             []*models.User
		token         *Token

		flt    = models.NewFilter()
		email  = req.FormValue("email")
		id     = ctx.Value("Id")
		passwd = req.FormValue("password")
	)

	if em := strings.Split(email, "@"); len(em) != 2 {
		s.Error("%s: Can't parse email=(%s)", id, email)

		resp = NewResponseError(401, nil)
		return
	} else {
		login = em[0]
		domain = em[1]
	}

	if login == "" ||
		domain == "" ||
		passwd == "" {

		resp = NewResponseError(401, nil)
		return
	}

	flt.Where("login", login).
		Where("domain", domain).
		Where("passwd", passwd)

	if m, _, err = s.models.Users(flt, false); err != nil || len(m) != 1 {
		if err != nil {
			s.Error("%s: %s", id, err.Error())
		} else {
			if len(m) != 1 {
				s.Error("%s: User=(%s) with password=(%s) not found", id, email, passwd)
			}
		}

		resp = NewResponseError(401, nil)
		return
	}

	claims = NewClaims()
	claims.Uid = m[0].Id
	claims.Subject = email

	token = NewToken([]byte("secret")).
		Sign(claims)

	resp = NewResponseOk(token)
	return
}

// Get user by id
func (s *Controller) User(r *http.Request, ctx context.Context) ResponseIface {
	var (
		err    error
		params routerParams
		uid    int64
		u      []*models.User

		flt = models.NewFilter()
		id  = ctx.Value("Id")
	)

	params = ctx.Value("Params").(routerParams)

	if uid_str := params.ByName("uid"); uid_str == "me" {
		uid = ctx.Value("Token").(*Token).Identity()
	} else {
		uid, err = strconv.ParseInt(uid_str, 10, 32)
	}

	if err != nil || uid < 1 {
		if err != nil {
			s.Error("%s: %s", id, err.Error())
		}

		return NewResponseError(404, nil)
	}

	flt.Where("id", uid)

	if u, _, err = s.models.Users(flt, false); err != nil {
		s.Error("%s: %s", id, err.Error())

		return NewResponseError(404, nil)
	}

	if l := len(u); l == 1 {
		return NewResponseOk(u[0])
	} else {
		s.Error("%s: Can't find user with uid=(%d), rows count=(%d)", id, uid, l)
	}

	return NewResponseError(404, nil)
}

func (s *Controller) Users(r *http.Request, ctx context.Context) ResponseIface {
	var (
		count uint64
		err   error
		resp  *Response
		u     []*models.User

		limit, offset uint64

		flt = models.NewFilter()
		id  = ctx.Value("Id")
	)

	if err = r.ParseForm(); err != nil {
		s.Error("%s, %s", id, err.Error())

		return NewResponseError(404, nil)
	}

	if _, ok := r.Form["email"]; ok {
		flt.Where("emlike", r.Form.Get("email")+"%")
	}

	limit = 0
	offset = 100

	if _, ok := r.Form["limit"]; ok {
		limit, _ = strconv.ParseUint(r.Form.Get("limit"), 10, 64)
	}

	if _, ok := r.Form["offset"]; ok {
		offset, _ = strconv.ParseUint(r.Form.Get("offset"), 10, 64)
	}

	flt.Limit(limit, offset)

	if u, count, err = s.models.Users(flt, true); err != nil {
		s.Error("%s: %s", id, err.Error())

		return NewResponseError(404, nil)
	}

	resp = NewResponseOk(u)
	resp.Count = count

	return resp
}

func (s *Controller) Spam(r *http.Request, ctx context.Context) ResponseIface {
	var (
		count uint64
		err   error
		resp  *Response
		t     []*models.Spam

		interval,
		limit,
		offset uint64

		flt = models.NewFilter()
		id  = ctx.Value("Id")
	)

	if err = r.ParseForm(); err != nil {
		s.Error("%s, %s", id, err.Error())

		return NewResponseError(404, nil)
	}

	limit = 0
	offset = 100

	if _, ok := r.Form["interval"]; ok {
		interval, _ = strconv.ParseUint(r.Form.Get("interval"), 10, 64)
	}

	if _, ok := r.Form["limit"]; ok {
		limit, _ = strconv.ParseUint(r.Form.Get("limit"), 10, 64)
	}

	if _, ok := r.Form["offset"]; ok {
		offset, _ = strconv.ParseUint(r.Form.Get("offset"), 10, 64)
	}

	if _, ok := r.Form["sort"]; ok {
		flt.Order(r.Form.Get("sort"), true)
	}

	if interval < 1 {
		interval = 60
	}

	flt.Where("interval", interval).
		Limit(limit, offset)

	if t, count, err = s.models.Spam(flt, true); err != nil {
		s.Error("%s: %s", id, err.Error())

		if err == models.ErrFilterArgument {
			return NewResponseError(500, err)
		}

		return NewResponseError(500, "Server internal error")
	}

	resp = NewResponseOk(t)
	resp.Count = count

	return resp
}

func (s *Controller) Transports(r *http.Request, ctx context.Context) ResponseIface {
	var (
		count  uint64
		doCnt  bool
		err    error
		params routerParams
		resp   *Response
		m      []*models.Transport

		limit, offset uint64

		tid int64 = -1
		flt       = models.NewFilter()
		id        = ctx.Value("Id")
	)

	params = ctx.Value("Params").(routerParams)

	if tid_str := params.ByName("tid"); tid_str != "" {
		tid, err = strconv.ParseInt(tid_str, 10, 32)

		if err != nil {
			s.Error("%s: %s", id, err.Error())
		}

		flt.Where("id", tid)
	}

	if tid < 0 {
		if err = r.ParseForm(); err != nil {
			s.Error("%s, %s", id, err.Error())

			return NewResponseError(404, nil)
		}

		if _, ok := r.Form["domain"]; ok {
			flt.Where("domain", r.Form.Get("domain")+"%")
		}

		limit = 0
		offset = 100

		if _, ok := r.Form["limit"]; ok {
			limit, _ = strconv.ParseUint(r.Form.Get("limit"), 10, 64)
		}

		if _, ok := r.Form["offset"]; ok {
			offset, _ = strconv.ParseUint(r.Form.Get("offset"), 10, 64)
		}

		flt.Limit(limit, offset)
		doCnt = true
	}

	if m, count, err = s.models.Transports(flt, doCnt); err != nil {
		s.Error("%s: %s", id, err.Error())

		return NewResponseError(404, nil)
	}

	if tid > -1 {
		if l := len(m); l == 1 {
			return NewResponseOk(m[0])
		} else {
			s.Error("%s: Can't find user with uid=(%d), rows count=(%d)", id, tid, l)
		}

		return NewResponseError(404, nil)
	}

	resp = NewResponseOk(m)
	resp.Count = count

	return resp
}

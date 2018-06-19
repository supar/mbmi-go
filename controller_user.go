package main

import (
	"errors"
	"mbmi-go/models"
	"net/http"
	"strconv"
)

// Get users/mailboxes list
func Users(r *http.Request, env Enviroment) ResponseIface {
	var (
		count uint64
		err   error
		resp  *Response
		u     []*models.User

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

	if _, ok := r.Form["email"]; ok {
		if v := r.Form.Get("email"); v != "" {
			flt.Where("emlike", v+"%")
		}
	}

	if _, ok := r.Form["query"]; ok {
		if v := r.Form.Get("query"); v != "" {
			flt.Where("search", "%"+v+"%")
		}
	}

	if _, ok := r.Form["mode"]; ok {
		switch r.Form.Get("mode") {
		case "on":
			flt.Where("mode_on", nil)

		case "off":
			flt.Where("mode_off", nil)
		}
	}

	// Apply page limitation
	helperLimit(r, flt)

	if u, count, err = env.Users(flt, true); err != nil {
		env.Error("%s: %s", id, err.Error())

		return NewResponse(&Error{
			Code:    500,
			Message: "Cannot fetch users from database",
			Title:   http.StatusText(500),
		})
	}

	resp = NewResponse(u)
	resp.Count = count

	return resp
}

// Get user by id
func User(r *http.Request, env Enviroment) ResponseIface {
	var (
		err    error
		params routerParams
		uid    int64
		u      []*models.User

		flt = models.NewFilter()
		id  = r.Context().Value("Id")
	)

	params = r.Context().Value("Params").(routerParams)

	if uid_str := params.ByName("uid"); uid_str == "me" {
		uid = r.Context().Value("Token").(*Token).Identity()
	} else {
		uid, err = strconv.ParseInt(uid_str, 10, 32)
	}

	if err != nil || uid < 1 {
		if err != nil {
			env.Error("%s: %s", id, err.Error())
		}

		return NewResponse(&Error{
			Code:    404,
			Message: "empty user id",
			Title:   http.StatusText(404),
		})
	}

	flt.Where("id", uid)

	if u, _, err = env.Users(flt, false); err != nil {
		env.Error("%s: %s", id, err.Error())

		return NewResponse(&Error{
			Code:    500,
			Message: "Cannot fetch user from database",
			Title:   http.StatusText(500),
		})
	}

	if l := len(u); l == 1 {
		return NewResponse(u[0])
	}

	env.Error("%s: Can't find user with id=(%d)", id, uid)

	return NewResponse(&Error{
		Code:    404,
		Message: http.StatusText(404),
		Title:   http.StatusText(404),
	})
}

func SetUser(r *http.Request, env Enviroment) ResponseIface {
	var (
		uid       int64
		err       error
		transport []*models.Transport
		user      []*models.User

		form   = models.User{}
		flt    = models.NewFilter()
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

	if r.Method == "PUT" {
		// Check
		if uid, err = strconv.ParseInt(params.ByName("uid"), 10, 64); err != nil || uid < 1 {
			if err == nil {
				err = errors.New("Invalid record id")
			}

			env.Error("%s: %s (id=%d)", id, err.Error(), uid)

			return NewResponse(&Error{
				Code:    500,
				Message: err.Error(),
				Title:   http.StatusText(500),
			})
		}

		form.Id = uid
	} else {
		form.Id = 0
	}

	// Validate domain
	// Identify domain by id from request
	flt.Where("id", form.Domain)
	if transport, _, err = env.Transports(flt, false); err != nil {
		env.Error("%s: %s", id, err.Error())

		return NewResponse(&Error{
			Code:    500,
			Message: "Cannot fetch transports from database",
			Title:   http.StatusText(500),
		})
	}

	if l := len(transport); l != 1 {
		if l < 1 {
			err = errors.New("Unknown transport, empty result")
		} else {
			err = errors.New("Unknown transport, found more than one record")
		}

		env.Error("%s: %s", id, err.Error())

		return NewResponse(&Error{
			Code:    500,
			Message: "Cannot fetch transports from database",
			Title:   http.StatusText(500),
		})
	}

	// Put domain name to the form object
	form.DomainName = transport[0].Domain

	// Concat login with domain before validation
	form.Email = models.Email(form.Login + "@" + form.DomainName)
	// Validate full email
	if _, _, err = form.Email.Split(); err != nil {
		env.Error("%s: %s", id, err.Error())

		return NewResponse(&Error{
			Code:    500,
			Message: err.Error(),
			Title:   http.StatusText(500),
		})
	}

	if form.Id > 0 {
		flt = models.NewFilter().Where("id", form.Id)
		if user, _, err = env.Users(flt, false); err != nil {
			env.Error("%s: %s", id, err.Error())

			return NewResponse(&Error{
				Code:    500,
				Message: err.Error(),
				Title:   http.StatusText(500),
			})
		}

		if l := len(user); l != 1 {
			if l < 1 {
				err = errors.New("Unknown user, empty result")
			} else {
				err = errors.New("Unknown user, found more than one record")
			}

			env.Error("%s: %s", id, err.Error())

			return NewResponse(&Error{
				Code:    500,
				Message: err.Error(),
				Title:   http.StatusText(500),
			})
		}
	} else {
		if form.Password == "" {
			err = errors.New("Password required")

			return NewResponse(&Error{
				Code:    500,
				Message: err.Error(),
				Title:   http.StatusText(500),
			})
		}
	}

	if form.Gid < 1 {
		form.Gid = transport[0].Gid
	}

	if form.Uid < 1 {
		form.Uid = transport[0].Uid
	}

	env.Debug("%s: User data is valid", id)

	if err = env.SetUser(&form); err != nil {
		env.Error("%s: %s", id, err.Error())

		return NewResponse(&Error{
			Code:    500,
			Message: "Cannot save user data",
			Title:   http.StatusText(500),
		})
	}

	return NewResponse(nil)
}

// GetUserJWT returns user JWT and updates secret
func GetUserJWT(r *http.Request, env Enviroment) ResponseIface {
	var (
		err    error
		params routerParams
		uid    int64
		model  []*models.User
		secret string
		token  *Token
		claim  TokenClaims

		flt = models.NewFilter()
		id  = r.Context().Value("Id")
	)

	params = r.Context().Value("Params").(routerParams)

	if uidStr := params.ByName("uid"); uidStr == "me" {
		uid = r.Context().Value("Token").(*Token).Identity()
	} else {
		uid, err = strconv.ParseInt(uidStr, 10, 32)
	}

	if err != nil || uid < 1 {
		if err != nil {
			env.Error("%s: %s", id, err.Error())
		}

		return NewResponse(&Error{
			Code:    404,
			Message: "empty user id",
			Title:   http.StatusText(404),
		})
	}

	flt.Where("id", uid)

	if model, _, err = env.Users(flt, false); err != nil {
		env.Error("%s: %s", id, err.Error())

		return NewResponse(&Error{
			Code:    500,
			Message: "Cannot fetch user from database",
			Title:   http.StatusText(500),
		})
	}

	if len(model) != 1 {
		env.Error("%s: Can't find user with id=(%d)", id, uid)

		return NewResponse(&Error{
			Code:    404,
			Message: http.StatusText(404),
			Title:   http.StatusText(404),
		})
	}

	if len(model[0].Token()) < 32 {
		env.Error("%s: It seems application empty token=(%s)", id, model[0].Token())

		return NewResponse(&Error{
			Code:    500,
			Message: "Application token required",
			Title:   http.StatusText(500),
		})
	}

	if secret, err = createSecret(64, false, false, false); err != nil {
		env.Error("%s: %s", id, err.Error())

		return NewResponse(&Error{
			Code:    500,
			Message: "Internal Server Error",
			Title:   http.StatusText(500),
		})
	}

	model[0].SetSecret(secret)

	env.Debug("%s: Update user secret", id)

	if err = env.SetUserSecret(model[0]); err != nil {
		env.Error("%s: %s", id, err.Error())

		return NewResponse(&Error{
			Code:    500,
			Message: "Cannot save user data",
			Title:   http.StatusText(500),
		})
	}

	claim = NewClaims(model[0].Id, "authentication")
	claim.Issuer = model[0].Login + "@" + model[0].DomainName

	env.Debug("%s: claim: %#v", id, claim)

	token = NewToken([]byte(secret)).Sign(claim)

	return NewResponse(token)
}

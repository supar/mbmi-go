package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"github.com/julienschmidt/httprouter"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
)

type Values struct {
	url.Values
}

type Tmock struct {
	code   int
	method string
	values url.Values
}

func (v Values) New() Values {
	return Values{
		Values: url.Values{},
	}
}

func (v Values) Set(key, value string) Values {
	v.Values[key] = []string{value}
	return v
}

func initDBMock(t *testing.T) (db *sql.DB, mock sqlmock.Sqlmock) {
	var (
		err error
	)

	// open database stub
	db, mock, err = sqlmock.New()
	if err != nil {
		t.Fatalf("An error '%s' was not expected when opening a stub database connection", err)
	}

	return
}

func initTestBus(t *testing.T, skipErr bool) *Bus {
	return &Bus{LogIface: &TestingLogWrap{
		T:          t,
		skipErrors: skipErr,
	}}
}

// create request with ready context
func request(method, url string, body io.Reader) (r *http.Request, err error) {
	var (
		ctx context.Context
	)

	if r, err = http.NewRequest(method, url, body); err != nil {
		return
	}

	if method == "POST" || method == "PUT" {
		r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	}

	ctx = context.WithValue(r.Context(), "Id", "test_"+method)
	r = r.WithContext(ctx)

	return
}

func Test_GetAliasesList(t *testing.T) {
	db, mock := initDBMock(t)
	req, _ := request("GET", "/aliases?limit=50&offset=0&alias=alerts%40doamin.com", nil)
	env := initTestBus(t, false)

	rows := sqlmock.NewRows([]string{"id", "alias", "recipient", "comment"}).
		AddRow(1, "alert@doamin.com", "some@domain.com", "Any text")

	count := sqlmock.NewRows([]string{"count"}).AddRow(1)

	mock.ExpectQuery("SELECT").WillReturnRows(rows)
	mock.ExpectQuery("SELECT COUNT").WillReturnRows(count)

	if err := env.openDB(db); err != nil {
		t.Error(err)
	}

	Aliases(req, env)
}

func Test_GetAliasGroupsList(t *testing.T) {
	db, mock := initDBMock(t)
	req, _ := request("GET", "/aliases?limit=50&offset=0&alias=alert%40domain.com&recipient=", nil)
	env := initTestBus(t, false)

	rows := sqlmock.NewRows([]string{"id", "alias", "recipient", "comment"}).
		AddRow(1, "alert@doamin.com", "some@domain.com", "Any text")

	count := sqlmock.NewRows([]string{"count"}).AddRow(1)

	mock.ExpectQuery("SELECT.*GROUP BY").WillReturnRows(rows)
	mock.ExpectQuery("SELECT COUNT.*GROUP BY").WillReturnRows(count)

	if err := env.openDB(db); err != nil {
		t.Error(err)
	}

	fn := aliasGroupWrap(Aliases)
	fn(req, env)
}

func Test_GetUsersList(t *testing.T) {
	db, mock := initDBMock(t)
	req, _ := request("GET", "/users", nil)
	env := initTestBus(t, false)

	rows := sqlmock.NewRows([]string{
		"id",
		"name",
		"login",
		"domid",
		"passwd",
		"uid",
		"gid",
		"smtp",
		"imap",
		"pop3",
		"sieve",
		"manager",
		"domainname",
		"secret",
		"token",
	}).
		AddRow(1, "Alert User Name", "alert", 1, "anypass", 8, 8, 1, 1, 0, 1, 1, "doamin.com", "", "")

	count := sqlmock.NewRows([]string{"count"}).AddRow(1)

	mock.ExpectQuery("SELECT").WillReturnRows(rows)
	mock.ExpectQuery("SELECT COUNT").WillReturnRows(count)

	if err := env.openDB(db); err != nil {
		t.Error(err)
	}

	resp := Users(req, env)

	if !resp.Ok() {
		t.Errorf("Required success response, but got %d", resp.Status())
	}
}

func Test_SearchUserByEmailOrName(t *testing.T) {
	db, mock := initDBMock(t)
	req, _ := request("GET", "/users?query=alert%40domain.com", nil)
	env := initTestBus(t, false)

	rows := sqlmock.NewRows([]string{
		"id",
		"name",
		"login",
		"domid",
		"passwd",
		"uid",
		"gid",
		"smtp",
		"imap",
		"pop3",
		"sieve",
		"manager",
		"domainname",
		"secret",
		"token",
	}).
		AddRow(1, "Alert User Name", "alert", 1, "anypass", 8, 8, 1, 1, 0, 1, 1, "doamin.com", "", "")

	count := sqlmock.NewRows([]string{"count"}).AddRow(1)

	mock.ExpectQuery("SELECT.+WHERE.+LIKE.+OR.+CONCAT").
		WithArgs("%alert@domain.com%", "%alert@domain.com%", 0, 10).
		WillReturnRows(rows)

	mock.ExpectQuery("SELECT COUNT").WillReturnRows(count)

	if err := env.openDB(db); err != nil {
		t.Error(err)
	}

	resp := Users(req, env)

	if !resp.Ok() {
		t.Errorf("Required success response, but got %d", resp.Status())
	}
}

func Test_GetUnknownUser_404ShoulBe(t *testing.T) {
	db, mock := initDBMock(t)
	req, _ := request("GET", "/user/4", nil)
	req = req.WithContext(context.WithValue(req.Context(), "Params", httprouter.Params{httprouter.Param{"uid", "4"}}))
	env := initTestBus(t, true)

	rows := sqlmock.NewRows([]string{
		"id", "name", "login", "domid",
		"passwd", "uid", "gid", "smtp", "imap", "pop3",
		"sieve", "manager", "domainname",
	})

	count := sqlmock.NewRows([]string{"count"}).AddRow(1)

	mock.ExpectQuery("SELECT").WithArgs(4).WillReturnRows(rows)
	mock.ExpectQuery("SELECT COUNT").WillReturnRows(count)

	if err := env.openDB(db); err != nil {
		t.Error(err)
	}

	resp := User(req, env)

	if resp.Ok() {
		t.Errorf("Required success response, but got %d", resp.Status())
	}
}

func Test_GetMailSearchList(t *testing.T) {
	db, mock := initDBMock(t)
	req, _ := request("GET", "/aliases/search?query=a", nil)
	env := initTestBus(t, false)

	rows := sqlmock.NewRows([]string{"mail"}).
		AddRow("alert@doamin.com").
		AddRow("some@domain.com").
		AddRow("s_foo@domain.com").
		AddRow("stepby@domid.com").
		AddRow("foo@network.net").
		AddRow("abrabr@keep.com").
		AddRow("123dd@domain.com")

	mock.ExpectQuery("SELECT [\\`\\w]+ FROM[\\s(]+SELECT CONCAT.+UNION SELECT.+UNION SELECT").WillReturnRows(rows)

	if err := env.openDB(db); err != nil {
		t.Error(err)
	}

	resp := MailSearch(req, env)

	if !resp.Ok() {
		t.Errorf("Required success response, but got %d", resp.Status())
	}
}

func Test_SaveAlias_Suites(t *testing.T) {
	db, mock := initDBMock(t)
	env := initTestBus(t, true)

	aliases := []Tmock{
		// Update entry
		// Empty data
		Tmock{
			code:   500,
			method: "PUT",
			values: url.Values(map[string][]string{
				"id": []string{"0"},
			}),
		},
		// Invalid Alias data
		Tmock{
			code:   500,
			method: "PUT",
			values: url.Values(map[string][]string{
				"alias": []string{"s f,@asdf,,.com @"},
				"id":    []string{"0"},
			}),
		},
		// Invalid recipient data
		Tmock{
			code:   500,
			method: "PUT",
			values: url.Values(map[string][]string{
				"alias":     []string{"any@domain.com"},
				"id":        []string{"0"},
				"recipient": []string{"s f,@asdf,,.com"},
			}),
		},
		// Invalid record id
		Tmock{
			code:   500,
			method: "PUT",
			values: url.Values(map[string][]string{
				"id":        []string{"0"},
				"alias":     []string{"any@domain.com"},
				"recipient": []string{"tosomewahere@domain.com"},
				"comment":   []string{""},
			}),
		},
		Tmock{
			code:   200,
			method: "PUT",
			values: url.Values(map[string][]string{
				"id":        []string{"1"},
				"alias":     []string{"any@domain.com"},
				"recipient": []string{"tosomewahere@domain.com"},
				"comment":   []string{""},
			}),
		},
		// New entry
		// Empty data
		Tmock{
			code:   500,
			method: "POST",
			values: url.Values(map[string][]string{}),
		},
		// Invalid Alias data
		Tmock{
			code:   500,
			method: "POST",
			values: url.Values(map[string][]string{
				"alias": []string{"s f,@asdf,,.com @"},
			}),
		},
		// Invalid recipient data
		Tmock{
			code:   500,
			method: "POST",
			values: url.Values(map[string][]string{
				"alias":     []string{"any@domain.com"},
				"recipient": []string{"s f,@asdf,,.com"},
			}),
		},
		Tmock{
			code:   200,
			method: "POST",
			values: url.Values(map[string][]string{
				"alias":     []string{"any@domain.com"},
				"recipient": []string{"tosomewahere@domain.com"},
				"comment":   []string{""},
			}),
		},
	}

	if err := env.openDB(db); err != nil {
		t.Error(err)
	}

	for _, data := range aliases {
		var req *http.Request

		if data.method != "PUT" && data.method != "POST" {
			t.Fatalf("Unexpected method in the test: %s", data.method)
		}

		router := NewRouter()
		w := httptest.NewRecorder()
		id, _ := strconv.ParseInt(data.values.Get("id"), 10, 64)

		if data.method == "POST" {
			router.Handle(data.method, "/alias", NewHandler(SetAlias, env))
			req, _ = request(data.method, "/alias", strings.NewReader(data.values.Encode()))
		} else {
			router.Handle(data.method, "/alias/:aid", NewHandler(SetAlias, env))
			req, _ = request(data.method, "/alias/"+data.values.Get("id"), strings.NewReader(data.values.Encode()))
		}

		if data.code == 200 {
			if data.method == "POST" {
				mock.ExpectExec("^INSERT INTO.+VALUES").WithArgs(
					data.values.Get("alias"),
					data.values.Get("recipient"),
					data.values.Get("comment"),
				).WillReturnResult(sqlmock.NewResult(0, 0))
			} else {
				mock.ExpectExec("^UPDATE.+SET.+WHERE").WithArgs(
					data.values.Get("alias"),
					data.values.Get("recipient"),
					data.values.Get("comment"),
					id,
				).WillReturnResult(sqlmock.NewResult(0, 0))
			}
		}

		router.ServeHTTP(w, req)

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf(err.Error())
		}

		if w.Code != data.code {
			t.Errorf("Unexpected code was returned code=%d, body=%s", w.Code, w.Body)
		}
	}
}

func Test_SaveUser_Suites(t *testing.T) {
	db, mock := initDBMock(t)
	env := initTestBus(t, true)

	users := []Tmock{
		Tmock{
			code:   200,
			method: "PUT",
			values: url.Values(map[string][]string{
				"id":       []string{"1"},
				"name":     []string{"Any User"},
				"login":    []string{"domain.com"},
				"password": []string{"123"},
				"domain":   []string{"1"},
				"uid":      []string{"8"},
				"gid":      []string{"8"},
				"smtp":     []string{"1"},
				"imap":     []string{"1"},
			}),
		},
		Tmock{
			code:   200,
			method: "POST",
			values: url.Values(map[string][]string{
				"id":       []string{"1"},
				"name":     []string{"Any User"},
				"login":    []string{"domain.com"},
				"password": []string{"123"},
				"domain":   []string{"1"},
				"uid":      []string{"8"},
				"gid":      []string{"8"},
				"smtp":     []string{"1"},
				"imap":     []string{"1"},
			}),
		},
	}

	if err := env.openDB(db); err != nil {
		t.Error(err)
	}

	for _, data := range users {
		var req *http.Request

		if data.method != "PUT" && data.method != "POST" {
			t.Fatalf("Unexpected method in the test: %s", data.method)
		}

		router := NewRouter()
		w := httptest.NewRecorder()
		if data.method == "POST" {
			router.Handle(data.method, "/user", NewHandler(SetUser, env))
			req, _ = request(data.method, "/user", strings.NewReader(data.values.Encode()))
		} else {
			router.Handle(data.method, "/user/:uid", NewHandler(SetUser, env))
			req, _ = request(data.method, "/user/"+data.values.Get("id"), strings.NewReader(data.values.Encode()))
		}

		if data.code == 200 {
			rows := sqlmock.NewRows([]string{
				"id", "domain", "transport", "rootdir", "uid", "gid",
			}).
				AddRow(1, "doamin.com", "virtual", "/mail", 8, 8)

			mock.ExpectQuery("^SELECT").WillReturnRows(rows)

			if data.method == "PUT" {
				mock.ExpectQuery("^SELECT").WillReturnRows(
					sqlmock.NewRows([]string{
						"id",
						"name",
						"login",
						"domid",
						"passwd",
						"uid",
						"gid",
						"smtp",
						"imap",
						"pop3",
						"sieve",
						"manager",
						"domainname",
						"secret",
						"token",
					}).
						AddRow(
							data.values.Get("id"),
							data.values.Get("name"),
							data.values.Get("login"),
							data.values.Get("domain"),
							data.values.Get("password"),
							data.values.Get("uid"),
							data.values.Get("gid"),
							data.values.Get("smtp"),
							data.values.Get("imap"),
							data.values.Get("pop3"),
							data.values.Get("sieve"),
							data.values.Get("manager"),
							data.values.Get("domainname"),
							"",
							"",
						))

				mock.ExpectExec("^UPDATE.+users.+SET.+WHERE").WillReturnResult(sqlmock.NewResult(0, 1))
			} else {
				mock.ExpectExec("^INSERT\\s+INTO.+users.+VALUES").WillReturnResult(sqlmock.NewResult(1, 0))
			}

			if data.values.Get("password") != "" {
				mock.ExpectExec("^UPDATE[\\s`]+users[\\s`]+SET[\\s`]+passwd[\\s`=\\?]+WHERE").WillReturnResult(sqlmock.NewResult(0, 1))
			}
		}

		router.ServeHTTP(w, req)

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf(err.Error())
		}

		if w.Code != data.code {
			t.Errorf("Unexpected code was returned code=%d, body=%s", w.Code, w.Body)
		}
	}
}

func Test_SetStatImapLogin(t *testing.T) {
	db, mock := initDBMock(t)
	env := initTestBus(t, true)

	users := []Tmock{
		Tmock{
			code:   200,
			method: "POST",
			values: url.Values(map[string][]string{
				"email": []string{"alert@somedomain.net"},
				"ip":    []string{"127.0.0.1"},
			}),
		},
		Tmock{
			code:   200,
			method: "POST",
			values: url.Values(map[string][]string{
				"email": []string{"alert@somedomain.net"},
			}),
		},
	}

	if err := env.openDB(db); err != nil {
		t.Error(err)
	}

	for _, data := range users {
		if data.method != "POST" {
			t.Fatalf("Unexpected method in the test: %s", data.method)
		}

		router := NewRouter()
		w := httptest.NewRecorder()
		router.Handle(data.method, "/stat/imap/:uid", NewHandler(StatImapLogin, env))
		req, err := request(data.method, "/stat/imap/"+data.values.Get("email"), strings.NewReader(data.values.Encode()))

		if err != nil {
			t.Errorf(err.Error())
		}

		if data.code == 200 {
			mock.ExpectQuery("^SELECT").WillReturnRows(
				sqlmock.NewRows([]string{
					"id",
					"name",
					"login",
					"domid",
					"passwd",
					"uid",
					"gid",
					"smtp",
					"imap",
					"pop3",
					"sieve",
					"manager",
					"domainname",
					"secert",
					"token",
				}).
					AddRow(
						"1",
						"Any user",
						"alert",
						"1",
						"123",
						"8",
						"8",
						"0",
						"0",
						"0",
						"0",
						"1",
						"somedomain.net",
						"",
						"",
					))

			mock.ExpectExec("^INSERT\\sINTO.+statistics").WillReturnResult(sqlmock.NewResult(1, 0))
		}

		router.ServeHTTP(w, req)

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf(err.Error())
		}

		if w.Code != data.code {
			t.Errorf("Unexpected code was returned code=%d, body=%s", w.Code, w.Body)
		}
	}
}

func Test_GetPassword(t *testing.T) {
	env := initTestBus(t, true)

	data := []Tmock{
		Tmock{
			code:   200,
			method: "GET",
			values: url.Values(map[string][]string{}),
		},
		Tmock{
			code:   200,
			method: "GET",
			values: url.Values(map[string][]string{
				"length": []string{"22"},
			}),
		},
	}

	for _, data := range data {
		var req *http.Request

		router := NewRouter()
		w := httptest.NewRecorder()

		router.Handle(data.method, "/password", NewHandler(Password, env))
		req, _ = request(data.method, "/password?"+data.values.Encode(), nil)

		router.ServeHTTP(w, req)

		if w.Code != data.code {
			t.Errorf("Unexpected code was returned code=%d, body=%s", w.Code, w.Body)
		}
	}
}

func Test_ChangeAndGetUserJWT(t *testing.T) {
	db, mock := initDBMock(t)
	env := initTestBus(t, true)

	users := []Tmock{
		Tmock{
			code:   200,
			method: "PUT",
			values: url.Values(map[string][]string{
				"id":       []string{"1"},
				"name":     []string{"Any User"},
				"login":    []string{"domain.com"},
				"password": []string{"123"},
				"domain":   []string{"1"},
				"uid":      []string{"8"},
				"gid":      []string{"8"},
				"smtp":     []string{"1"},
				"imap":     []string{"1"},
				"token":    []string{"b81d3f1a9f0cfde152e770ccf64718b3"},
			}),
		},
	}

	if err := env.openDB(db); err != nil {
		t.Error(err)
	}

	for _, data := range users {
		if data.method != "PUT" {
			t.Fatalf("Unexpected method in the test: %s", data.method)
		}

		router := NewRouter()
		w := httptest.NewRecorder()
		router.Handle(data.method, "/application/jwt/:uid", NewHandler(GetUserJWT, env))
		req, err := request(data.method, "/application/jwt/"+data.values.Get("id"), nil)

		if err != nil {
			t.Errorf(err.Error())
		}

		if data.code == 200 {
			mock.ExpectQuery("^SELECT").WillReturnRows(
				sqlmock.NewRows([]string{
					"id",
					"name",
					"login",
					"domid",
					"passwd",
					"uid",
					"gid",
					"smtp",
					"imap",
					"pop3",
					"sieve",
					"manager",
					"domainname",
					"secert",
					"token",
				}).
					AddRow(
						data.values.Get("id"),
						data.values.Get("name"),
						data.values.Get("login"),
						data.values.Get("domain"),
						data.values.Get("password"),
						data.values.Get("uid"),
						data.values.Get("gid"),
						data.values.Get("smtp"),
						data.values.Get("imap"),
						data.values.Get("pop3"),
						data.values.Get("sieve"),
						data.values.Get("manager"),
						data.values.Get("domainname"),
						"",
						data.values.Get("token"),
					))

			mock.ExpectExec("^UPDATE.+users.+SET").WillReturnResult(sqlmock.NewResult(1, 0))
		}

		router.ServeHTTP(w, req)

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf(err.Error())
		}

		if w.Code != data.code {
			t.Errorf("Unexpected code was returned code=%d, body=%s", w.Code, w.Body)
		}
	}
}

func Test_Authentication(t *testing.T) {
	db, mock := initDBMock(t)
	env := initTestBus(t, true)

	users := []Tmock{
		Tmock{
			code:   200,
			method: "POST",
			values: url.Values(map[string][]string{
				"id":         []string{"1"},
				"name":       []string{"Any User"},
				"login":      []string{"some"},
				"password":   []string{"123"},
				"domain":     []string{"1"},
				"domainname": []string{"user.net"},
				"uid":        []string{"8"},
				"gid":        []string{"8"},
				"smtp":       []string{"1"},
				"imap":       []string{"1"},
				"token":      []string{"b81d3f1a9f0cfde152e770ccf64718b3"},
			}),
		},
	}

	if err := env.openDB(db); err != nil {
		t.Error(err)
	}

	router := NewRouter()
	router.Handle("POST", "/login", NewHandler(secretWrap(Login, "anysecret"), env))
	router.Handle("GET", "/user/:uid", NewHandler(Protect(User), env))

	mw := Middlewares(
		router,
		JWT("anysecret", env),
	)

	for _, data := range users {
		if data.code == 200 {
			mock.ExpectQuery("^SELECT.+users").WillReturnRows(
				sqlmock.NewRows([]string{
					"id",
					"name",
					"login",
					"domid",
					"passwd",
					"uid",
					"gid",
					"smtp",
					"imap",
					"pop3",
					"sieve",
					"manager",
					"domainname",
					"secert",
					"token",
				}).
					AddRow(
						data.values.Get("id"),
						data.values.Get("name"),
						data.values.Get("login"),
						data.values.Get("domain"),
						data.values.Get("password"),
						data.values.Get("uid"),
						data.values.Get("gid"),
						data.values.Get("smtp"),
						data.values.Get("imap"),
						data.values.Get("pop3"),
						data.values.Get("sieve"),
						data.values.Get("manager"),
						data.values.Get("domainname"),
						"",
						data.values.Get("token"),
					))
		}

		w := httptest.NewRecorder()
		req, _ := request("POST", "/login", strings.NewReader("email="+data.values.Get("login")+"@"+data.values.Get("domainname")+"&password="+data.values.Get("password")))
		mw.ServeHTTP(w, req)

		if w.Code != data.code {
			t.Errorf("Unexpected code was returned code=%d, body=%s", w.Code, w.Body)
		}

		resp := &Response{
			Data: &Token{},
		}

		if err := json.Unmarshal(w.Body.Bytes(), resp); err != nil {
			t.Error(err)
		}

		if data.code == 200 {
			mock.ExpectQuery("^SELECT.+users").WillReturnRows(
				sqlmock.NewRows([]string{
					"id",
					"name",
					"login",
					"domid",
					"passwd",
					"uid",
					"gid",
					"smtp",
					"imap",
					"pop3",
					"sieve",
					"manager",
					"domainname",
					"secert",
					"token",
				}).
					AddRow(
						data.values.Get("id"),
						data.values.Get("name"),
						data.values.Get("login"),
						data.values.Get("domain"),
						data.values.Get("password"),
						data.values.Get("uid"),
						data.values.Get("gid"),
						data.values.Get("smtp"),
						data.values.Get("imap"),
						data.values.Get("pop3"),
						data.values.Get("sieve"),
						data.values.Get("manager"),
						data.values.Get("domainname"),
						"",
						data.values.Get("token"),
					))
		}

		w = httptest.NewRecorder()
		req, _ = request("GET", "/user/me", nil)
		req.Header.Add("Authorization", "Bearer "+resp.Data.(*Token).JWT)
		mw.ServeHTTP(w, req)

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf(err.Error())
		}

		if w.Code != data.code {
			t.Errorf("Unexpected code was returned code=%d, body=%s", w.Code, w.Body)
		}
	}
}

func Test_GetAccessesList(t *testing.T) {
	db, mock := initDBMock(t)
	req, _ := request("GET", "/accesses", nil)
	env := initTestBus(t, false)

	rows := sqlmock.NewRows([]string{
		"client",
		"access",
	}).
		AddRow("1.1.1.1", "REJECT")

	count := sqlmock.NewRows([]string{"count"}).AddRow(1)

	mock.ExpectQuery("SELECT").WillReturnRows(rows)
	mock.ExpectQuery("SELECT COUNT").WillReturnRows(count)

	if err := env.openDB(db); err != nil {
		t.Error(err)
	}

	resp := Accesses(req, env)

	if !resp.Ok() {
		t.Errorf("Required success response, but got %d", resp.Status())
	}
}

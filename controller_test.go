package main

import (
	"context"
	"database/sql"
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
		"id", "name", "login", "domid",
		"passwd", "uid", "gid", "smtp", "imap", "pop3",
		"sieve", "manager", "domainname",
	}).
		AddRow(1, "Alert User Name", "alert", 1, "anypass", 8, 8, 1, 1, 0, 1, 1, "doamin.com")

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

package main

import (
	"context"
	"database/sql"
	"github.com/julienschmidt/httprouter"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
	"io"
	"net/http"
	"testing"
)

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

package models

import (
	"database/sql"
	"errors"
	_ "github.com/go-sql-driver/mysql"
)

var (
	ErrFilterArgument = errors.New("Unsupported filter argument")
	ErrFilterRequired = errors.New("Filter is required")
)

type Datastore interface {
	Aliases(FilterIface, bool) ([]*Alias, uint64, error)
	SetAlias(*Alias) error
	DelAlias(int64) error
	Users(FilterIface, bool) ([]*User, uint64, error)
	Spam(FilterIface, bool) ([]*Spam, uint64, error)
	Transports(FilterIface, bool) ([]*Transport, uint64, error)
	MailSearch(FilterIface, bool) ([]string, uint64, error)
	SetUser(*User) error
	SetUserSecret(*User) error
	SetStatImapLogin(*Stat) error
	Accesses(FilterIface, bool) ([]*Access, uint64, error)
}

type Debug func(v ...interface{})

type DB struct {
	*sql.DB
	Debug
}

func Init(driver *sql.DB, fn Debug) Datastore {
	d := &DB{DB: driver}

	if fn != nil {
		d.Debug = fn
	} else {
		d.Debug = func(v ...interface{}) {}
	}

	return d
}

// Wrap parent function to log query string and arguments
func (s *DB) Query(q string, args ...interface{}) (*sql.Rows, error) {
	// Write query
	s.Debug(q)
	// Write arguments
	s.Debug("%v", args)

	// Execute query
	return s.DB.Query(q, args...)
}

// Wrap parent function to log query string and arguments
func (s *DB) QueryRow(q string, args ...interface{}) *sql.Row {
	// Write query
	s.Debug(q)
	// Write arguments
	s.Debug("%v", args)

	// Execute query
	return s.DB.QueryRow(q, args...)
}

// Wrap parent function to log query string and arguments
func (s *DB) Exec(q string, args ...interface{}) (sql.Result, error) {
	// Write query
	s.Debug(q)
	// Write arguments
	s.Debug("%v", args)

	// Execute query
	return s.DB.Exec(q, args...)
}

package main

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/supar/dsncfg"
	"mbmi-go/models"
)

type Enviroment interface {
	LogIface
	models.Datastore
}

type Bus struct {
	LogIface
	models.Datastore
}

func (b *Bus) openDB(driver *sql.DB) (err error) {
	var (
		dsn *dsncfg.Database
	)

	if driver == nil {
		dsn = b.dsn()
		if err = dsn.Init(); err != nil {
			return
		}

		if driver, err = sql.Open(dsn.Type, dsn.DSN()); err != nil {
			return
		}
	}

	b.Datastore = models.Init(driver, b.Debug)
	return
}

func (b *Bus) dsn() *dsncfg.Database {
	return &dsncfg.Database{
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
}

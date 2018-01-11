package models

import (
	"database/sql"
	"errors"
	"strings"
)

type User struct {
	Id         int64  `json:"id"`
	Name       string `json:"name"`
	Login      string `json:"login"`
	Domain     uint   `json:"domain"`
	DomainName string `json:"domainname"`
	Password   string `json:"password" schema:"password"`
	Uid        uint   `json:"uid"`
	Gid        uint
	Smtp       bool
	Imap       bool
	Pop3       bool
	Sieve      bool
	Manager    bool   `json:"manager"`
	Email      string `json:"email" schema:"email"`
}

func (u *User) SplitEmail() (err error) {
	var parts []string

	parts = strings.Split(u.Email, "@")

	switch len(parts) {
	case 2:
		u.Login = parts[0]
		u.DomainName = parts[1]

	case 0:
		err = errors.New("split: empty result")

	default:
		err = errors.New("split: multiple result")
	}

	return
}

func (s *DB) Users(flt FilterIface, cnt bool) (m []*User, count uint64, err error) {
	var (
		query     *Query
		query_str string
		args      []interface{}
		rows      *sql.Rows
	)

	if flt == nil {
		flt = NewFilter()
	}

	query = flt.(*Query)

	for _, expr := range query.expressions {
		switch expr.name {
		case "WHERE":
			expr.CbFunc(userWhere)
		case "ORDER BY":
			expr.CbFunc(userOrder)
		}
	}

	// Base query
	query.raw = "SELECT `u`.`id` `id`" +
		", `u`.`name` `name`" +
		", `u`.`login` `login`" +
		", `u`.`domid` `domid`" +
		", `u`.`passwd` `passwd`" +
		", `u`.`uid` `uid`" +
		", `u`.`gid`  `gid`" +
		", `u`.`smtp` `smtp`" +
		", `u`.`imap` `imap`" +
		", `u`.`pop3` `pop3`" +
		", `u`.`sieve` `sieve`" +
		", `u`.`manager` `manager`" +
		", `t`.`domain` `domainname`" +
		" " +
		"FROM `users` AS `u` " +
		"LEFT JOIN `transport` `t` ON (`u`.`domid` = `t`.`id`) "

	// Add where
	if query_str, args, err = query.Compile(); err != nil {
		return
	}

	if rows, err = s.Query(query_str, args...); err != nil {
		return nil, 0, err
	}

	defer rows.Close()
	// Create empty slice
	m = make([]*User, 0)

	for rows.Next() {
		var i = &User{}

		err = rows.Scan(
			&i.Id,
			&i.Name,
			&i.Login,
			&i.Domain,
			&i.Password,
			&i.Uid,
			&i.Gid,
			&i.Smtp,
			&i.Imap,
			&i.Pop3,
			&i.Sieve,
			&i.Manager,
			&i.DomainName,
		)

		if err != nil {
			return nil, 0, err
		}

		m = append(m, i)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, err
	}

	if cnt {
		query.raw = "SELECT COUNT(*) " +
			"FROM `users` AS `u` " +
			"LEFT JOIN `transport` `t` ON (`u`.`domid` = `t`.`id`) "

		query.Un("LIMIT")

		if query_str, args, err = query.Compile(); err != nil {
			return
		}

		err = s.QueryRow(query_str, args...).Scan(&count)

		if err != nil && err == sql.ErrNoRows {
			err = nil
		}
	}

	return
}

func userWhere(arg sql.NamedArg) (string, error) {
	switch arg.Name {
	case "emlike":
		return "CONCAT(`u`.`login`, '@', `t`.`domain`) LIKE ?", nil

	case "id":
		return "`u`.`id` = ?", nil

	case "login":
		return "`u`.`login` = ?", nil

	case "domain":
		return "`t`.`domain` = ?", nil

	case "passwd":
		return "`u`.`passwd` = ?", nil

	case "manager":
		return "`u`.`manager` = ?", nil
	}

	return "", ErrFilterArgument
}

func userOrder(arg sql.NamedArg) (string, error) {
	switch arg.Name {
	case "id":
		return "`u`.`id`", nil
	}

	return "", ErrFilterArgument
}

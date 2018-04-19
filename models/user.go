package models

import (
	"database/sql"
)

type User struct {
	Id         int64   `json:"id" schema:"id"`
	Name       string  `json:"name" schema:"name"`
	Login      string  `json:"login" schema:"login"`
	Domain     uint    `json:"domain" schema:"domain"`
	DomainName string  `json:"domainname"`
	Password   string  `json:"password" schema:"password"`
	Uid        uint    `json:"uid" schema:"uid"`
	Gid        uint    `json:"gid" schema:"gid"`
	Smtp       Boolean `json:"smtp" schema:"smtp"`
	Imap       Boolean `json:"imap" schema:"imap"`
	Pop3       Boolean `json:"pop3" schema:"pop3"`
	Sieve      Boolean `json:"sieve" schema:"sieve"`
	Manager    Boolean `json:"manager" schema:"manager"`
	Email      Email   `json:"email" schema:"email"`
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

func (s *DB) SetUser(user *User) (err error) {
	var (
		result sql.Result
	)

	if user.Id > 0 {
		_, err = s.Exec("UPDATE `users` SET "+
			"`name` = ? "+
			", `login` = ?"+
			", `domid` = ?"+
			", `gid` = ?"+
			", `uid` = ?"+
			", `smtp` = ?"+
			", `imap` = ?"+
			", `pop3` = ?"+
			", `sieve` = ?"+
			", `manager` = ?"+
			" WHERE `id` = ?",
			user.Name,
			user.Login,
			user.Domain,
			user.Gid,
			user.Uid,
			user.Smtp,
			user.Imap,
			user.Pop3,
			user.Sieve,
			user.Manager,
			user.Id)
	} else {
		result, err = s.Exec("INSERT INTO `users` ("+
			"`name`"+
			", `login`"+
			", `domid`"+
			", `gid`"+
			", `uid`"+
			", `smtp`"+
			", `imap`"+
			", `pop3`"+
			", `sieve`"+
			", `manager`"+
			") VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
			user.Name,
			user.Login,
			user.Domain,
			user.Gid,
			user.Uid,
			user.Smtp,
			user.Imap,
			user.Pop3,
			user.Sieve,
			user.Manager)

		if err != nil {
			return
		}

		user.Id, err = result.LastInsertId()
	}

	if err == nil && user.Password != "" {
		_, err = s.Exec("UPDATE `users` SET `passwd` = ? WHERE `id` = ?", user.Password, user.Id)
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

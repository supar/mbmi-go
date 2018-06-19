package models

import (
	"database/sql"
	"strings"
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

	// protected
	secret string
	token  string
}

func (s *DB) Users(flt FilterIface, cnt bool) (m []*User, count uint64, err error) {
	var (
		query    *Query
		queryStr string
		args     []interface{}
		rows     *sql.Rows
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
		", `u`.`secret` `secret`" +
		", `u`.`token` `token`" +
		" " +
		"FROM `users` AS `u` " +
		"LEFT JOIN `transport` `t` ON (`u`.`domid` = `t`.`id`) "

	// Add where
	if queryStr, args, err = query.Compile(); err != nil {
		return
	}

	if rows, err = s.Query(queryStr, args...); err != nil {
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
			&i.secret,
			&i.token,
		)

		if err != nil {
			return nil, 0, err
		}

		i.Email = Email(strings.Join([]string{i.Login, i.DomainName}, "@"))

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

		if queryStr, args, err = query.Compile(); err != nil {
			return
		}

		err = s.QueryRow(queryStr, args...).Scan(&count)

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

// SetUserSecret updates user token secret
func (s *DB) SetUserSecret(user *User) (err error) {
	_, err = s.Exec("UPDATE `users` SET "+
		"`secret` = ? "+
		" WHERE `id` = ?",
		user.secret,
		user.Id)

	return
}

// Secret returns user secret saved to the struct before
func (u *User) Secret() string {
	return u.secret
}

// SetSecret updates secret in the struct
func (u *User) SetSecret(secret string) {
	u.secret = secret
}

// Token returns application token saved to the struct before
func (u *User) Token() string {
	return u.token
}

func userWhere(arg *NamedArg) (string, error) {
	switch arg.Name {
	case "emlike":
		return "CONCAT(`u`.`login`, '@', `t`.`domain`) LIKE ?", nil

	case "search":
		if arg.Value != nil && len(arg.Value) == 1 {
			arg.Fill(arg.Value[0], 2)
		}
		return "(`u`.`name` LIKE ? OR CONCAT(`u`.`login`, '@', `t`.`domain`) LIKE ?)", nil

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

	case "token":
		return "`u`.`token` = ?", nil

	case "imap":
		return "`u`.`imap` = ?", nil

	case "pop3":
		return "`u`.`pop3` = ?", nil

	case "mode_on":
		return "(`u`.`smtp` = 1 OR `u`.`imap` = 1 OR `u`.`pop3` = 1)", nil

	case "mode_off":
		return "(`u`.`smtp` = 0 AND `u`.`imap` = 0 AND `u`.`pop3` = 0)", nil
	}

	return "", ErrFilterArgument
}

func userOrder(arg *NamedArg) (string, error) {
	switch arg.Name {
	case "id":
		return "`u`.`id`", nil
	}

	return "", ErrFilterArgument
}

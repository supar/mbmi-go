package models

import (
	"database/sql"
)

type Transport struct {
	Id        int64  `json:"id"`
	Domain    string `json:"domain"`
	Uid       uint   `json:"uid" schema:"uid"`
	Gid       uint   `json:"gid" schema:"gid"`
	Transport string `json:"transport"`
	Root      string `json:"rootdir"`
}

func (s *DB) Transports(flt FilterIface, cnt bool) (m []*Transport, count uint64, err error) {
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
			expr.CbFunc(transportWhere)
		case "ORDER BY":
			expr.CbFunc(transportOrder)
		}
	}

	// Base query
	query.raw = "SELECT `t`.`id` `id`" +
		", `t`.`domain` `domain`" +
		", `t`.`transport` `transport`" +
		", `t`.`rootdir` `rootdir`" +
		", `t`.`uid` `uid`" +
		", `t`.`gid` `gid`" +
		" " +
		"FROM `transport` AS `t` "

	// Add where
	if query_str, args, err = query.Compile(); err != nil {
		return
	}

	if rows, err = s.Query(query_str, args...); err != nil {
		return nil, 0, err
	}

	defer rows.Close()
	// Create empty slice
	m = make([]*Transport, 0)

	for rows.Next() {
		var i = &Transport{}

		err = rows.Scan(
			&i.Id,
			&i.Domain,
			&i.Transport,
			&i.Root,
			&i.Uid,
			&i.Gid,
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
			"FROM `transport` AS `t` "

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

func transportWhere(arg *NamedArg) (string, error) {
	switch arg.Name {
	case "id":
		return "`t`.`id` = ?", nil

	case "domain":
		return "`t`.`domain` = ?", nil
	}

	return "", ErrFilterArgument
}

func transportOrder(arg *NamedArg) (string, error) {
	switch arg.Name {
	case "id":
		return "`t`.`id`", nil
	}

	return "", ErrFilterArgument
}

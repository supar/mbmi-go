package models

import (
	"database/sql"
)

func (s *DB) MailSearch(flt FilterIface, cnt bool) (m []string, count uint64, err error) {
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
			expr.CbFunc(mailSearchWhere)
		case "ORDER BY":
			expr.CbFunc(mailSearchOrder)
		}
	}

	// Base query
	query.raw = "SELECT `mail` FROM (" +
		"SELECT CONCAT(`u`.`login`, '@', `t`.`domain`) AS `mail` " +
		"FROM `users` AS `u` LEFT JOIN `transport` AS `t` ON (`u`.`domid` = `t`.`id`) " +
		"UNION " +
		"SELECT `alias` AS `mail` FROM `aliases` " +
		"UNION " +
		"SELECT `recipient` AS `mail` FROM aliases " +
		") as `maillist` "

	// Add where
	if query_str, args, err = query.Compile(); err != nil {
		return
	}

	if rows, err = s.Query(query_str, args...); err != nil {
		return nil, 0, err
	}

	defer rows.Close()
	// Create empty slice
	m = make([]string, 0)

	for rows.Next() {
		var i string

		err = rows.Scan(
			&i,
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
		count = uint64(len(m))
	}

	return
}

func mailSearchWhere(arg *NamedArg) (string, error) {
	switch arg.Name {
	case "mail":
		return "`mail` LIKE ?", nil

	}

	return "", ErrFilterArgument
}

func mailSearchOrder(arg *NamedArg) (string, error) {
	switch arg.Name {
	case "mail":
		return "`mail`", nil
	}

	return "", ErrFilterArgument
}

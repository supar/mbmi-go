package models

import (
	"database/sql"
)

type Alias struct {
	Id        int64  `json:"id"`
	Alias     string `json:"alias"`
	Recepient string `json:"recipient"`
	Comment   string `json:"comment"`
}

func (s *DB) Aliases(flt FilterIface, cnt bool) (m []*Alias, count uint64, err error) {
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
			expr.CbFunc(aliasWhere)
		case "GROUP BY":
			expr.CbFunc(aliasGroup)
		case "ORDER BY":
			expr.CbFunc(aliasOrder)
		}
	}

	// Base query
	query.raw = "SELECT `a`.`id` `id`" +
		", `a`.`alias` `alias`" +
		", `a`.`recipient` `recipient`" +
		", `a`.`comment` `comment`" +
		" " +
		"FROM `aliases` AS `a` "

	// Add where
	if query_str, args, err = query.Compile(); err != nil {
		return
	}

	if rows, err = s.Query(query_str, args...); err != nil {
		return nil, 0, err
	}

	defer rows.Close()
	// Create empty slice
	m = make([]*Alias, 0)

	for rows.Next() {
		var i = &Alias{}

		err = rows.Scan(
			&i.Id,
			&i.Alias,
			&i.Recepient,
			&i.Comment,
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
			"FROM `aliases` AS `a` "

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

func aliasWhere(arg sql.NamedArg) (string, error) {
	switch arg.Name {
	case "id":
		return "`a`.`id` = ?", nil

	case "alias":
		return "`a`.`alias` = ?", nil

	case "recipient":
		return "`a`.`recipient` LIKE ?", nil
	}

	return "", ErrFilterArgument
}

func aliasGroup(arg sql.NamedArg) (string, error) {
	switch arg.Name {
	case "alias":
		return "`alias`", nil
	}

	return "", ErrFilterArgument
}

func aliasOrder(arg sql.NamedArg) (string, error) {
	switch arg.Name {
	case "id":
		return "`a`.`id`", nil
	}

	return "", ErrFilterArgument
}

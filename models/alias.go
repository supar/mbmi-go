package models

import (
	"database/sql"
)

type Alias struct {
	Id        int64  `json:"id" schema:"id"`
	Alias     Email  `json:"alias" schema: "alias"`
	Recipient Email  `json:"recipient" schema:"recipient"`
	Comment   string `json:"comment" schema:"comment"`
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
			&i.Recipient,
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

	if cnt && len(m) > 0 {
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

func (s *DB) SetAlias(alias *Alias) (err error) {
	if alias.Id > 0 {
		_, err = s.Exec("UPDATE `aliases` SET "+
			"`alias` = ?, `recipient` = ?, `comment` = ? "+
			"WHERE id = ?",
			alias.Alias,
			alias.Recipient,
			alias.Comment,
			alias.Id)
	} else {
		_, err = s.Exec("INSERT INTO `aliases` ("+
			"`alias`, `recipient`, `comment`"+
			") VALUES (?, ?, ?)",
			alias.Alias,
			alias.Recipient,
			alias.Comment)
	}

	return
}

func (s *DB) DelAlias(id int64) (err error) {
	_, err = s.Exec("DELETE FROM `aliases` WHERE `id` = ?", id)

	return
}

func aliasWhere(arg *NamedArg) (string, error) {
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

func aliasGroup(arg *NamedArg) (string, error) {
	switch arg.Name {
	case "alias":
		return "`alias`", nil
	}

	return "", ErrFilterArgument
}

func aliasOrder(arg *NamedArg) (string, error) {
	switch arg.Name {
	case "id":
		return "`a`.`id`", nil
	}

	return "", ErrFilterArgument
}

package models

import (
	"database/sql"
)

type Spam struct {
	Client  string  `json:"client"`
	From    string  `json:"from"`
	Ip      string  `json:"ip"`
	Attempt uint64  `json:"attempt"`
	Index   float64 `json:"index"`
}

func (s *DB) Spam(flt FilterIface, cnt bool) (m []*Spam, count uint64, err error) {
	var (
		interval  interface{}
		query     *Query
		query_str string
		args      []interface{}
		rows      *sql.Rows
	)

	if flt == nil {
		flt = NewFilter()
	}

	flt.Group("client")

	query = flt.(*Query)

	query.raw = "SELECT " +
		" `s`.`client` `client`" +
		", `s`.`from` `from`" +
		", `s`.`ip` `ip`" +
		", SUM(`s`.`spam_victims_score`) `attempt` "

	for _, expr := range query.expressions {
		switch expr.name {
		case "WHERE":
			expr.CbFunc(spamWhere)

			// Find if interval was passed
			for _, a := range expr.args {
				if a.Name == "interval" {
					query.raw += ", (1 - POW(EXP(1), -(SUM(`s`.`spam_victims_score`) / ?))) `index` "
					interval = a.Value[0]

					break
				}
			}

		case "GROUP BY":
			expr.CbFunc(spamGroup)
		case "ORDER BY":
			expr.CbFunc(spamOrder)
		}
	}

	query.raw += "FROM `spammers` AS `s` "

	if query_str, args, err = query.Compile(); err != nil {
		return
	}

	args = append([]interface{}{interval}, args...)
	if rows, err = s.Query(query_str, args...); err != nil {
		return
	}

	defer rows.Close()
	// Create empty slice
	m = make([]*Spam, 0)

	for rows.Next() {
		var i = &Spam{}

		err = rows.Scan(
			&i.Client,
			&i.From,
			&i.Ip,
			&i.Attempt,
			&i.Index,
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
		query.raw = "SELECT 1 " +
			"FROM `spammers` AS `s` "

		query.Un("LIMIT")
		query.Un("ORDER BY")

		args = append([]interface{}{interval}, args...)
		if query_str, args, err = query.Compile(); err != nil {
			return
		}

		query_str = "SELECT COUNT(*) FROM (" + query_str + ") `tmp`"
		err = s.QueryRow(query_str, args...).Scan(&count)

		if err != nil && err == sql.ErrNoRows {
			err = nil
		}
	}

	return
}

func spamWhere(arg *NamedArg) (string, error) {
	switch arg.Name {
	case "client":
		return "`s`.`client` LIKE ?", nil
	case "interval":
		return "`s`.`created` >= NOW() - INTERVAL ? DAY", nil
	}

	return "", ErrFilterArgument
}

func spamGroup(arg *NamedArg) (string, error) {
	switch arg.Name {
	case "client":
		return "`client`", nil
	}

	return "", ErrFilterArgument
}

func spamOrder(arg *NamedArg) (string, error) {
	switch arg.Name {
	case "attempt":
		return "`attempt`", nil
	case "index":
		return "`index`", nil
	}

	return "", ErrFilterArgument
}

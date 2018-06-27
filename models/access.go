package models

import (
	"database/sql"
)

// Access respresents data model for the IP or domain
// access list
type Access struct {
	Client string `json:"client"`
	Access string `json:"access"`
}

// Accesses returns the the list of the rejected or granted addresses or IPs
func (s *DB) Accesses(flt FilterIface, cnt bool) (m []*Access, count uint64, err error) {
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

	query.raw = "SELECT `client`, `access` FROM `client_access`"

	for _, expr := range query.expressions {
		switch expr.name {
		case "WHERE":
			expr.CbFunc(accessWhere)
		case "ORDER BY":
			expr.CbFunc(accessOrder)
		}
	}

	if queryStr, args, err = query.Compile(); err != nil {
		return
	}

	if rows, err = s.Query(queryStr, args...); err != nil {
		return
	}

	defer rows.Close()
	// Create empty slice
	m = make([]*Access, 0)

	for rows.Next() {
		var i = &Access{}

		err = rows.Scan(
			&i.Client,
			&i.Access,
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
		query.raw = "SELECT COUNT(*) `client_access`"

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

func accessWhere(arg *NamedArg) (string, error) {
	switch arg.Name {
	case "search":
		return "`client` LIKE ?", nil

	case "access":
		return "`access` = ?", nil
	}

	return "", ErrFilterArgument
}

func accessOrder(arg *NamedArg) (string, error) {
	switch arg.Name {
	case "client":
		return "`client`", nil

	case "access":
		return "`access`", nil
	}

	return "", ErrFilterArgument
}

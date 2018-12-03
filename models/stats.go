package models

import (
	"database/sql"
	"time"
)

// Stat represents metric item by service usage
type Stat struct {
	UID     int64      `json:"uid" schema:"-"`
	Service string     `json:"service" schema:"-"`
	Email   Email      `json:"-" schema:"email"`
	IP      string     `json:"ip" schema:"ip"`
	Time    *time.Time `json:"updated" schema:"-"`
	Count   int        `json:"attempt" schema:"-"`
}

func (s *DB) ServicesStat(flt FilterIface, cnt bool) (m []*Stat, count uint64, err error) {
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
			expr.CbFunc(serviceWhere)
		case "ORDER BY":
			expr.CbFunc(serviceOrder)
		}
	}

	// Base query
	query.raw = "SELECT `s`.`uid` `uid`" +
		", `s`.`service` `service`" +
		", IF(`s`.`ip` IS NOT NULL, INET_NTOA(`s`.`ip`), '') `ip`" +
		", `s`.`updated` `updated`" +
		", `s`.`attempt` `attempt`" +
		" " +
		"FROM `statistics` AS `s` " +
		"INNER JOIN (SELECT `in_s`.`uid`, MAX(`in_s`.`updated`) `updated` FROM `statistics` AS `in_s` GROUP BY `in_s`.`uid`, `in_s`.`service`) `tmp_s` " +
		"ON `s`.`uid` = `tmp_s`.`uid` AND `s`.`updated` = `tmp_s`.`updated`"

	if queryStr, args, err = query.Compile(); err != nil {
		return
	}

	if rows, err = s.Query(queryStr, args...); err != nil {
		return
	}

	defer rows.Close()
	// Create empty slice
	m = make([]*Stat, 0)

	for rows.Next() {
		var i = &Stat{}

		err = rows.Scan(
			&i.UID,
			&i.Service,
			&i.IP,
			&i.Time,
			&i.Count,
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
			"FROM `statistics` AS `s` "

		query.Un("LIMIT")
		query.Un("ORDER BY")

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

// SetStatImapLogin updates metric data for the service
func (s *DB) SetStatImapLogin(stat *Stat) (err error) {
	var query = "INSERT INTO `statistics` (" +
		"`uid`" +
		", `service`" +
		", `created`" +
		", `ip`" +
		", `updated`" +
		") VALUES (?, ?, NOW(), INET_ATON(?), NOW()) " +
		"ON DUPLICATE KEY UPDATE `attempt` = `attempt` + 1" +
		", `updated` = NOW()"

	_, err = s.Exec(
		query,
		stat.UID,
		stat.Service,
		stat.IP)

	if err != nil {
		return
	}

	return
}

func serviceWhere(arg *NamedArg) (string, error) {
	switch arg.Name {
	case "uid":
		return "`s`.`uid` = ?", nil
	case "ip":
		return "`s`.`ip` = INET_NTOA(?)", nil
	}

	return "", ErrFilterArgument
}

func serviceOrder(arg *NamedArg) (string, error) {
	var dir = arg.First().(string)

	switch arg.Name {
	case "attempt":
		return "`attempt` " + dir, nil
	case "updated":
		return "`updated` " + dir, nil
	}

	return "", ErrFilterArgument
}

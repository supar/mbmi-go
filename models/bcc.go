package models

import (
	"database/sql"
)

type BccItem struct {
	ID        int64  `json:"id" schema:"id"`
	Sender    Email  `json:"sender" schema: "sender"`
	Recipient Email  `json:"recipient" schema:"recipient"`
	Copy      Email  `json:"copy" schema:"copy"`
	Comment   string `json:"comment" schema:"comment"`
}

func (s *DB) Bccs(flt FilterIface, cnt bool) (m []*BccItem, count uint64, err error) {
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
			expr.CbFunc(bccWhere)
		case "GROUP BY":
			expr.CbFunc(bccGroup)
		case "ORDER BY":
			expr.CbFunc(bccOrder)
		}
	}

	// Base query
	query.raw = "SELECT `b`.`id` `id`" +
		", `b`.`sender` `sender`" +
		", `b`.`recipient` `recipient`" +
		", `b`.`copy` `copy`" +
		", `b`.`comment` `comment`" +
		" " +
		"FROM `bcc` AS `b` "

	// Add where
	if query_str, args, err = query.Compile(); err != nil {
		return
	}

	if rows, err = s.Query(query_str, args...); err != nil {
		return nil, 0, err
	}

	defer rows.Close()
	// Create empty slice
	m = make([]*BccItem, 0)

	for rows.Next() {
		var i = &BccItem{}

		err = rows.Scan(
			&i.ID,
			&i.Sender,
			&i.Recipient,
			&i.Copy,
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
			"FROM `bcc` AS `b` "

		query.Un("ORDER BY")
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

func (s *DB) SetBcc(b *BccItem) (err error) {
	if b.ID > 0 {
		_, err = s.Exec("UPDATE `bcc` SET "+
			"`sender` = ?, `recipient` = ?, `copy` = ?, `comment` = ? "+
			"WHERE id = ?",
			b.Sender,
			b.Recipient,
			b.Copy,
			b.Comment,
			b.ID)
	} else {
		_, err = s.Exec("INSERT INTO `bcc` ("+
			"`sender`, `recipient`, `copy`, `comment`"+
			") VALUES (?, ?, ?, ?)",
			b.Sender,
			b.Recipient,
			b.Copy,
			b.Comment)
	}

	return
}

func (s *DB) DelBcc(id int64) (err error) {
	_, err = s.Exec("DELETE FROM `bcc` WHERE `id` = ?", id)

	return
}

func bccWhere(arg *NamedArg) (string, error) {
	switch arg.Name {
	case "id":
		return "`b`.`id` = ?", nil

	case "sender":
		return "`b`.`sender` LIKE ?", nil

	case "recipient":
		return "`b`.`recipient` LIKE ?", nil

	case "copy":
		return "`b`.`copy` LIKE ?", nil

	case "search":
		if arg.Value != nil && len(arg.Value) == 1 {
			arg.Fill(arg.Value[0], 3)
		}
		return "(`b`.`sender` LIKE ? OR `b`.`recipient` LIKE ? OR `b`.`copy` LIKE ?)", nil
	}

	return "", ErrFilterArgument
}

func bccGroup(arg *NamedArg) (string, error) {
	return "", ErrFilterArgument
}

func bccOrder(arg *NamedArg) (string, error) {
	switch arg.Name {
	case "id":
		return "`b`.`id`", nil

	case "sender":
		return "`b`.`sender`", nil

	case "recipient":
		return "`b`.`recipient`", nil

	case "copy":
		return "`b`.`copy`", nil
	}

	return "", ErrFilterArgument
}

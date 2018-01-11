package models

import (
	"database/sql"
	"testing"
)

func Test_NotEmptyFilterWhere(t *testing.T) {
	flt := NewFilter()
	flt.Where("id", 1).
		Where("login", 2)

	query := flt.(*Query)

	if expr := query.Expression("WHERE"); expr != nil {
		if l := len(expr.args); l != 2 {
			t.Errorf("Expected 2 arguments, but got %d", l)
		}
	} else {
		t.Error("Condition WHERE not found")
	}
}

func Test_CompileValidQueryWithWhereOrderLimit(t *testing.T) {
	var (
		str  string
		err  error
		args []interface{}
	)

	flt := NewFilter()

	flt.Where("id", 1).
		Where("login", 2).
		Order("id", true).
		Order("login", false).
		Limit(1, 100)

	query := flt.(*Query)

	for _, expr := range query.expressions {
		switch expr.name {
		case "WHERE":
			expr.callback = func(a sql.NamedArg) (string, error) {
				switch a.Name {
				case "id":
					return "`u`.`user` = ?", nil
				case "login":
					return "`u`.`login` = ?", nil
				}

				return "", UnsupportedFilterArgument
			}

		case "ORDER":
			expr.callback = func(a sql.NamedArg) (string, error) {
				var ascDesc = a.Value.(string)

				switch a.Name {
				case "id":
					return "`u`.`user` " + ascDesc, nil
				case "login":
					return "`u`.`login` " + ascDesc, nil
				}

				return "", UnsupportedFilterArgument
			}
		}
	}

	query.raw = "SELECT * FROM `u`.`users`"

	if str, args, err = query.Compile(); err != nil {
		t.Error(err)
	}

	query_mock := "SELECT * FROM `u`.`users` WHERE `u`.`user` = ? AND `u`.`login` = ? ORDER `u`.`user` ASC,`u`.`login` DESC LIMIT ?,?"

	if query_mock != str {
		t.Errorf("Expecting %s, but got %s", query_mock, str)
	}

	if l := len(args); l != 4 {
		t.Errorf("Expecting 4 arguments, but got %d", l)
	}

	for _, i := range args {
		switch i.(type) {
		case int, int8, int16, int32, int64,
			uint, uint8, uint16, uint32, uint64, uintptr,
			string, bool:

		default:
			t.Errorf("Unsupported type: %T", i)
		}
	}
}

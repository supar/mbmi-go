package models

import (
	"database/sql"
)

// Query interface to set base query expressions:
// WHERE, GROUP BY, ORDER, LIMIT
type FilterIface interface {
	// GROUP BY interface
	Group(string) FilterIface
	Limit(uint64, uint64) FilterIface
	Order(string, bool) FilterIface
	Where(string, interface{}) FilterIface
}

// Create query object and return filter interface
//
// To build sql string there must be defined expression's
// callback and raw statement as main query string part
// Callback function must identify expression argument and
// return valid sql part
//
// Example:
// flt.Where("id", 1)
//
// cbFunc = func(arg sql.NamedArg) (string, error) {
//     switch args.Name {
//     case "id":
//         return "`userTable`.`id` = ?". nil
//     }
//
//     return "", UnsupportedFilterArgument
// }
func NewFilter() FilterIface {
	return &Query{
		expressions: make([]*expression, 0),
	}
}

func (s *Query) Group(name string) FilterIface {
	var expr *expression

	if expr = s.expression("GROUP BY"); expr == nil {
		expr = &expression{
			name:       "GROUP BY",
			glue:       ",",
			order:      3,
			pushValues: false,
			args:       make([]sql.NamedArg, 0),
		}

		s.expressions = append(s.expressions, expr)
	}

	expr.set(name, nil)

	return s
}

func (s *Query) Limit(linit, offset uint64) FilterIface {
	var expr *expression

	if expr = s.expression("LIMIT"); expr == nil {
		expr = &expression{
			name:       "LIMIT",
			glue:       ",",
			order:      7,
			pushValues: true,
			args:       make([]sql.NamedArg, 2),
			callback: func(arg sql.NamedArg) (string, error) {
				switch arg.Name {
				case "rowslimit":
					return "?", nil

				case "rowsoffset":
					return "?", nil

				}

				return "", UnsupportedFilterArgument
			},
		}

		s.expressions = append(s.expressions, expr)
	}

	expr.args = []sql.NamedArg{
		sql.NamedArg{
			Name:  "rowslimit",
			Value: limit,
		},
		sql.NamedArg{
			Name:  "rowsoffset",
			Value: offset,
		},
	}

	return s
}

func (s *Query) Order(name string, order bool) FilterIface {
	var (
		expr    *expression
		ascDesc string
	)

	if order {
		ascDesc = "ASC"
	} else {
		ascDesc = "DESC"
	}

	if expr = s.expression("ORDER"); expr == nil {
		expr = &expression{
			name:       "ORDER",
			glue:       ",",
			order:      4,
			pushValues: false,
			args:       make([]sql.NamedArg, 0),
		}

		s.expressions = append(s.expressions, expr)
	}

	expr.set(name, ascDesc)

	return s
}

func (s *Query) Where(name string, v interface{}) FilterIface {
	var expr *expression

	if expr = s.expression("WHERE"); expr == nil {
		expr = &expression{
			name:       "WHERE",
			glue:       " AND ",
			order:      2,
			pushValues: true,
			args:       make([]sql.NamedArg, 0),
		}

		s.expressions = append(s.expressions, expr)
	}

	expr.set(name, v)

	return s
}

func (s *expression) set(name string, v interface{}) {
	if name != "" {
		s.args = append(s.args, sql.NamedArg{
			Name:  name,
			Value: v,
		})
	}
}

func (s *expression) each(fn namedArgFunc) (expr []string, args []interface{}, err error) {
	expr = make([]string, 0)
	args = make([]interface{}, 0)

	for _, i := range s.args {
		var str string

		if str, err = fn(i); err != nil {
			return
		}

		if str == "" {
			continue
		}

		expr = append(expr, str)
		args = append(args, i.Value)
	}

	return
}

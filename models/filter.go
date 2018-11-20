package models

// This is equivalent to the sql.NamedArg function othervise
// Value is the slice of interfaces
type NamedArg struct {
	Name  string
	Value []interface{}
}

// Query interface to set base query expressions:
// WHERE, GROUP BY, ORDER, LIMIT
type FilterIface interface {
	// GROUP BY interface
	Group(string) FilterIface
	Limit(uint64, uint64) FilterIface
	Order(string, bool) FilterIface
	Where(string, interface{}) FilterIface
}

func (n *NamedArg) Set(v ...interface{}) {
	var data []interface{}

	if v != nil && len(v) > 0 {
		data = make([]interface{}, 0)

		for _, item := range v {
			if item == nil {
				continue
			}

			data = append(data, item)
		}
	}

	n.Value = data
}

// First returns first value from the Values
func (n *NamedArg) First() (v interface{}) {
	if n.Value != nil && len(n.Value) > 0 {
		return n.Value[0]
	}

	return ""
}

func (n *NamedArg) Fill(v interface{}, repeat int) {
	if repeat < 2 {
		return
	}

	var data = make([]interface{}, 0, repeat)

	for i := 0; i < repeat; i++ {
		data = append(data, v)
	}

	n.Set(data...)
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
// cbFunc = func(arg namedArg) (string, error) {
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

	if _, expr = s.Expression("GROUP BY"); expr == nil {
		expr = &expression{
			name:       "GROUP BY",
			glue:       ",",
			order:      3,
			pushValues: false,
			args:       make([]NamedArg, 0),
		}

		s.expressions = append(s.expressions, expr)
	}

	expr.set(name, nil)

	return s
}

func (s *Query) Limit(limit, offset uint64) FilterIface {
	var expr *expression

	if _, expr = s.Expression("LIMIT"); expr == nil {
		expr = &expression{
			name:       "LIMIT",
			glue:       ",",
			order:      7,
			pushValues: true,
			callback: func(arg *NamedArg) (string, error) {
				switch arg.Name {
				case "rowslimit":
					return "?", nil

				case "rowsoffset":
					return "?", nil

				}

				return "", ErrFilterArgument
			},
		}

		s.expressions = append(s.expressions, expr)
	}

	expr.args = make([]NamedArg, 0, 2)
	expr.set("rowsoffset", offset)
	expr.set("rowslimit", limit)

	return s
}

func (s *Query) Un(name string) FilterIface {
	var (
		expr *expression
		idx  int
	)

	if name != "" {
		if idx, expr = s.Expression(name); expr != nil {
			s.expressions = append(s.expressions[:idx], s.expressions[idx+1:]...)
		}
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

	if _, expr = s.Expression("ORDER BY"); expr == nil {
		expr = &expression{
			name:       "ORDER BY",
			glue:       ",",
			order:      4,
			pushValues: false,
			args:       make([]NamedArg, 0),
		}

		s.expressions = append(s.expressions, expr)
	}

	expr.set(name, ascDesc)

	return s
}

func (s *Query) Where(name string, v interface{}) FilterIface {
	var expr *expression

	if _, expr = s.Expression("WHERE"); expr == nil {
		expr = &expression{
			name:       "WHERE",
			glue:       " AND ",
			order:      2,
			pushValues: true,
			args:       make([]NamedArg, 0),
		}

		s.expressions = append(s.expressions, expr)
	}

	expr.set(name, v)

	return s
}

func (s *expression) set(name string, v ...interface{}) {
	if name != "" {
		var arg = NamedArg{
			Name: name,
		}

		arg.Set(v...)
		s.args = append(s.args, arg)
	}
}

func (s *expression) each(fn namedArgFunc) (expr []string, args []interface{}, err error) {
	expr = make([]string, 0)
	args = make([]interface{}, 0)

	for _, i := range s.args {
		var str string

		if str, err = fn(&i); err != nil {
			return
		}

		if str == "" {
			continue
		}

		expr = append(expr, str)

		if i.Value != nil {
			args = append(args, i.Value...)
		}
	}

	return
}

package models

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"
)

type Query struct {
	raw         string
	expressions []*expression
}

type expression struct {
	name       string
	glue       string
	order      uint8
	pushValues bool
	args       []sql.NamedArg
	callback   namedArgFunc
}

type rowScanIface interface {
	Scan(...interface{}) error
}

type namedArgFunc func(sql.NamedArg) (string, error)

type exprOrder []*expression

func (s exprOrder) Len() int           { return len(s) }
func (s exprOrder) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s exprOrder) Less(i, j int) bool { return s[i].order < s[j].order }

func (s *expression) CbFunc(fn namedArgFunc) *expression {
	s.callback = fn
	return s
}

func (s *Query) Compile() (query string, args []interface{}, err error) {
	query = s.raw

	sort.Sort(exprOrder(s.expressions))

	for _, i := range s.expressions {
		var (
			str []string
			a   []interface{}
		)

		if i == nil {
			continue
		}

		if i.callback == nil {
			err = errors.New("Callback required for the `" + i.name + "`")
			return
		}

		if str, a, err = i.each(i.callback); err != nil {
			return
		}

		if len(str) == 0 {
			continue
		}

		query += " " + i.name + " " + strings.Join(str, i.glue)

		if i.pushValues {
			args = append(args, a...)
		}
	}

	return
}

func (s *Query) Expression(name string) (int, *expression) {
	for idx, item := range s.expressions {
		if item.name == name {
			return idx, item
		}
	}

	return -1, nil
}

func limit(q string, ctx context.Context) string {
	var (
		limit, offset uint64
		rq            url.Values
	)

	limit = 100

	if i := ctx.Value("Query"); i != nil {
		rq = i.(url.Values)

		if v := rq.Get("limit"); v != "" {
			limit, _ = strconv.ParseUint(v, 10, 64)
		}

		if v := rq.Get("offset"); v != "" {
			offset, _ = strconv.ParseUint(v, 10, 64)
		}
	}

	return fmt.Sprintf("%s LIMIT %d, %d", q, offset, limit)
}

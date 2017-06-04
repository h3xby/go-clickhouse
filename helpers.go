package clickhouse

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
)

type (
	Column  string
	Columns []string
	Row     []interface{}
	Rows    []Row
	Array   []interface{}
)

func NewHttpTransport() HttpTransport {
	return HttpTransport{}
}

func NewConn(host string, t Transport) *Conn {
	if strings.Index(host, "http://") < 0 && strings.Index(host, "https://") < 0 {
		host = "http://" + host
	}
	connUrl, err := url.Parse(host)
	if err != nil {
		return nil
	}
	connUrl.Path = strings.TrimRight(connUrl.Path, "/") + "/"
	host = fmt.Sprintf("%s://%s%s", connUrl.Scheme, connUrl.Host, connUrl.Path)

	values, err := url.ParseQuery(connUrl.RawQuery)
	if err != nil {
		values = url.Values{}
	}

	return &Conn{
		Host:      host,
		Params:    values,
		transport: t,
	}
}

func NewQuery(stmt string, args ...interface{}) Query {
	return Query{
		Stmt: stmt,
		args: args,
	}
}

func BuildInsert(tbl string, cols Columns, row Row) (Query, error) {
	return BuildMultiInsert(tbl, cols, Rows{row})
}

func BuildMultiInsert(tbl string, cols Columns, rows Rows) (Query, error) {
	var (
		stmt string
		args []interface{}
	)

	if len(cols) == 0 || len(rows) == 0 {
		return Query{}, errors.New("rows and cols cannot be empty")
	}

	colCount := len(cols)
	rowCount := len(rows)
	args = make([]interface{}, colCount*rowCount)
	argi := 0

	for _, row := range rows {
		if len(row) != colCount {
			return Query{}, errors.New("Amount of row items does not match column count")
		}
		for _, val := range row {
			args[argi] = val
			argi++
		}
	}

	binds := strings.Repeat("?,", colCount)
	binds = "(" + binds[:len(binds)-1] + "),"
	batch := strings.Repeat(binds, rowCount)
	batch = batch[:len(batch)-1]

	stmt = fmt.Sprintf("INSERT INTO %s (%s) VALUES %s", tbl, strings.Join(cols, ","), batch)

	return NewQuery(stmt, args...), nil
}

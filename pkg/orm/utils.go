package orm

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

// SQLGetParams one sql get query
type SQLSelectParams struct {
	What    string
	Table   string
	Options SQLOption
	Joins   []SQLJoin
	Args    []interface{}
}

// SQLDeleteParams one sql delete query
type SQLDeleteParams struct {
	Table   string
	Options SQLOption
}

// SQLUpdateParams one sql update query
type SQLUpdateParams struct {
	Table   string
	Couples map[string]string
	Options SQLOption
}

// SQLInsertParams one sql insert query
type SQLInsertParams struct {
	Table  string
	Datas  string
	Values []interface{}
}

// SQLOption set option
type SQLOption struct {
	Where string
	Order string
	Limit string
	Args  []interface{}
}

// SQLJoin is joins implement
type SQLJoin struct {
	JoinType     string
	JoinTable    string
	Intersection string
	Args         []interface{}
}

// const sql
const (
	INSERTQ = "INSERT INTO table "
	SELECTQ = "SELECT what "
	UPDATEQ = "UPDATE table "
	DELETEQ = "DELETE FROM table "

	WHEREQ = "WHERE condition "
	ORDERQ = "ORDER BY order "
	LIMITQ = "LIMIT limit "

	VALUESQ = "VALUES(values) "
	SETQ    = "SET values "
	FROMQ   = "FROM table "
	INJOINQ = "INNER JOIN table "
	LOJOINQ = "LEFT OUTER JOIN table "
	ONQ     = "ON intersection "
)

/* ------------------------------------ INSERT -------------------------------------- */
func prepareInsertQueryAndArgs(params SQLInsertParams) string {
	return strings.ReplaceAll(INSERTQ, "table", params.Table) + strings.ReplaceAll(VALUESQ, "values", params.Datas)
}

func insertSQL(params SQLInsertParams) (sql.Result, error) {
	st, e := ConnToDB.PrepareContext(
		context.Background(),
		prepareInsertQueryAndArgs(params),
	)
	if e != nil {
		return nil, e
	}
	return st.Exec(params.Values...)
}

func insertBySelect(table, datas string, values []interface{}, op SQLOption) (sql.Result, error) {
	q := strings.ReplaceAll(INSERTQ, "table", table) + strings.ReplaceAll(SELECTQ, "what", "null,"+strings.ReplaceAll(fmt.Sprint(values...), " ", ","))
	if op.Where != "" {
		q += strings.ReplaceAll(WHEREQ, "condition", op.Where)
	}
	return ConnToDB.ExecContext(context.Background(), q, op.Args...)
}

/* ------------------------------------ UPDATE -------------------------------------- */
func prepareUpdateQueryAndArgs(params SQLUpdateParams) string {
	values := ""
	for k, v := range params.Couples {
		values += k + "='" + v + "',"
	}
	values = values[:len(values)-1]

	q := strings.ReplaceAll(UPDATEQ, "table", params.Table) + strings.ReplaceAll(SETQ, "values", values)
	if params.Options.Where != "" {
		q += strings.ReplaceAll(WHEREQ, "condition", params.Options.Where)
	}
	if params.Options.Order != "" {
		q += strings.ReplaceAll(ORDERQ, "order", params.Options.Order)
	}
	if params.Options.Limit != "" {
		q += strings.ReplaceAll(LIMITQ, "limit", params.Options.Limit)
	}
	return q
}

func updateSQL(params SQLUpdateParams) (sql.Result, error) {
	q := prepareUpdateQueryAndArgs(params)
	r, e := ConnToDB.ExecContext(context.Background(), q, params.Options.Args...)
	if e != nil {
		return nil, errors.New("not updated")
	}
	return r, nil
}

/* ------------------------------------ DELETE -------------------------------------- */
func prepareDeleteQueryAndArgs(params SQLDeleteParams) string {
	q := strings.ReplaceAll(DELETEQ, "table", params.Table)
	if params.Options.Where != "" {
		q += strings.ReplaceAll(WHEREQ, "condition", params.Options.Where)
	}
	if params.Options.Order != "" {
		q += strings.ReplaceAll(ORDERQ, "order", params.Options.Order)
	}
	if params.Options.Limit != "" {
		q += strings.ReplaceAll(LIMITQ, "limit", params.Options.Limit)
	}
	return q
}

func deleteSQL(params SQLDeleteParams) (sql.Result, error) {
	q := prepareDeleteQueryAndArgs(params)
	r, e := ConnToDB.ExecContext(context.Background(), q, params.Options.Args...)
	if e != nil {
		return nil, errors.New("not deleted")
	}
	return r, nil
}

/* ------------------------------------ SELECT -------------------------------------- */
func prepareGetQueryAndArgs(params SQLSelectParams) (string, []interface{}) {
	q := strings.ReplaceAll(SELECTQ, "what", params.What) + strings.ReplaceAll(FROMQ, "table", params.Table)

	args := params.Args
	if args == nil {
		args = []interface{}{}
	}
	if params.Joins != nil {
		for _, v := range params.Joins {
			args = append(args, v.Args...)
			q += strings.ReplaceAll(v.JoinType, "table", v.JoinTable) + strings.ReplaceAll(ONQ, "intersection", v.Intersection)
		}
	}

	if params.Options.Where != "" {
		q += strings.ReplaceAll(WHEREQ, "condition", params.Options.Where)
	}
	if params.Options.Order != "" {
		q += strings.ReplaceAll(ORDERQ, "order", params.Options.Order)
	}
	if params.Options.Limit != "" {
		q += strings.ReplaceAll(LIMITQ, "limit", params.Options.Limit)
	}
	args = append(args, params.Options.Args...)

	return q, args
}

func selectSQL(query string, args []interface{}) ([][]interface{}, error) {
	// fmt.Println(query, args)
	// fmt.Println("-----------------------")
	res := [][]interface{}{}
	rows, e := ConnToDB.QueryContext(context.Background(), query, args...)
	if e != nil {
		return nil, e
	}
	cols, _ := rows.Columns()

	for rows.Next() {
		currentRow := make([]interface{}, len(cols))
		pointers := make([]interface{}, len(cols))
		for i := range currentRow {
			pointers[i] = &currentRow[i]
		}

		if e = rows.Scan(pointers...); e != nil {
			return nil, e
		}
		res = append(res, currentRow)
	}
	rows.Close()
	return res, nil
}

/*
	general functions: crud op funcs
*/

package dbfuncs

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

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

// insert - insert new note to all type
func insert(table, whatInsert string, values []interface{}) (sql.Result, error) {
	st, e := ConnToDB.PrepareContext(context.Background(), strings.ReplaceAll(INSERTQ, "table", table)+
		strings.ReplaceAll(VALUESQ, "values", whatInsert))
	if e != nil {
		return nil, e
	}
	return st.Exec(values...)
}

// insertBySelect - insert by select
func insertBySelect(table string, values []interface{}, op SQLOption) (sql.Result, error) {
	q := strings.ReplaceAll(INSERTQ, "table", table) + strings.ReplaceAll(SELECTQ, "what", "null,"+strings.ReplaceAll(fmt.Sprint(values...), " ", ","))
	if op.Where != "" {
		q += strings.ReplaceAll(WHEREQ, "condition", op.Where)
	}
	return ConnToDB.ExecContext(context.Background(), q, op.Args...)
}

// update - update all using table, set values, condition
func update(table, couples string, op SQLOption) (sql.Result, error) {
	q := strings.ReplaceAll(UPDATEQ, "table", table) + strings.ReplaceAll(SETQ, "values", couples)
	if op.Where != "" {
		q += strings.ReplaceAll(WHEREQ, "condition", op.Where)
	}
	if op.Order != "" {
		q += strings.ReplaceAll(ORDERQ, "order", op.Order)
	}
	if op.Limit != "" {
		q += strings.ReplaceAll(LIMITQ, "limit", op.Limit)
	}
	r, e := ConnToDB.ExecContext(context.Background(), q, op.Args...)
	return r, e
}

// deleteSQL ...
func deleteSQL(table string, op SQLOption) (sql.Result, error) {
	q := strings.ReplaceAll(DELETEQ, "table", table)
	if op.Where != "" {
		q += strings.ReplaceAll(WHEREQ, "condition", op.Where)
	}
	if op.Order != "" {
		q += strings.ReplaceAll(ORDERQ, "order", op.Order)
	}
	if op.Limit != "" {
		q += strings.ReplaceAll(LIMITQ, "limit", op.Limit)
	}
	return ConnToDB.ExecContext(context.Background(), q, op.Args...)
}

// get - wrapper to sql select
func get(what, table string, op SQLOption, joins []SQLJoin) ([][]interface{}, error) {
	res := [][]interface{}{}
	q := strings.ReplaceAll(SELECTQ, "what", what) + strings.ReplaceAll(FROMQ, "table", table)

	args := []interface{}{}
	if joins != nil {
		for _, v := range joins {
			args = append(args, v.Args...)
			q += strings.ReplaceAll(v.JoinType, "table", v.JoinTable) + strings.ReplaceAll(ONQ, "intersection", v.Intersection)
		}
	}

	if op.Where != "" {
		q += strings.ReplaceAll(WHEREQ, "condition", op.Where)
	}
	if op.Order != "" {
		q += strings.ReplaceAll(ORDERQ, "order", op.Order)
	}
	if op.Limit != "" {
		q += strings.ReplaceAll(LIMITQ, "limit", op.Limit)
	}
	args = append(args, op.Args...)

	rows, e := ConnToDB.QueryContext(context.Background(), q, args...)
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
		e = rows.Scan(pointers...)
		if e != nil {
			return nil, e
		}
		res = append(res, currentRow)
	}
	rows.Close()
	return res, nil
}

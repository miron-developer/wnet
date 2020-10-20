/*
	This file define "graphql" query
*/

package dbfuncs

import (
	"errors"
)

// GetFrom get more than one
func GetFrom(table, what string, op SQLOption, joins []SQLJoin) ([][]interface{}, error) {
	if what == "" || table == "" {
		return nil, errors.New("n/d")
	}
	result, e := get(what, table, op, joins)
	if len(result) == 0 {
		return nil, errors.New("n/d")
	}
	return result, e
}

// GetOneFrom get one value
func GetOneFrom(table, what string, op SQLOption, joins []SQLJoin) ([]interface{}, error) {
	result, e := GetFrom(table, what, op, joins)
	if e != nil {
		return nil, e
	}
	return result[0], e
}

// DoSQLOption create new sqloption & return
func DoSQLOption(where, order, limit string, args ...interface{}) SQLOption {
	return SQLOption{Where: where, Order: order, Limit: limit, Args: args}
}

// DoSQLJoin create new sqloption & return
func DoSQLJoin(jtype, jtable, inter string, args ...interface{}) SQLJoin {
	return SQLJoin{JoinType: jtype, JoinTable: jtable, Intersection: inter, Args: args}
}

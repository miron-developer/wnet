package dbfuncs

import (
	"errors"
)

// GetFrom get more than one
func GetFrom(params SQLSelectParams) ([][]interface{}, error) {
	if params.What == "" || params.Table == "" {
		return nil, errors.New("n/d")
	}
	result, e := selectSQL(prepareGetQueryAndArgs(params))
	if len(result) == 0 || e != nil {
		return nil, errors.New("n/d")
	}
	return result, nil
}

// GetOneFrom get one value
func GetOneFrom(params SQLSelectParams) ([]interface{}, error) {
	result, e := GetFrom(params)
	if e != nil {
		return nil, e
	}
	return result[0], e
}

// GetWithSubqueries nested querys
func GetWithSubqueries(mainQ SQLSelectParams, querys []SQLSelectParams, as []string, sampleStruct interface{}) ([]map[string]interface{}, error) {
	if querys == nil || len(querys) == 0 {
		return nil, errors.New("n/d")
	}
	for _, v := range querys {
		curQ, curArgs := prepareGetQueryAndArgs(v)
		mainQ.What += ", (" + curQ + ")"
		mainQ.Args = append(mainQ.Args, curArgs...)
	}

	result, e := GetFrom(mainQ)
	if len(result) == 0 || e != nil {
		return nil, errors.New("n/d")
	}
	return MapFromStructAndMatrix(result, sampleStruct, as...), nil
}

// GetWithQueryAndArgs get with query and args
func GetWithQueryAndArgs(query string, args []interface{}) ([][]interface{}, error) {
	return selectSQL(query, args)
}

// DoSQLOption create new sqloption & return
func DoSQLOption(where, order, limit string, args ...interface{}) SQLOption {
	return SQLOption{Where: where, Order: order, Limit: limit, Args: args}
}

// DoSQLJoin create new sqljoin & return
func DoSQLJoin(jtype, jtable, inter string, args ...interface{}) SQLJoin {
	return SQLJoin{JoinType: jtype, JoinTable: jtable, Intersection: inter, Args: args}
}

package app

import (
	"anifor/app/dbfuncs"
	"encoding/json"
	"net/http"
	"strconv"
)

type getAPIOption struct {
	tableName string
	sample    interface{}
	whatGet   string
	op        dbfuncs.SQLOption
	joins     []dbfuncs.SQLJoin
}

// do json and write it
func doJS(w http.ResponseWriter, data interface{}) {
	js, _ := json.Marshal(data)
	w.Header().Set("Content-Type", "Application/json")
	w.Write(js)
}

func getLimit(r *http.Request) (int, int) {
	first, e := strconv.Atoi(r.FormValue("firstIndex"))
	if e != nil {
		first = 0
	}
	count, e := strconv.Atoi(r.FormValue("count"))
	if e != nil {
		count = 10
	}
	return first, count
}

func generalGet(w http.ResponseWriter, r *http.Request, getOp getAPIOption, joinArgs ...interface{}) {
	results, e := dbfuncs.GetFrom(getOp.tableName, getOp.whatGet, getOp.op, getOp.joins)
	if e != nil && len(results) != 0 {
		return
	}
	doJS(w, dbfuncs.MapFromStructAndMatrix(results, getOp.sample, joinArgs...))
}

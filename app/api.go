package app

import (
	"encoding/json"
	"net/http"
	"strconv"
	"wnet/app/dbfuncs"
)

// do json and write it
func doJS(w http.ResponseWriter, data interface{}) {
	js, _ := json.Marshal(data)
	w.Header().Set("Content-Type", "Application/json")
	w.Write(js)
}

func getLimit(r *http.Request) (int, int) {
	first, e := strconv.Atoi(r.FormValue("from"))
	if e != nil {
		first = 0
	}
	count, e := strconv.Atoi(r.FormValue("step"))
	if e != nil {
		count = 10
	}
	return first, count
}

func generalGet(w http.ResponseWriter, r *http.Request, selectParams dbfuncs.SQLSelectParams, sampleStruct interface{}, additionalFields ...string) []map[string]interface{} {
	results, e := dbfuncs.GetFrom(selectParams)
	if e != nil || len(results) == 0 {
		return []map[string]interface{}{}
	}
	return dbfuncs.MapFromStructAndMatrix(results, sampleStruct, additionalFields...)
}

func isHaveAccessToPost(id, userID int) bool {
	q := `SELECT 
			userID == ? OR
			CASE postType
				WHEN "public"
					THEN 1
				WHEN "private"
					THEN (
						SELECT id IS NOT NULL FROM Relations 
						WHERE (
							(userID IS NOT NULL AND ((senderUserID=? AND receiverUserID = userID) OR (receiverUserID=? AND senderUserID = userID AND value = 0))) OR
							(groupID IS NOT NULL AND (senderUserID=? AND receiverGroupID = groupID))
						)
					)
				WHEN "almost_private"
					THEN instr(allowedUsers, ?)
			END
		FROM Posts WHERE id = ?`
	args := []interface{}{userID, userID, userID, userID, userID, id}

	res, e := dbfuncs.GetWithQueryAndArgs(q, args)
	if e != nil || res == nil || (res != nil && (res[0][0] == 0 || res[0][0] == nil)) {
		return false
	}
	return true
}

func isHaveAccess(id, userID int, table string) bool {
	q := `SELECT
			userID == ? OR
			(
				SELECT id IS NOT NULL FROM Relations 
				WHERE (
					(userID IS NOT NULL AND ((senderUserID=? AND receiverUserID = userID) OR (receiverUserID=? AND senderUserID = userID AND value = 0))) OR
					(groupID IS NOT NULL AND (senderUserID=? AND receiverGroupID = groupID))
				)
			)
		FROM ` + table + ` WHERE id = ?`
	args := []interface{}{userID, userID, userID, userID, id}

	res, e := dbfuncs.GetWithQueryAndArgs(q, args)
	if e != nil || res == nil || (res != nil && (res[0][0] == 0 || res[0][0] == nil)) {
		return false
	}
	return true
}

func userJoin(joinCondition string) dbfuncs.SQLJoin {
	return dbfuncs.DoSQLJoin(dbfuncs.LOJOINQ, "Users AS u", joinCondition)
}

func groupJoin(joinCondition string) dbfuncs.SQLJoin {
	return dbfuncs.DoSQLJoin(dbfuncs.LOJOINQ, "Groups AS g", joinCondition)
}

func likeJoin(joinCondition string) dbfuncs.SQLJoin {
	return dbfuncs.DoSQLJoin(dbfuncs.LOJOINQ, "Likes AS l", joinCondition)
}

func carmaQ(column string, columnID int) dbfuncs.SQLSelectParams {
	return dbfuncs.SQLSelectParams{
		Table:   "Likes",
		What:    "COUNT(id)",
		Options: dbfuncs.DoSQLOption(column+" = ?", "", "", columnID),
	}
}

func eventAnswerQ(answer, eventID int) dbfuncs.SQLSelectParams {
	return dbfuncs.SQLSelectParams{
		Table:   "EventAnswers",
		What:    "COUNT(id)",
		Options: dbfuncs.DoSQLOption("answer = ? AND eventID = ?", "", "", answer, eventID),
	}
}

package app

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"wnet/app/dbfuncs"
)

const (
	FOLLOWING_USER_ONE_Q   string = "SELECT receiverUserID FROM Relations WHERE senderUserID = ? AND (value = 1 OR value = 0)"
	FOLLOWING_USER_BOTH_Q  string = "SELECT senderUserID FROM Relations WHERE receiverUserID = ? AND value = 0"
	FOLLOWING_USER_GROUP_Q string = "SELECT receiverGroupID FROM Relations WHERE senderUserID = ? AND value = 1"
	FOLLOWERS_USER_ONE_Q   string = "SELECT senderUserID FROM Relations WHERE receiverUserID = ? AND (value = 1 OR value = 0)"
	FOLLOWERS_USER_BOTH_Q  string = "SELECT receiverUserID FROM Relations WHERE senderUserID = ? AND value = 0"
	FOLLOWERS_USER_GROUP_Q string = "SELECT receiverGroupID FROM Relations WHERE senderUserID = ? AND value = 1"
	FOLLOWERS_GROUP_ONE_Q  string = "SELECT senderUserID FROM Relations WHERE receiverGroupID = ? AND (value = 1 OR value = 0)"
	FOLLOWERS_GROUP_BOTH_Q string = "SELECT receiverUserID FROM Relations WHERE senderGroupID = ? AND value = 0"
	REQUEST_USER_Q         string = "SELECT senderUserID FROM Relations WHERE receiverUserID = ? AND value=-1"
	REQUEST_GROUP_Q        string = "SELECT senderGroupID FROM Relations WHERE receiverUserID = ? AND value=-1"
)

// do json and write it
func doJS(w http.ResponseWriter, data interface{}) {
	js, _ := json.Marshal(data)
	w.Header().Set("Content-Type", "Application/json")
	w.Write(js)
}

// get first&count
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

func carmaCountQ(column string, columnID int) dbfuncs.SQLSelectParams {
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

func getUserID(w http.ResponseWriter, r *http.Request, reqID string) (int, error) {
	if reqID != "" {
		return strconv.Atoi(reqID)
	}
	userID := getUserIDfromReq(w, r)
	if userID == -1 {
		return -1, errors.New("not logged")
	}
	return userID, nil
}

func (app *Application) News(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	userID, e := getUserID(w, r, r.FormValue("id"))
	if e != nil {
		return nil, errors.New("wrong id")
	}

	first, count := getLimit(r)

	// res ex: [[4 post 1588670115000] [3 event 1588670115000]]
	newsIDS, e := dbfuncs.GetWithQueryAndArgs(
		`SELECT id, type, datetime FROM Posts 
			WHERE
				userID = ? OR 
				userID IN(`+FOLLOWING_USER_ONE_Q+`) OR
				userID IN(`+FOLLOWING_USER_BOTH_Q+`) OR
				groupID IN(`+FOLLOWING_USER_GROUP_Q+`)
			UNION ALL
			SELECT id, type, datetime FROM Events 
			WHERE 
				userID = ? OR 
				userID IN(`+FOLLOWING_USER_ONE_Q+`) OR
				userID IN(`+FOLLOWING_USER_BOTH_Q+`) OR
				groupID IN(`+FOLLOWING_USER_GROUP_Q+`) 
			ORDER BY datetime DESC LIMIT ?,?`,
		[]interface{}{userID, userID, userID, userID, userID, userID, userID, userID, first, count},
	)
	if e != nil {
		return nil, errors.New("wrong data")
	}

	return dbfuncs.MapFromStructAndMatrix(newsIDS, struct {
		ID       int    `json:"id"`
		Type     string `json:"type"`
		Datetime int    `json:"datetime"`
	}{}), nil
}

func (app *Application) Publications(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	ID, e := getUserID(w, r, r.FormValue("id"))
	if e != nil {
		return nil, errors.New("wrong id")
	}

	first, count := getLimit(r)
	publType := r.FormValue("publicationType")
	which := r.FormValue("type")
	if which != "user" && which != "group" {
		return nil, errors.New("wrong type")
	}

	q := ""
	args := []interface{}{ID}
	if publType == "post" {
		q = "SELECT id, type, datetime FROM Posts WHERE " + which + "ID=? "
	} else if publType == "event" {
		q = "SELECT id, type, datetime FROM Events WHERE " + which + "ID=? "
	} else {
		q = `SELECT id, type, datetime FROM Posts 
			WHERE ` + which + `ID=?
			UNION ALL
			SELECT id, type, datetime FROM Events 
			WHERE ` + which + `ID=? `
		args = append(args, ID)
	}
	q += "ORDER BY datetime DESC LIMIT ?,?"
	args = append(args, first, count)

	// res ex: [[4 post 1588670115000] [3 event 1588670115000]]
	publIDS, e := dbfuncs.GetWithQueryAndArgs(q, args)
	if e != nil {
		return nil, errors.New("wrong data")
	}

	return dbfuncs.MapFromStructAndMatrix(publIDS, struct {
		ID       int    `json:"id"`
		Type     string `json:"type"`
		Datetime int    `json:"datetime"`
	}{}), nil
}

func (app *Application) Notifications(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	ID, e := getUserID(w, r, r.FormValue("id"))
	if e != nil {
		return nil, errors.New("wrong id")
	}

	first, count := getLimit(r)
	return generalGet(
		w,
		r,
		dbfuncs.SQLSelectParams{
			Table: "Notifications",
			What:  "id, type",
			Options: dbfuncs.DoSQLOption(
				`receiverUserID=? OR
				senderUserID IN(`+FOLLOWING_USER_ONE_Q+`) OR
				senderUserID IN(`+FOLLOWING_USER_BOTH_Q+`)`,
				"datetime DESC",
				"?,?",
				ID, ID, ID, first, count,
			),
		},
		struct {
			ID   int `json:"id"`
			Type int `json:"type"`
		}{},
	), nil
}

func (app *Application) Gallery(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	ID, e := getUserID(w, r, r.FormValue("ownerID"))
	if e != nil {
		return nil, errors.New("wrong id")
	}

	first, count := getLimit(r)
	isUser := true
	if r.FormValue("type") == "group" {
		isUser = false
	}

	op := dbfuncs.DoSQLOption("", "m.datetime DESC", "?,?")
	if isUser {
		op.Where = "m.userID = ?"
	} else {
		op.Where = "m.groupID = ?"
	}

	galleryType := r.FormValue("galleryType")
	if galleryType == "all" {
		op.Where += ` AND (m.type="photo" OR m.type=?)`
		galleryType = "video"
	} else {
		op.Where += " AND m.type=?"
	}

	op.Args = append(op.Args, ID, galleryType, first, count)

	getJoin := func(isUser bool) (dbfuncs.SQLJoin, string) {
		if isUser {
			return userJoin("m.userID=u.id"), "u.nName, u.ava"
		}
		return groupJoin("m.groupID=g.id"), "g.title, g.ava"
	}
	join, getStr := getJoin(isUser)

	return generalGet(
		w,
		r,
		dbfuncs.SQLSelectParams{
			Table:   "Media as m",
			What:    "m.*, " + getStr,
			Options: op,
			Joins:   []dbfuncs.SQLJoin{join},
		},
		dbfuncs.Media{},
		"name", "avatar",
	), nil
}

func (app *Application) Users(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	ID, e := getUserID(w, r, r.FormValue("id"))
	if e != nil {
		return nil, errors.New("wrong id")
	}

	first, count := getLimit(r)

	usersType := r.FormValue("type")
	op := dbfuncs.DoSQLOption(
		"",
		"nName ASC",
		"?,?",
		ID, ID, first, count,
	)

	if usersType == "followers" {
		flwType := r.FormValue("flwType")
		if flwType == "all" {
			op.Where = "id IN(" + FOLLOWERS_USER_ONE_Q + ") OR id IN(" + FOLLOWERS_USER_BOTH_Q + ")"
		} else if flwType == "online" {
			op.Where = "(id IN(" + FOLLOWERS_USER_ONE_Q + ") OR id IN(" + FOLLOWERS_USER_BOTH_Q + ")) AND status='online'"
		} else {
			op.Where = "id IN(" + REQUEST_USER_Q + ")"
			op.Args = op.Args[1:]
		}
	} else if usersType == "following" {
		op.Where = "id IN(" + FOLLOWING_USER_ONE_Q + ") OR id IN(" + FOLLOWING_USER_BOTH_Q + ")"
	} else {
		op.Where = "id IN(" + FOLLOWERS_GROUP_ONE_Q + ") OR id IN (" + FOLLOWERS_GROUP_BOTH_Q + ")"
	}

	return generalGet(
		w,
		r,
		dbfuncs.SQLSelectParams{
			Table:   "Users",
			What:    "*",
			Options: op,
		},
		dbfuncs.User{},
	), nil
}

func (app *Application) Groups(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	ID, e := getUserID(w, r, r.FormValue("id"))
	if e != nil {
		return nil, errors.New("wrong id")
	}

	first, count := getLimit(r)
	groupType := r.FormValue("type")
	op := dbfuncs.DoSQLOption(
		"",
		"title ASC",
		"?,?",
		ID, first, count,
	)
	if groupType == "all" {
		op.Where = "id IN(" + FOLLOWERS_USER_GROUP_Q + ")"
	} else {
		op.Where = "id IN(" + REQUEST_GROUP_Q + ")"
	}

	return generalGet(
		w,
		r,
		dbfuncs.SQLSelectParams{
			Table:   "Groups",
			What:    "*",
			Options: op,
		},
		dbfuncs.Group{},
	), nil
}

func (app *Application) Events(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	ID, e := getUserID(w, r, r.FormValue("id"))
	if e != nil {
		return nil, errors.New("wrong id")
	}

	which := r.FormValue("which")
	if which != "user" && which != "group" {
		return nil, errors.New("wrong owner")
	}
	first, count := getLimit(r)

	return generalGet(
		w,
		r,
		dbfuncs.SQLSelectParams{
			Table: "Events",
			What:  "id, type, datetime",
			Options: dbfuncs.DoSQLOption(
				which+"ID = ?",
				"",
				"?,?",
				ID, first, count,
			),
		},
		struct {
			ID       int    `json:"id"`
			Type     string `json:"type"`
			Datetime int    `json:"datetime"`
		}{},
	), nil
}

func (app *Application) Messages(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	ID, e := getUserID(w, r, "")
	if e != nil {
		return nil, errors.New("wrong id")
	}

	first, count := getLimit(r)
	receiverID := r.FormValue("id")
	chatType := r.FormValue("type")

	op := dbfuncs.DoSQLOption(
		"senderUserID=? AND receiverGroupID=?",
		"datetime(datetime) DESC",
		"?,?",
		ID, receiverID,
	)
	if chatType == "user" {
		op.Where = "senderUserID=? AND receiverUserID=? OR senderUserID=? AND receiverUserID=?"
		op.Args = append(op.Args, receiverID, ID)
	}
	op.Args = append(op.Args, first, count)

	userJ := userJoin("m.senderUserID = u.id")
	fileJ := dbfuncs.DoSQLJoin(dbfuncs.LOJOINQ, "Files AS f", "f.messageID=m.id")
	return generalGet(
		w,
		r,
		dbfuncs.SQLSelectParams{
			Table:   "Messages AS m",
			What:    "m.*, u.ava, u.nName, u.status, f.src",
			Options: op,
			Joins:   []dbfuncs.SQLJoin{userJ, fileJ},
		},
		dbfuncs.Message{},
		"avatar", "nickname", "status", "src",
	), nil
}

func (app *Application) Chats(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	ID, e := getUserID(w, r, r.FormValue("id"))
	if e != nil {
		return nil, errors.New("wrong id")
	}

	first, count := getLimit(r)

	messageSelectOp := dbfuncs.DoSQLOption(
		`(c.senderUserID = m.senderUserID AND c.receiverUserID = m.receiverUserID) OR
		(c.senderUserID = m.receiverUserID AND c.receiverUserID = m.senderUserID) OR
		(c.senderUserID = m.senderUserID AND c.receiverGroupID = m.receiverGroupID)`,
		"m.datetime DESC",
		"?",
		1,
	)
	messageBodyQ := dbfuncs.SQLSelectParams{
		Table:   "Messages as m",
		What:    "m.body",
		Options: messageSelectOp,
	}
	messageDatetimeQ := dbfuncs.SQLSelectParams{
		Table:   "Messages as m",
		What:    "m.datetime",
		Options: messageSelectOp,
	}

	userJ := userJoin("c.receiverUserID=u.id")
	groupJ := groupJoin("c.receiverGroupID=g.id")
	mainQ := dbfuncs.SQLSelectParams{
		Table: "Chats c",
		What:  "c.*, u.ava, u.nName, u.status, g.ava, g.title",
		Options: dbfuncs.DoSQLOption(
			"(c.senderUserID = ? AND c.receiverUserID != ?) OR (c.senderUserID = ? AND c.receiverUserID ISNULL)",
			"",
			"?,?",
			ID, ID, ID, first, count,
		),
		Joins: []dbfuncs.SQLJoin{userJ, groupJ},
	}

	return dbfuncs.GetWithSubqueries(
		mainQ,
		[]dbfuncs.SQLSelectParams{messageBodyQ, messageDatetimeQ},
		[]string{"userAvatar", "nickname", "status", "groupAvatar", "groupTitle", "msgBody", "msgDatetime"},
		dbfuncs.Chat{},
	)
}

func (app *Application) Notification(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	userID := getUserIDfromReq(w, r)
	if userID == -1 {
		return nil, errors.New("not logged")
	}

	ID, e := strconv.Atoi(r.FormValue("id"))
	if e != nil {
		return nil, errors.New("wrong id")
	}

	noteType, e := strconv.Atoi(r.FormValue("type"))
	if e != nil {
		return nil, errors.New("wrong type")
	}

	selectType := func(noteType int) (string, string, string) {
		if noteType == 1 || noteType == 10 || noteType == 20 {
			return "Posts", "title", "postID"
		}
		if noteType == 2 {
			return "Events", "title", "eventID"
		}
		if noteType == 3 || noteType == 4 {
			return "Groups", "title", "groupID"
		}
		if noteType == 11 || noteType == 21 {
			return "Comments", "body", "commentID"
		}
		return "Media", "title", "mediaID"
	}
	getType, whatData, whatID := selectType(noteType)

	selectGetTypeQ := dbfuncs.SQLSelectParams{
		Table:   getType,
		What:    whatData,
		Options: dbfuncs.DoSQLOption("id = n."+whatID, "", ""),
	}
	userQ := dbfuncs.SQLSelectParams{
		Table:   "Users",
		What:    "nName",
		Options: dbfuncs.DoSQLOption("id = n.senderUserID", "", ""),
	}

	mainQ := dbfuncs.SQLSelectParams{
		Table:   "Notifications AS n",
		What:    "n.*",
		Options: dbfuncs.DoSQLOption("n.id=?", "", "", ID),
	}

	return dbfuncs.GetWithSubqueries(mainQ, []dbfuncs.SQLSelectParams{selectGetTypeQ, userQ}, []string{"whatData", "nickname"}, dbfuncs.Notification{})
}

func (app *Application) User(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	ID, e := strconv.Atoi(r.FormValue("id"))
	if e != nil {
		return nil, errors.New("wrong id")
	}

	mainQ := dbfuncs.SQLSelectParams{
		Table:   "Users",
		What:    "*",
		Options: dbfuncs.DoSQLOption("id=?", "", "", ID),
	}
	followersQ := dbfuncs.SQLSelectParams{
		Table: "Users",
		What:  "COUNT(id)",
		Options: dbfuncs.DoSQLOption(
			"id IN("+FOLLOWERS_USER_ONE_Q+") OR id IN("+FOLLOWERS_USER_BOTH_Q+")",
			"",
			"",
			ID, ID,
		),
	}
	followingQ := dbfuncs.SQLSelectParams{
		Table: "Users",
		What:  "COUNT(id)",
		Options: dbfuncs.DoSQLOption(
			"id IN("+FOLLOWING_USER_ONE_Q+") OR id IN("+FOLLOWING_USER_BOTH_Q+")",
			"",
			"",
			ID, ID,
		),
	}
	eventsQ := dbfuncs.SQLSelectParams{
		Table:   "Events",
		What:    "COUNT(id)",
		Options: dbfuncs.DoSQLOption("userID=?", "", "", ID),
	}
	groupsQ := dbfuncs.SQLSelectParams{
		Table:   "Relations",
		What:    "COUNT(id)",
		Options: dbfuncs.DoSQLOption("senderUserID=? AND receiverGroupID IS NOT NULL", "", "", ID),
	}
	mediaQ := dbfuncs.SQLSelectParams{
		Table:   "Media",
		What:    "COUNT(id)",
		Options: dbfuncs.DoSQLOption("userID=?", "", "", ID),
	}
	querys := []dbfuncs.SQLSelectParams{followersQ, followingQ, eventsQ, groupsQ, mediaQ}
	as := []string{"followersCount", "followingCount", "eventsCount", "groupsCount", "galleryCount"}

	if r.FormValue("type") == "profile" {
		userID := getUserIDfromReq(w, r)
		if userID == -1 {
			return nil, errors.New("not logged")
		}

		querys = append(querys, dbfuncs.SQLSelectParams{
			Table:   "Relations",
			What:    "value",
			Options: dbfuncs.DoSQLOption("senderUserID=? AND receiverUserID=?", "", "", ID, userID),
		}, dbfuncs.SQLSelectParams{
			Table:   "Relations",
			What:    "value",
			Options: dbfuncs.DoSQLOption("senderUserID=? AND receiverUserID=?", "", "", userID, ID),
		})
		as = append(as, "InRlshState", "OutRlshState")
	}

	return dbfuncs.GetWithSubqueries(
		mainQ,
		querys,
		as,
		dbfuncs.User{},
	)
}

func (app *Application) Group(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	ID, e := strconv.Atoi(r.FormValue("id"))
	if e != nil {
		return nil, errors.New("wrong id")
	}

	mainQ := dbfuncs.SQLSelectParams{
		Table:   "Groups",
		What:    "*",
		Options: dbfuncs.DoSQLOption("id=?", "", "", ID),
	}
	membersQ := dbfuncs.SQLSelectParams{
		Table: "Users",
		What:  "COUNT(id)",
		Options: dbfuncs.DoSQLOption(
			"id IN ("+FOLLOWERS_GROUP_ONE_Q+") OR id IN ("+FOLLOWERS_GROUP_BOTH_Q+")",
			"",
			"",
			ID, ID,
		),
	}
	eventsQ := dbfuncs.SQLSelectParams{
		Table:   "Events",
		What:    "COUNT(id)",
		Options: dbfuncs.DoSQLOption("groupID=?", "", "", ID),
	}
	mediaQ := dbfuncs.SQLSelectParams{
		Table:   "Media",
		What:    "COUNT(id)",
		Options: dbfuncs.DoSQLOption("groupID=?", "", "", ID),
	}
	querys := []dbfuncs.SQLSelectParams{membersQ, eventsQ, mediaQ}
	as := []string{"membersCount", "eventsCount", "galleryCount"}

	if r.FormValue("type") == "profile" {
		userID := getUserIDfromReq(w, r)
		if userID == -1 {
			return nil, errors.New("not logged")
		}

		querys = append(querys, dbfuncs.SQLSelectParams{
			Table:   "Relations",
			What:    "value",
			Options: dbfuncs.DoSQLOption("senderGroupID=? AND receiverUserID=?", "", "", ID, userID),
		}, dbfuncs.SQLSelectParams{
			Table:   "Relations",
			What:    "value",
			Options: dbfuncs.DoSQLOption("senderUserID=? AND receiverGroupID=?", "", "", userID, ID),
		})
		as = append(as, "InRlshState", "OutRlshState")
	}

	return dbfuncs.GetWithSubqueries(
		mainQ,
		querys,
		as,
		dbfuncs.Group{},
	)
}

func (app *Application) Post(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	ID, e := strconv.Atoi(r.FormValue("id"))
	if e != nil {
		return nil, errors.New("wrong id")
	}

	userID := getUserIDfromReq(w, r)
	if !isHaveAccessToPost(ID, userID) {
		return nil, errors.New("not have access")
	}

	carmaQ := carmaCountQ("postID", ID)
	userJoin := userJoin("u.id = p.userID")
	groupJoin := groupJoin("g.id = p.groupID")
	likesJoin := likeJoin("l.userID = p.userID AND l.postID = p.id")
	mainQ := dbfuncs.SQLSelectParams{
		Table:   "Posts as p",
		What:    "p.*, u.nName, u.ava, u.status, g.title, g.ava, l.id IS NOT NULL",
		Options: dbfuncs.DoSQLOption("p.id = ?", "", "", ID),
		Joins:   []dbfuncs.SQLJoin{userJoin, groupJoin, likesJoin},
	}

	return dbfuncs.GetWithSubqueries(
		mainQ,
		[]dbfuncs.SQLSelectParams{carmaQ},
		[]string{"nickname", "userAvatar", "status", "groupTitle", "groupAvatar", "isLiked", "carma"},
		dbfuncs.Post{},
	)
}

func (app *Application) Event(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	ID, e := strconv.Atoi(r.FormValue("id"))
	if e != nil {
		return nil, errors.New("wrong id")
	}

	userID := getUserIDfromReq(w, r)
	if !isHaveAccess(ID, userID, "Events") {
		return nil, errors.New("not have access")
	}

	userJoin := userJoin("u.id = e.userID")
	groupJoin := groupJoin("g.id = e.groupID")
	eventAnswersJoin := dbfuncs.DoSQLJoin(dbfuncs.LOJOINQ, "EventAnswers AS ea", "ea.userID = ? AND ea.eventID = e.id", userID)

	eventAnswersGoingQ := eventAnswerQ(0, ID)
	eventAnswersNotGoingQ := eventAnswerQ(1, ID)
	eventAnswersIDKQ := eventAnswerQ(2, ID)
	mainQ := dbfuncs.SQLSelectParams{
		Table:   "Events as e",
		What:    "e.*, u.nName, u.ava, u.status, g.title, g.ava, ea.answer",
		Options: dbfuncs.DoSQLOption("e.id = ?", "", "", ID),
		Joins:   []dbfuncs.SQLJoin{userJoin, groupJoin, eventAnswersJoin},
	}

	return dbfuncs.GetWithSubqueries(
		mainQ,
		[]dbfuncs.SQLSelectParams{eventAnswersGoingQ, eventAnswersNotGoingQ, eventAnswersIDKQ},
		[]string{"nickname", "userAvatar", "status", "groupTitle", "groupAvatar", "myVote", "votes0", "votes1", "votes2"},
		dbfuncs.Event{},
	)
}

// GetFile save one file
func (app *Application) GetFile(w http.ResponseWriter, r *http.Request) {
	file, content, e := dbfuncs.GetFileFromDrive(strings.Split(r.URL.Path, "/")[2])
	if e != nil {
		return
	}

	ftype := http.DetectContentType(content[:512])
	w.Header().Set("Content-Disposition", "attachment; filename="+file.Name)
	w.Header().Set("Content-Type", ftype)
	io.Copy(w, bytes.NewReader(content))
}

package app

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"wnet/pkg/orm"
)

type API_RESPONSE struct {
	Err  string      `json:"err"`
	Data interface{} `json:"data"`
}

const (
	FOLLOWING_USER_ONE_Q   string = "SELECT receiverUserID FROM Relations WHERE senderUserID = ? AND (value = 1 OR value = 0)"
	FOLLOWING_USER_BOTH_Q  string = "SELECT senderUserID FROM Relations WHERE receiverUserID = ? AND value = 0"
	FOLLOWING_USER_GROUP_Q string = "SELECT receiverGroupID FROM Relations WHERE senderUserID = ? AND value = 1"
	FOLLOWERS_USER_ONE_Q   string = "SELECT senderUserID FROM Relations WHERE receiverUserID = ? AND (value = 1 OR value = 0)"
	FOLLOWERS_USER_BOTH_Q  string = "SELECT receiverUserID FROM Relations WHERE senderUserID = ? AND value = 0"
	FOLLOWERS_USER_GROUP_Q string = "SELECT receiverGroupID FROM Relations WHERE senderUserID = ? AND value = 1"
	FOLLOWERS_GROUP_ONE_Q  string = "SELECT senderUserID FROM Relations WHERE receiverGroupID = ? AND (value = 1 OR value = 0)"
	FOLLOWERS_GROUP_BOTH_Q string = "SELECT receiverUserID FROM Relations WHERE senderGroupID = ? AND value = 0"
	REQUEST_USER_USER_Q    string = "SELECT senderUserID FROM Relations WHERE receiverUserID = ? AND value=-1"
	REQUEST_USER_GROUP_Q   string = "SELECT senderGroupID FROM Relations WHERE receiverUserID = ? AND value=-1"
	REQUEST_GROUP_USER_Q   string = "SELECT senderUserID FROM Relations WHERE receiverGroupID = ? AND value=-1"
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

func generalGet(w http.ResponseWriter, r *http.Request, selectParams orm.SQLSelectParams, sampleStruct interface{}, additionalFields ...string) []map[string]interface{} {
	results, e := orm.GetFrom(selectParams)
	if e != nil || len(results) == 0 {
		return []map[string]interface{}{}
	}
	return orm.MapFromStructAndMatrix(results, sampleStruct, additionalFields...)
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

	res, e := orm.GetWithQueryAndArgs(q, args)
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

	res, e := orm.GetWithQueryAndArgs(q, args)
	if e != nil || res == nil || (res != nil && (res[0][0] == 0 || res[0][0] == nil)) {
		return false
	}
	return true
}

func userJoin(joinCondition string) orm.SQLJoin {
	return orm.DoSQLJoin(orm.LOJOINQ, "Users AS u", joinCondition)
}

func groupJoin(joinCondition string) orm.SQLJoin {
	return orm.DoSQLJoin(orm.LOJOINQ, "Groups AS g", joinCondition)
}

func likeJoin(joinCondition string) orm.SQLJoin {
	return orm.DoSQLJoin(orm.LOJOINQ, "Likes AS l", joinCondition)
}

func carmaCountQ(column string, columnID int) orm.SQLSelectParams {
	return orm.SQLSelectParams{
		Table:   "Likes",
		What:    "COUNT(id)",
		Options: orm.DoSQLOption(column+" = ?", "", "", columnID),
	}
}

func eventAnswerQ(answer, eventID int) orm.SQLSelectParams {
	return orm.SQLSelectParams{
		Table:   "EventAnswers",
		What:    "COUNT(id)",
		Options: orm.DoSQLOption("answer = ? AND eventID = ?", "", "", answer, eventID),
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
	newsIDS, e := orm.GetWithQueryAndArgs(
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

	return orm.MapFromStructAndMatrix(newsIDS, struct {
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
	publIDS, e := orm.GetWithQueryAndArgs(q, args)
	if e != nil {
		return nil, errors.New("wrong data")
	}

	return orm.MapFromStructAndMatrix(publIDS, struct {
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
		orm.SQLSelectParams{
			Table: "Notifications",
			What:  "id, type",
			Options: orm.DoSQLOption(
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

	op := orm.DoSQLOption("", "m.datetime DESC", "?,?")
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

	getJoin := func(isUser bool) (orm.SQLJoin, string) {
		if isUser {
			return userJoin("m.userID=u.id"), "u.nName, u.ava"
		}
		return groupJoin("m.groupID=g.id"), "g.title, g.ava"
	}
	join, getStr := getJoin(isUser)

	return generalGet(
		w,
		r,
		orm.SQLSelectParams{
			Table:   "Media as m",
			What:    "m.*, " + getStr,
			Options: op,
			Joins:   []orm.SQLJoin{join},
		},
		orm.Media{},
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
	op := orm.DoSQLOption(
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
			if flwType == "request_g" {
				op.Where = "id IN(" + REQUEST_GROUP_USER_Q + ")"
			} else {
				op.Where = "id IN(" + REQUEST_USER_USER_Q + ")"
			}
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
		orm.SQLSelectParams{
			Table:   "Users",
			What:    "*",
			Options: op,
		},
		orm.User{},
	), nil
}

func (app *Application) Groups(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	ID, e := getUserID(w, r, r.FormValue("id"))
	if e != nil {
		return nil, errors.New("wrong id")
	}

	first, count := getLimit(r)
	groupType := r.FormValue("type")
	op := orm.DoSQLOption(
		"",
		"title ASC",
		"?,?",
		ID, first, count,
	)
	if groupType == "all" {
		op.Where = "id IN(" + FOLLOWERS_USER_GROUP_Q + ")"
	} else {
		op.Where = "id IN(" + REQUEST_USER_GROUP_Q + ")"
	}

	return generalGet(
		w,
		r,
		orm.SQLSelectParams{
			Table:   "Groups",
			What:    "*",
			Options: op,
		},
		orm.Group{},
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
		orm.SQLSelectParams{
			Table: "Events",
			What:  "id, type, datetime",
			Options: orm.DoSQLOption(
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

func (app *Application) Comments(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	ID, e := strconv.Atoi(r.FormValue("id"))
	if e != nil {
		return nil, errors.New("wrong id")
	}

	userID := getUserIDfromReq(w, r)
	if userID == -1 {
		return nil, errors.New("not logged")
	}

	commentType := r.FormValue("type")
	if commentType != "post" && commentType != "comment" && commentType != "media" {
		return nil, errors.New("wrong comment type")
	}

	op := orm.DoSQLOption("c.id = ?", "", "", ID)
	if r.FormValue("count") != "single" {
		first, count := getLimit(r)
		op = orm.DoSQLOption("c."+commentType+"ID = ?", "c.datetime DESC", "?,?", ID, first, count)
	}

	carmaQ := carmaCountQ("commentID", ID)
	userJoin := userJoin("u.id = c.userID")
	likesJoin := likeJoin("l.userID = c.userID AND l.commentID = c.id")
	mainQ := orm.SQLSelectParams{
		Table:   "Comments as c",
		What:    "c.*, u.nName, u.ava, u.status, l.id IS NOT NULL",
		Options: op,
		Joins:   []orm.SQLJoin{userJoin, likesJoin},
	}

	return orm.GetWithSubqueries(
		mainQ,
		[]orm.SQLSelectParams{carmaQ},
		[]string{"nickname", "avatar", "status", "isLiked"},
		[]string{"carma"},
		orm.Comment{},
	)
}

func (app *Application) Messages(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	ID, e := getUserID(w, r, "")
	if e != nil {
		return nil, errors.New("wrong id")
	}

	first, count := getLimit(r)
	receiverID := r.FormValue("id")
	chatType := r.FormValue("type")

	op := orm.DoSQLOption(
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
	fileJ := orm.DoSQLJoin(orm.LOJOINQ, "Files AS f", "f.messageID=m.id")
	return generalGet(
		w,
		r,
		orm.SQLSelectParams{
			Table:   "Messages AS m",
			What:    "m.*, u.ava, u.nName, u.status, f.src",
			Options: op,
			Joins:   []orm.SQLJoin{userJ, fileJ},
		},
		orm.Message{},
		"avatar", "nickname", "status", "src",
	), nil
}

func (app *Application) Chats(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	ID, e := getUserID(w, r, r.FormValue("id"))
	if e != nil {
		return nil, errors.New("wrong id")
	}

	first, count := getLimit(r)
	messageSelectOp := orm.DoSQLOption(
		`(c.senderUserID = m.senderUserID AND c.receiverUserID = m.receiverUserID) OR
		(c.senderUserID = m.receiverUserID AND c.receiverUserID = m.senderUserID) OR
		(c.senderUserID = m.senderUserID AND c.receiverGroupID = m.receiverGroupID)`,
		"m.datetime DESC",
		"?",
		1,
	)
	messageBodyQ := orm.SQLSelectParams{
		Table:   "Messages AS m",
		What:    "m.body",
		Options: messageSelectOp,
	}
	messageDatetimeQ := orm.SQLSelectParams{
		Table:   "Messages AS m",
		What:    "m.datetime",
		Options: messageSelectOp,
	}

	userJ := userJoin("c.receiverUserID=u.id")
	groupJ := groupJoin("c.receiverGroupID=g.id")
	mainQ := orm.SQLSelectParams{
		Table: "Chats c",
		What:  "c.*, u.ava, u.nName, u.status, g.ava, g.title",
		Options: orm.DoSQLOption(
			"(c.users LIKE '%|"+strconv.Itoa(ID)+" %' AND c.closed NOT LIKE '%|"+strconv.Itoa(ID)+" %') ",
			"",
			"?,?",
			first, count,
		),
		Joins: []orm.SQLJoin{userJ, groupJ},
	}

	return orm.GetWithSubqueries(
		mainQ,
		[]orm.SQLSelectParams{messageBodyQ, messageDatetimeQ},
		[]string{"userAvatar", "nickname", "status", "groupAvatar", "groupTitle"},
		[]string{"msgBody", "msgDatetime"},
		orm.Chat{},
	)
}

func (app *Application) ClippedFiles(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	ID, e := strconv.Atoi(r.FormValue("id"))
	if e != nil {
		return nil, errors.New("wrong id")
	}

	clippedType := r.FormValue("type")
	if clippedType != "comment" && clippedType != "post" && clippedType != "message" {
		return nil, errors.New("wrong type")
	}

	return generalGet(
		w,
		r,
		orm.SQLSelectParams{
			Table:   "Files",
			What:    "*",
			Options: orm.DoSQLOption(clippedType+"ID=?", "", "", ID),
		},
		orm.ClippedFile{},
	), nil
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
		if noteType == 5 {
			return "", "", ""
		}
		if noteType == 1 || noteType == 10 || noteType == 20 {
			return "Posts", "title", "postID"
		}
		if noteType == 2 {
			return "Events", "title", "eventID"
		}
		if noteType == 3 || noteType == 4 || noteType == 6 {
			return "Groups", "title", "groupID"
		}
		if noteType == 11 || noteType == 21 {
			return "Comments", "body", "commentID"
		}
		return "Media", "title", "mediaID"
	}
	getType, whatData, whatID := selectType(noteType)

	querys := []orm.SQLSelectParams{{
		Table:   "Users",
		What:    "nName",
		Options: orm.DoSQLOption("id = n.senderUserID", "", ""),
	}}
	qAs := []string{"nickname"}

	if noteType != 5 {
		querys = append(querys, orm.SQLSelectParams{
			Table:   getType,
			What:    whatData,
			Options: orm.DoSQLOption("id = n."+whatID, "", ""),
		})
		qAs = append(qAs, "whatData")
	}

	mainQ := orm.SQLSelectParams{
		Table:   "Notifications AS n",
		What:    "n.*",
		Options: orm.DoSQLOption("n.id=?", "", "", ID),
	}

	return orm.GetWithSubqueries(mainQ, querys, []string{}, qAs, orm.Notification{})
}

func (app *Application) User(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	ID, e := strconv.Atoi(r.FormValue("id"))
	if e != nil {
		return nil, errors.New("wrong id")
	}

	mainQ := orm.SQLSelectParams{
		Table:   "Users",
		What:    "*",
		Options: orm.DoSQLOption("id=?", "", "", ID),
	}
	followersQ := orm.SQLSelectParams{
		Table: "Users",
		What:  "COUNT(id)",
		Options: orm.DoSQLOption(
			"id IN("+FOLLOWERS_USER_ONE_Q+") OR id IN("+FOLLOWERS_USER_BOTH_Q+")",
			"",
			"",
			ID, ID,
		),
	}
	followingQ := orm.SQLSelectParams{
		Table: "Users",
		What:  "COUNT(id)",
		Options: orm.DoSQLOption(
			"id IN("+FOLLOWING_USER_ONE_Q+") OR id IN("+FOLLOWING_USER_BOTH_Q+")",
			"",
			"",
			ID, ID,
		),
	}
	eventsQ := orm.SQLSelectParams{
		Table:   "Events",
		What:    "COUNT(id)",
		Options: orm.DoSQLOption("userID=?", "", "", ID),
	}
	groupsQ := orm.SQLSelectParams{
		Table:   "Relations",
		What:    "COUNT(id)",
		Options: orm.DoSQLOption("senderUserID=? AND receiverGroupID IS NOT NULL", "", "", ID),
	}
	mediaQ := orm.SQLSelectParams{
		Table:   "Media",
		What:    "COUNT(id)",
		Options: orm.DoSQLOption("userID=?", "", "", ID),
	}
	querys := []orm.SQLSelectParams{followersQ, followingQ, eventsQ, groupsQ, mediaQ}
	as := []string{"followersCount", "followingCount", "eventsCount", "groupsCount", "galleryCount"}

	if r.FormValue("type") == "profile" {
		userID := getUserIDfromReq(w, r)
		if userID == -1 {
			return nil, errors.New("not logged")
		}

		querys = append(querys, orm.SQLSelectParams{
			Table:   "Relations",
			What:    "value",
			Options: orm.DoSQLOption("senderUserID=? AND receiverUserID=?", "", "", ID, userID),
		}, orm.SQLSelectParams{
			Table:   "Relations",
			What:    "value",
			Options: orm.DoSQLOption("senderUserID=? AND receiverUserID=?", "", "", userID, ID),
		})
		as = append(as, "InRlshState", "OutRlshState")
	}

	return orm.GetWithSubqueries(
		mainQ,
		querys,
		[]string{},
		as,
		orm.User{},
	)
}

func (app *Application) Group(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	ID, e := strconv.Atoi(r.FormValue("id"))
	if e != nil {
		return nil, errors.New("wrong id")
	}

	mainQ := orm.SQLSelectParams{
		Table:   "Groups",
		What:    "*",
		Options: orm.DoSQLOption("id=?", "", "", ID),
	}
	membersQ := orm.SQLSelectParams{
		Table: "Users",
		What:  "COUNT(id)",
		Options: orm.DoSQLOption(
			"id IN ("+FOLLOWERS_GROUP_ONE_Q+") OR id IN ("+FOLLOWERS_GROUP_BOTH_Q+")",
			"",
			"",
			ID, ID,
		),
	}
	requestQ := orm.SQLSelectParams{
		Table: "Users",
		What:  "COUNT(id)",
		Options: orm.DoSQLOption(
			"id IN ("+REQUEST_GROUP_USER_Q+")",
			"",
			"",
			ID, ID,
		),
	}
	eventsQ := orm.SQLSelectParams{
		Table:   "Events",
		What:    "COUNT(id)",
		Options: orm.DoSQLOption("groupID=?", "", "", ID),
	}
	mediaQ := orm.SQLSelectParams{
		Table:   "Media",
		What:    "COUNT(id)",
		Options: orm.DoSQLOption("groupID=?", "", "", ID),
	}
	querys := []orm.SQLSelectParams{membersQ, eventsQ, mediaQ, requestQ}
	as := []string{"membersCount", "eventsCount", "galleryCount", "requestsCount"}

	if r.FormValue("type") == "profile" {
		userID := getUserIDfromReq(w, r)
		if userID == -1 {
			return nil, errors.New("not logged")
		}

		querys = append(querys, orm.SQLSelectParams{
			Table:   "Relations",
			What:    "value",
			Options: orm.DoSQLOption("senderGroupID=? AND receiverUserID=?", "", "", ID, userID),
		}, orm.SQLSelectParams{
			Table:   "Relations",
			What:    "value",
			Options: orm.DoSQLOption("senderUserID=? AND receiverGroupID=?", "", "", userID, ID),
		})
		as = append(as, "InRlshState", "OutRlshState")
	}

	return orm.GetWithSubqueries(
		mainQ,
		querys,
		[]string{},
		as,
		orm.Group{},
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
	mainQ := orm.SQLSelectParams{
		Table:   "Posts as p",
		What:    "p.*, u.nName, u.ava, u.status, g.title, g.ava, l.id IS NOT NULL",
		Options: orm.DoSQLOption("p.id = ?", "", "", ID),
		Joins:   []orm.SQLJoin{userJoin, groupJoin, likesJoin},
	}

	return orm.GetWithSubqueries(
		mainQ,
		[]orm.SQLSelectParams{carmaQ},
		[]string{"nickname", "userAvatar", "status", "groupTitle", "groupAvatar", "isLiked"},
		[]string{"carma"},
		orm.Post{},
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
	eventAnswersJoin := orm.DoSQLJoin(orm.LOJOINQ, "EventAnswers AS ea", "ea.userID = ? AND ea.eventID = e.id", userID)

	eventAnswersGoingQ := eventAnswerQ(0, ID)
	eventAnswersNotGoingQ := eventAnswerQ(1, ID)
	eventAnswersIDKQ := eventAnswerQ(2, ID)
	mainQ := orm.SQLSelectParams{
		Table:   "Events as e",
		What:    "e.*, u.nName, u.ava, u.status, g.title, g.ava, ea.answer",
		Options: orm.DoSQLOption("e.id = ?", "", "", ID),
		Joins:   []orm.SQLJoin{userJoin, groupJoin, eventAnswersJoin},
	}

	return orm.GetWithSubqueries(
		mainQ,
		[]orm.SQLSelectParams{eventAnswersGoingQ, eventAnswersNotGoingQ, eventAnswersIDKQ},
		[]string{"nickname", "userAvatar", "status", "groupTitle", "groupAvatar", "myVote"},
		[]string{"votes0", "votes1", "votes2"},
		orm.Event{},
	)
}

func (app *Application) Media(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	ID, e := strconv.Atoi(r.FormValue("id"))
	if e != nil {
		return nil, errors.New("wrong id")
	}
	userID := getUserIDfromReq(w, r)
	if !isHaveAccess(ID, userID, "Media") {
		return nil, errors.New("not have access")
	}
	mediaType := r.FormValue("type")
	if mediaType != "photo" && mediaType != "video" {
		return nil, errors.New("wrong type")
	}

	carmaQ := carmaCountQ("mediaID", ID)
	userJoin := userJoin("u.id = m.userID")
	groupJoin := groupJoin("g.id = m.groupID")
	likesJoin := likeJoin("l.userID = m.userID AND l.mediaID = m.id")
	mainQ := orm.SQLSelectParams{
		Table:   "Media as m",
		What:    "m.*, u.nName, u.ava, u.status, g.title, g.ava, l.id IS NOT NULL",
		Options: orm.DoSQLOption("m.id = ? AND m.type = ?", "", "", ID, mediaType),
		Joins:   []orm.SQLJoin{userJoin, groupJoin, likesJoin},
	}

	return orm.GetWithSubqueries(
		mainQ,
		[]orm.SQLSelectParams{carmaQ},
		[]string{"nickname", "userAvatar", "status", "groupTitle", "groupAvatar", "isLiked"},
		[]string{"carma"},
		orm.Media{},
	)
}

// GetFile save one file
func (app *Application) GetFile(w http.ResponseWriter, r *http.Request) {
	file, content, e := orm.GetFileFromDrive(strings.Split(r.URL.Path, "/")[2])
	if e != nil {
		return
	}

	ftype := http.DetectContentType(content[:512])
	w.Header().Set("Content-Disposition", "attachment; filename="+file.Name)
	w.Header().Set("Content-Type", ftype)
	io.Copy(w, bytes.NewReader(content))
}

func searchGetCountFilter(where, formVal string, defVal int, op *orm.SQLOption) {
	if formVal != "" {
		val, e := strconv.Atoi(formVal)
		if e != nil {
			val = defVal
		}
		op.Where += where + " ? AND"
		op.Args = append(op.Args, val)
	}
}

func searchGetBtnFilter(val string, choises [][]string, isOrder bool, op *orm.SQLOption) {
	for _, choise := range choises {
		if val == choise[0] {
			if isOrder {
				op.Order = choise[1] + " DESC"
				return
			}
			op.Where += choise[1] + " ? AND"
			op.Args = append(op.Args, choise[0])
			return
		}
	}
}

func searchGetSwitchFilter(where, formVal, onVal, offVal string, op *orm.SQLOption) {
	if formVal == "1" {
		op.Where += where + " ? AND"
		if offVal == "" {
			op.Args = append(op.Args, onVal)
		} else {
			op.Args = append(op.Args, offVal)
		}
		return
	}
}

func removeLastFromStr(src, delim string) string {
	splitted := strings.Split(src, delim)
	return strings.Join(splitted[:len(splitted)-1], delim)
}

func searchGetTextFilter(q string, searchFields []string, op *orm.SQLOption) error {
	if q != "" {
		if XCSSOther(q) != nil {
			return errors.New("danger search text")
		}

		op.Where += "("
		for _, v := range searchFields {
			op.Where += v + " LIKE '%" + q + "%' OR "
		}
		op.Where = removeLastFromStr(op.Where, "OR ")
		op.Where += ")"
		return nil
	}
	op.Where = removeLastFromStr(op.Where, "AND")
	return nil
}

func doSearch(r *http.Request, q orm.SQLSelectParams, sampleStruct interface{}, addQs []orm.SQLSelectParams, joinAs, qAs []string) (interface{}, error) {
	first, count := getLimit(r)
	q.Options.Args = append(q.Options.Args, first, count)
	return orm.GetWithSubqueries(q, addQs, joinAs, qAs, sampleStruct)
}

func searchUser(userID int, r *http.Request) (interface{}, error) {
	op := orm.DoSQLOption("", "status DESC", "?,?")

	searchGetCountFilter(" age >=", r.FormValue("agemin"), 0, &op)
	searchGetCountFilter(" age <=", r.FormValue("agemax"), 100, &op)
	searchGetCountFilter(" followers >=", r.FormValue("subsmin"), 0, &op)
	searchGetCountFilter(" followers <=", r.FormValue("subsmax"), 0, &op)

	searchGetBtnFilter(
		r.FormValue("sort"),
		[][]string{
			{"subs", "followers"},
		},
		true,
		&op,
	)
	searchGetBtnFilter(
		r.FormValue("gender"),
		[][]string{
			{"Male", " gender ="},
			{"Female", " gender ="},
		},
		false,
		&op,
	)

	searchGetSwitchFilter(" status =", r.FormValue("online"), "online", "", &op)

	if e := searchGetTextFilter(r.FormValue("q"), []string{"u.fName", "u.lName", "u.nName"}, &op); e != nil {
		return nil, e
	}

	subsCountQ := orm.SQLSelectParams{
		Table:   "Relations AS r",
		What:    "COUNT(id)",
		Options: orm.DoSQLOption("receiverUserID = u.id", "", ""),
	}
	inRlsh := orm.SQLSelectParams{
		Table:   "Relations AS inR",
		What:    "value",
		Options: orm.DoSQLOption("senderUserID = u.id AND receiverUserID = ?", "", "", userID),
	}
	outRlsh := orm.SQLSelectParams{
		Table:   "Relations AS outR",
		What:    "value",
		Options: orm.DoSQLOption("senderUserID = ? AND receiverUserID = u.id", "", "", userID),
	}
	q := orm.SQLSelectParams{
		Table:   "Users AS u",
		What:    "u.*",
		Options: op,
	}
	return doSearch(r, q, orm.User{}, []orm.SQLSelectParams{subsCountQ, inRlsh, outRlsh}, []string{}, []string{"followers", "InRlshState", "OutRlshState"})
}

func searchGroup(userID int, r *http.Request) (interface{}, error) {
	op := orm.DoSQLOption("", "cdate DESC", "?,?")

	searchGetCountFilter(" members >=", r.FormValue("membmin"), 0, &op)
	searchGetCountFilter(" members <=", r.FormValue("membmax"), 100, &op)

	searchGetBtnFilter(
		r.FormValue("sort"),
		[][]string{
			{"member", "members"},
		},
		true,
		&op,
	)

	searchGetSwitchFilter("isPrivate =", r.FormValue("private"), "1", "0", &op)

	if e := searchGetTextFilter(r.FormValue("q"), []string{"g.title", "g.about"}, &op); e != nil {
		return nil, e
	}

	subsCountQ := orm.SQLSelectParams{
		Table:   "Relations AS r",
		What:    "COUNT(id)",
		Options: orm.DoSQLOption("receiverGroupID = g.id", "", ""),
	}
	inRlsh := orm.SQLSelectParams{
		Table:   "Relations AS inR",
		What:    "value",
		Options: orm.DoSQLOption("senderGroupID = g.id AND receiverUserID = ?", "", "", userID),
	}
	outRlsh := orm.SQLSelectParams{
		Table:   "Relations AS outR",
		What:    "value",
		Options: orm.DoSQLOption("senderUserID = ? AND receiverGroupID = g.id", "", "", userID),
	}
	q := orm.SQLSelectParams{
		Table:   "Groups AS g",
		What:    "g.*",
		Options: op,
	}
	return doSearch(r, q, orm.Group{}, []orm.SQLSelectParams{subsCountQ, inRlsh, outRlsh}, []string{}, []string{"members", "InRlshState", "OutRlshState"})
}

func searchPost(userID int, r *http.Request) (interface{}, error) {
	op := orm.DoSQLOption("", "datetime DESC", "?,?")

	searchGetCountFilter(" carma >=", r.FormValue("carmamin"), 0, &op)
	searchGetCountFilter(" carma <=", r.FormValue("carmamax"), 100, &op)

	searchGetBtnFilter(
		r.FormValue("sort"),
		[][]string{
			{"pop", "carma"},
		},
		true,
		&op,
	)

	if e := searchGetTextFilter(r.FormValue("q"), []string{"p.title", "p.body"}, &op); e != nil {
		return nil, e
	}

	userJoin := userJoin("u.id = p.userID")
	groupJoin := groupJoin("g.id = p.groupID")
	likesJoin := likeJoin("l.userID = ? AND l.postID = p.id")
	likesJoin.Args = append(likesJoin.Args, userID)
	carmaQ := orm.SQLSelectParams{
		Table:   "Likes AS c",
		What:    "COUNT(id)",
		Options: orm.DoSQLOption("c.postID = p.id", "", ""),
	}
	q := orm.SQLSelectParams{
		Table:   "Posts AS p",
		What:    "p.*, u.nName, u.ava, u.status, g.title, g.ava, l.id IS NOT NULL",
		Options: op,
		Joins:   []orm.SQLJoin{userJoin, groupJoin, likesJoin},
	}
	return doSearch(
		r,
		q,
		orm.Post{},
		[]orm.SQLSelectParams{carmaQ},
		[]string{"nickname", "userAvatar", "status", "groupTitle", "groupAvatar", "isLiked"},
		[]string{"carma"},
	)
}

func searchVideo(userID int, r *http.Request) (interface{}, error) {
	op := orm.DoSQLOption("m.type = 'video' AND ", "datetime DESC", "?,?")

	searchGetCountFilter(" carma >=", r.FormValue("carmamin"), 0, &op)
	searchGetCountFilter(" carma <=", r.FormValue("carmamax"), 100, &op)

	searchGetBtnFilter(
		r.FormValue("sort"),
		[][]string{
			{"pop", "carma"},
		},
		true,
		&op,
	)

	if e := searchGetTextFilter(r.FormValue("q"), []string{"m.title"}, &op); e != nil {
		return nil, e
	}

	userJoin := userJoin("u.id = m.userID")
	groupJoin := groupJoin("g.id = m.groupID")
	likesJoin := likeJoin("l.userID = ? AND l.mediaID = m.id")
	likesJoin.Args = append(likesJoin.Args, userID)
	carmaQ := orm.SQLSelectParams{
		Table:   "Likes AS c",
		What:    "COUNT(id)",
		Options: orm.DoSQLOption("c.mediaID = m.id", "", ""),
	}
	q := orm.SQLSelectParams{
		Table:   "Media AS m",
		What:    "m.*, u.nName, u.ava, u.status, g.title, g.ava, l.id IS NOT NULL",
		Options: op,
		Joins:   []orm.SQLJoin{userJoin, groupJoin, likesJoin},
	}
	return doSearch(
		r,
		q,
		orm.Media{},
		[]orm.SQLSelectParams{carmaQ},
		[]string{"nickname", "userAvatar", "status", "groupTitle", "groupAvatar", "isLiked"},
		[]string{"carma"},
	)
}

func (app *Application) Search(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	userID := getUserIDfromReq(w, r)
	if userID == -1 {
		return nil, errors.New("not logged")
	}

	searchType := r.FormValue("type")
	if searchType != "all" &&
		searchType != "user" &&
		searchType != "group" &&
		searchType != "post" &&
		searchType != "video" {
		return nil, errors.New("wrong type")
	}

	if searchType == "user" {
		return searchUser(userID, r)
	}
	if searchType == "group" {
		return searchGroup(userID, r)
	}
	if searchType == "post" {
		return searchPost(userID, r)
	}
	if searchType == "video" {
		return searchVideo(userID, r)
	}
	return nil, errors.New("'All' not supported yet")
}

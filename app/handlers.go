package app

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"
	"wnet/app/dbfuncs"
)

type API_RESPONSE struct {
	Err  string      `json:"err"`
	Data interface{} `json:"data"`
}

// SecureHeaderMiddleware set secure header option
func (app *Application) SecureHeaderMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("cross-origin-resource-policy", "cross-origin")
		w.Header().Set("X-XSS-Protection", "1;mode=block")
		w.Header().Set("X-Frame-Options", "deny")
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		next.ServeHTTP(w, r)
	})
}

// AccessLogMiddleware logging request
func (app *Application) AccessLogMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if app.CurrentRequestCount < app.MaxRequestCount {
			app.CurrentRequestCount++
			app.ILog.Printf(logingReq(r))
			next.ServeHTTP(w, r)
		} else {
			app.eHandler(w, errors.New("rate < curl"), "service is overloaded", 529)
		}
	})
}

/* ----------------------------------------------- Websocket ---------------------------------------------- */

// CreateWSUser create one WSUser
func (app *Application) CreateWSUser(w http.ResponseWriter, r *http.Request) {
	conn, e := upgrader.Upgrade(w, r, nil)
	if e != nil {
		app.ELog.Println(e, r.RemoteAddr)
		return
	}

	userID := getUserIDfromReq(w, r)
	if userID == -1 {
		app.ELog.Println(e, r.RemoteAddr)
		return
	}

	user := &WSUser{Conn: conn, ID: userID}
	app.m.Lock()
	app.OnlineUsers[user.ID] = user
	app.m.Unlock()

	go app.OnlineUsers[user.ID].HandleUserMsg(app)
	go app.OnlineUsers[user.ID].Pinger()
}

/* ------------------------------------------- API ------------------------------------------------ */

// HNews for handle '/api/news'
func (app *Application) HNews(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		data := API_RESPONSE{
			Err:  "ok",
			Data: "",
		}

		userID := getUserIDfromReq(w, r)
		if userID == -1 {
			data.Err = "Not logged"
			doJS(w, data)
			return
		}

		first, count := getLimit(r)

		followingS := `(SELECT receiverUserID FROM Relations WHERE senderUserID = ?),
			(SELECT senderUserID FROM Relations WHERE receiverUserID = ? AND value = 0),
			(SELECT receiverGroupID FROM Relations WHERE senderUserID = ?)`

		// res ex: [[4 post 1588670115000] [3 event 1588670115000]]
		newsIDS, e := dbfuncs.GetWithQueryAndArgs(
			`SELECT id, type, datetime FROM Posts WHERE userID IN(`+followingS+`)
			UNION ALL
			SELECT id, type, datetime FROM Events WHERE userID IN(`+followingS+`) ORDER BY datetime DESC LIMIT ?,?`,
			[]interface{}{userID, userID, userID, userID, userID, userID, first, count},
		)
		if e != nil {
			data.Err = "wrong data"
			doJS(w, data)
			return
		}

		data.Data = dbfuncs.MapFromStructAndMatrix(newsIDS, struct {
			ID       int    `json:"id"`
			Type     string `json:"type"`
			Datetime int    `json:"datetime"`
		}{})
		doJS(w, data)
	}
}

// HNotifications for handle '/api/notifications'
func (app *Application) HNotifications(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		data := API_RESPONSE{
			Err:  "ok",
			Data: "",
		}

		userID := getUserIDfromReq(w, r)
		if userID == -1 {
			data.Err = "Not logged"
			doJS(w, data)
			return
		}

		first, count := getLimit(r)
		followingS := `(SELECT receiverUserID FROM Relations WHERE senderUserID = ?),
			(SELECT senderUserID FROM Relations WHERE receiverUserID = ? AND value = 0),
			(SELECT receiverGroupID FROM Relations WHERE senderUserID = ?)`

		data.Data = generalGet(
			w,
			r,
			dbfuncs.SQLSelectParams{
				Table: "Notifications",
				What:  "id, type",
				Options: dbfuncs.DoSQLOption(
					"receiverUserID=? OR senderUserID IN("+followingS+")",
					"datetime DESC",
					"?,?",
					userID, userID, userID, userID, first, count,
				),
			},
			struct {
				ID   int `json:"id"`
				Type int `json:"type"`
			}{},
		)
		doJS(w, data)
	}
}

// HNotification for handle '/api/notification'
func (app *Application) HNotification(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		data := API_RESPONSE{
			Err:  "ok",
			Data: "",
		}

		userID := getUserIDfromReq(w, r)
		if userID == -1 {
			data.Err = "Not logged"
			doJS(w, data)
			return
		}

		ID, e := strconv.Atoi(r.FormValue("id"))
		if e != nil {
			data.Err = "wrong id"
			doJS(w, data)
			return
		}

		noteType, e := strconv.Atoi(r.FormValue("type"))
		if e != nil {
			data.Err = "wrong type"
			doJS(w, data)
			return
		}

		getType := ""
		whatData := "title"
		whatID := ""
		if noteType == 1 || noteType == 10 || noteType == 20 {
			getType = "Posts"
			whatID = "postID"
		} else if noteType == 2 {
			getType = "Events"
			whatID = "eventID"
		} else if noteType == 3 || noteType == 4 {
			getType = "Groups"
			whatID = "groupID"
		} else if noteType == 11 || noteType == 21 {
			getType = "Comments"
			whatData = "body"
			whatID = "commentID"
		} else {
			getType = "Media"
			whatID = "mediaID"
		}

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
			Table:   "Notifications as n",
			What:    "n.*",
			Options: dbfuncs.DoSQLOption("n.id=?", "", "", ID),
		}

		datas, e := dbfuncs.GetWithSubqueries(mainQ, []dbfuncs.SQLSelectParams{selectGetTypeQ, userQ}, []string{"whatData", "nickname"}, dbfuncs.Notifications{})
		if e != nil {
			data.Err = e.Error()
		}
		data.Data = datas
		doJS(w, data)

		fmt.Println("note", datas)
	}
}

// HUser for handle '/api/user/'
func (app *Application) HUser(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		data := API_RESPONSE{
			Err:  "ok",
			Data: "",
		}

		ID, e := strconv.Atoi(r.FormValue("id"))
		if e != nil {
			data.Err = "wrong id"
			doJS(w, data)
			return
		}

		mainQ := dbfuncs.SQLSelectParams{
			Table:   "Users",
			What:    "*",
			Options: dbfuncs.DoSQLOption("id=?", "", "", ID),
		}
		followersQ := dbfuncs.SQLSelectParams{
			Table:   "Relations",
			What:    "COUNT(id) as followersCount",
			Options: dbfuncs.DoSQLOption("receiverUserID=?", "", "", ID),
		}
		followingQ := dbfuncs.SQLSelectParams{
			Table:   "Relations",
			What:    "COUNT(id) as followingCount",
			Options: dbfuncs.DoSQLOption("senderUserID=? OR (receiverUserID=? AND value = 0)", "", "", ID, ID),
		}
		eventsQ := dbfuncs.SQLSelectParams{
			Table:   "Events",
			What:    "COUNT(id) as eventsCount",
			Options: dbfuncs.DoSQLOption("userID=?", "", "", ID),
		}
		groupsQ := dbfuncs.SQLSelectParams{
			Table:   "Relations",
			What:    "COUNT(id) as groupsCount",
			Options: dbfuncs.DoSQLOption("senderUserID=? AND receiverGroupID IS NOT NULL", "", "", ID),
		}
		mediaQ := dbfuncs.SQLSelectParams{
			Table:   "Media",
			What:    "COUNT(id) as galleryCount",
			Options: dbfuncs.DoSQLOption("userID=?", "", "", ID),
		}

		datas, e := dbfuncs.GetWithSubqueries(
			mainQ,
			[]dbfuncs.SQLSelectParams{followersQ, followingQ, eventsQ, groupsQ, mediaQ},
			[]string{"followersCount", "followingCount", "eventsCount", "groupsCount", "galleryCount"},
			dbfuncs.Users{},
		)
		if len(datas) == 0 || e != nil {
			data.Err = "no data"
		}
		data.Data = datas

		doJS(w, data)
	}
}

// HPost for handle '/api/post'
func (app *Application) HPost(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		data := API_RESPONSE{
			Err:  "ok",
			Data: "",
		}

		ID, e := strconv.Atoi(r.FormValue("id"))
		if e != nil {
			data.Err = "wrong id"
			doJS(w, data)
			return
		}

		userID := getUserIDfromReq(w, r)
		if !isHaveAccessToPost(ID, userID) {
			data.Err = "not have access"
			doJS(w, data)
			return
		}

		carmaQ := carmaQ("postID", ID)

		userJoin := userJoin("u.id = p.userID")
		groupJoin := groupJoin("g.id = p.groupID")
		likesJoin := likeJoin("l.userID = p.userID AND l.postID = p.id")
		mainQ := dbfuncs.SQLSelectParams{
			Table:   "Posts as p",
			What:    "p.*, u.nName, u.ava, u.status, g.title, g.ava, l.id IS NOT NULL",
			Options: dbfuncs.DoSQLOption("p.id = ?", "", "", ID),
			Joins:   []dbfuncs.SQLJoin{userJoin, groupJoin, likesJoin},
		}

		if data.Data, e = dbfuncs.GetWithSubqueries(
			mainQ,
			[]dbfuncs.SQLSelectParams{carmaQ},
			[]string{"nickname", "userAvatar", "status", "groupTitle", "groupAvatar", "isLiked", "carma"},
			dbfuncs.Posts{},
		); e != nil {
			data.Err = e.Error()
		}
		doJS(w, data)
	}
}

// HEvent for handle '/api/event'
func (app *Application) HEvent(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		data := API_RESPONSE{
			Err:  "ok",
			Data: "",
		}

		ID, e := strconv.Atoi(r.FormValue("id"))
		if e != nil {
			data.Err = "wrong id"
			doJS(w, data)
			return
		}

		userID := getUserIDfromReq(w, r)
		if !isHaveAccess(ID, userID, "Events") {
			data.Err = "not have access"
			doJS(w, data)
			return
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

		datas, e := dbfuncs.GetWithSubqueries(
			mainQ,
			[]dbfuncs.SQLSelectParams{eventAnswersGoingQ, eventAnswersNotGoingQ, eventAnswersIDKQ},
			[]string{"nickname", "userAvatar", "status", "groupTitle", "groupAvatar", "myVote", "votes0", "votes1", "votes2"},
			dbfuncs.Events{},
		)
		if e != nil {
			data.Err = "wrong data"
			doJS(w, data)
			return
		}
		data.Data = datas
		doJS(w, data)
	}
}

// Hposts for handle '/api/posts'
func (app *Application) Hposts(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		first, count := getLimit(r)
		tags := r.FormValue("tags")
		op := dbfuncs.DoSQLOption("tags LIKE '%"+tags+"%'", "datetime(date) DESC", "?,?", first, count)

		if tags == "all" {
			op.Where = ""
		} else if tags == "top" {
			op.Where = ""
			op.Order += ", carma DESC"
		} else if tags == "my" {
			userID := getUserIDfromReq(w, r)
			if userID != -1 {
				op.Where = "userID=?"
				op.Args = append([]interface{}{userID}, op.Args...)
			}
		}

		generalGet(
			w,
			r,
			dbfuncs.SQLSelectParams{
				Table:   "Posts",
				What:    "*",
				Options: op,
				Joins:   nil,
			},
			dbfuncs.Posts{},
		)
	}
}

// Hcomments for handle '/api/comments'
func (app *Application) Hcomments(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		first, count := getLimit(r)
		parentID, e := strconv.Atoi(r.FormValue("parentID"))
		if e != nil {
			return
		}
		commentType := r.FormValue("type")
		if commentType != "post" && commentType != "comment" && commentType != "user" {
			return
		}
		commentType += "ID"
		op := dbfuncs.DoSQLOption(commentType+"=?", "", "?,?", parentID, first, count)
		join := dbfuncs.DoSQLJoin(dbfuncs.INJOINQ, "Users as u", "u.id = c.userID")
		generalGet(
			w,
			r,
			dbfuncs.SQLSelectParams{
				Table:   "Comments as c",
				What:    "c.*, u.photo, u.nickname",
				Options: op,
				Joins:   []dbfuncs.SQLJoin{join},
			},
			dbfuncs.Comments{},
			"userPhoto",
			"nickname",
		)
	}
}

// Husers for handle '/api/users'
func (app *Application) Husers(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		first, count := getLimit(r)
		criterie := r.FormValue("criterie")
		if criterie != "all" && criterie != "friends" && criterie != "wrote" {
			return
		}

		params := dbfuncs.SQLSelectParams{
			Table: "Users as u",
			What:  "DISTINCT u.*",
			Joins: nil,
		}
		op := dbfuncs.DoSQLOption("", "u.nickname ASC", "?,?")
		if criterie == "friends" {
			userID := getUserIDfromReq(w, r)
			if userID == -1 {
				return
			}
			IDstring := strconv.Itoa(userID)
			op.Where = "friends LIKE '%" + IDstring + "%' AND ID != " + IDstring + ""
		} else if criterie == "wrote" {
			userID := getUserIDfromReq(w, r)
			if userID == -1 {
				return
			}
			op.Order = "datetime(m.date) DESC, " + op.Order
			params.Joins = []dbfuncs.SQLJoin{dbfuncs.DoSQLJoin(dbfuncs.LOJOINQ, "Messages as m", "u.id=m.receiverID AND m.senderID=?", userID)}
		}
		op.Args = append(op.Args, first, count)
		params.Options = op
		generalGet(
			w,
			r,
			params,
			dbfuncs.Users{},
		)
	}
}

// Hmessages for handle '/api/messages'
func (app *Application) Hmessages(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		userID := getUserIDfromReq(w, r)

		if userID == -1 {
			return
		}

		first, count := getLimit(r)
		receiverID := r.FormValue("nickname")
		addresserID := r.FormValue("username")
		op := dbfuncs.DoSQLOption("senderID=? AND receiverID=? OR senderID=? AND receiverID=?", "datetime(date) DESC", "?,?", addresserID, receiverID, receiverID, addresserID, first, count)

		generalGet(
			w,
			r,
			dbfuncs.SQLSelectParams{
				Table:   "Messages",
				What:    "*",
				Options: op,
				Joins:   nil,
			},
			dbfuncs.Messages{},
		)
	}
}

// HonlineUsers for handle '/api/onlines'
func (app *Application) HonlineUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		data := []*WSUser{}
		for _, v := range app.OnlineUsers {
			if v.ID > 0 {
				data = append(data, v)
			}
		}
		doJS(w, data)
	}
}

/* --------------------------------------------- Logical ---------------------------------- */
// ---------------------------------------------- Sign ---------------------------------------

// HcheckUserLogged for handle '/status'
func (app *Application) HCheckUserLogged(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		data := API_RESPONSE{
			Err:  "ok",
			Data: "",
		}

		userID := getUserIDfromReq(w, r)
		if userID == -1 {
			data.Err = "Not logged"
			doJS(w, data)
			return
		}

		// if app.findUserByID(userID) == nil {
		// 	data.Err = "you already logged"
		// 	doJS(w, data)
		// 	return
		// }
		data.Data = map[string]int{"id": userID}
		doJS(w, data)
	}
}

// HSaveUser for handle '/sign/s/'
func (app *Application) HSaveUser(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		data := API_RESPONSE{
			Err:  "ok",
			Data: "",
		}

		ID, e := app.SaveUser(w, r)
		if e != nil {
			data.Err = e.Error()
		} else {
			data.Data = map[string]interface{}{"id": ID}
		}
		doJS(w, data)
	}
}

// HSignUp for handle '/sign/up' && '/sign/oauth/up'
func (app *Application) HSignUp(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		data := API_RESPONSE{
			Err:  "ok",
			Data: "Message sended to your email. Check it",
		}

		oauth2Data, e := app.SignUp(w, r)
		if e != nil {
			data.Err = e.Error()
		}
		if oauth2Data != nil {
			data.Data = oauth2Data
		}
		doJS(w, data)
	}
}

// HSignIn for handle '/sign/in' && '/sign/oauth/in'
func (app *Application) HSignIn(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		data := API_RESPONSE{
			Err:  "ok",
			Data: "",
		}

		ID, e := app.SignIn(w, r)
		if e != nil {
			data.Err = e.Error()
		} else {
			data.Data = map[string]int{"id": ID}
		}
		doJS(w, data)
	}
}

// HSaveNewPassword for handle '/sign/rst/'
func (app *Application) HSaveNewPassword(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		data := API_RESPONSE{
			Err:  "ok",
			Data: "",
		}

		if e := app.SaveNewPassword(w, r); e != nil {
			data.Err = e.Error()
		}
		doJS(w, data)
	}
}

// HRestore for handle '/sign/re'
func (app *Application) HResetPassword(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		data := API_RESPONSE{
			Err:  "ok",
			Data: "",
		}

		if e := app.ResetPassword(w, r); e != nil {
			data.Err = e.Error()
		}
		doJS(w, data)
	}
}

// HLogout for handle '/sign/out'
func (app *Application) HLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {

		data := API_RESPONSE{
			Err:  "ok",
			Data: "",
		}

		userID := getUserIDfromReq(w, r)

		if userID == -1 {
			data.Err = "You are not logged"
			doJS(w, data)
			return
		}
		// app.Messages <- &WSMessage{MsgType: UserTypeOffType, AddresserID: "server", ReceiverID: "all", Body: *app.findUserByID(userID)}

		if e := app.Logout(w, r); e != nil {
			data.Err = e.Error()
		}
		doJS(w, data)
	}
}

// ------------------------------------------- Profile ------------------------------------------

// HChangeAvatar for handle '/profile/change-avatar'
func (app *Application) HChangeAvatar(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		userID := getUserIDfromReq(w, r)
		if userID == -1 {
			return
		}

		data := map[string]interface{}{"msg": "ok"}

		newPhoto, e := uploadFile("avatar", "img", r)
		if e != nil {
			data["msg"] = e.Error()
			doJS(w, data)
			return
		}
		data["link"] = newPhoto

		e = dbfuncs.ChangeUser(&dbfuncs.Users{ID: userID, Avatar: newPhoto})
		if e != nil {
			data["msg"] = e.Error()
		}
		doJS(w, data)
	}
}

// HChangeData for handle '/profile/change-profile'
func (app *Application) HChangeData(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		userID := getUserIDfromReq(w, r)
		if userID == -1 {
			return
		}

		data := map[string]interface{}{"msg": "ok"}
		e := checkEmailAndNick(false, r.PostFormValue("email"), r.PostFormValue("nick"))
		if e != nil {
			data["msg"] = e.Error()
			doJS(w, data)
			return
		}

		if app.XCSSOther(r.PostFormValue("nick")) != nil ||
			app.XCSSOther(r.PostFormValue("firstName")) != nil ||
			app.XCSSOther(r.PostFormValue("lastName")) != nil {
			data["msg"] = "wrong data"
			doJS(w, data)
			return
		}

		user := &dbfuncs.Users{ID: userID, NickName: r.PostFormValue("nick"), Email: r.PostFormValue("email"),
			FirstName: r.PostFormValue("firstName"), LastName: r.PostFormValue("lastName")}
		age, e := strconv.Atoi(r.PostFormValue("age"))
		if e == nil {
			user.Age = age
		}

		e = dbfuncs.ChangeUser(user)
		if e != nil {
			data["msg"] = e.Error()
		}
		doJS(w, data)
	}
}

// ------------------------------------------- Post ------------------------------------------

// HSavePost create one post
func (app *Application) HSavePost(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		userID := getUserIDfromReq(w, r)
		if userID == -1 {
			return
		}

		data := map[string]interface{}{"msg": "ok"}
		// XCSS
		if app.XCSSOther(r.PostFormValue("title")) != nil ||
			app.XCSSOther(r.PostFormValue("tags")) != nil {
			data["msg"] = "wrong data"
			doJS(w, data)
			return
		}

		p := dbfuncs.Posts{Title: r.PostFormValue("title"), Body: r.PostFormValue("body"),
			UserID: userID, UnixDate: int(time.Now().Unix())}
		pid, e := dbfuncs.CreatePost(&p)
		p.ID = pid
		if e != nil {
			data["msg"] = e.Error()
		}
		data["post"] = p
		doJS(w, data)
	}
}

// HSaveMessage create one message
func (app *Application) HSaveMessage(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		userID := getUserIDfromReq(w, r)
		if userID == -1 {
			return
		}

		data := API_RESPONSE{
			Err:  "ok",
			Data: "",
		}

		receiverID, e := strconv.Atoi(r.PostFormValue("receiverID"))
		if e != nil {
			data.Err = "not have receiver"
			doJS(w, data)
			return
		}

		m := dbfuncs.Messages{UnixDate: int(time.Now().Unix()), Body: r.PostFormValue("body"),
			SenderUserID: userID, ReceiverUserID: receiverID}

		if e := dbfuncs.CreateMessage(&m); e != nil {
			data.Err = "not save message"
		}
		doJS(w, data)
	}
}

// HSaveImage save one image
func (app *Application) HSaveImage(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		userID := getUserIDfromReq(w, r)
		if userID == -1 {
			return
		}

		data := map[string]interface{}{"msg": "ok"}
		filename, e := uploadFile("img", "img", r)
		if e != nil {
			data["msg"] = e.Error()
		}
		data["fname"] = filename
		doJS(w, data)
	}
}

// HSaveFile save one file
func (app *Application) HSaveFile(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		userID := getUserIDfromReq(w, r)
		if userID == -1 {
			return
		}

		data := map[string]interface{}{"msg": "ok"}
		filename, e := uploadFile("file", r.PostFormValue("type"), r)
		if e != nil {
			data["msg"] = e.Error()
		}
		data["fname"] = filename
		doJS(w, data)
	}
}

// HSaveComment save one comment
func (app *Application) HSaveComment(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		userID := getUserIDfromReq(w, r)
		if userID == -1 {
			return
		}

		data := map[string]interface{}{"msg": "ok"}
		ID, e := strconv.Atoi(r.PostFormValue("id"))
		if e != nil {
			data["msg"] = e.Error()
			doJS(w, data)
			return
		}

		c := dbfuncs.Comments{Body: r.PostFormValue("body"), UnixDate: int(time.Now().Unix()), UserID: userID, PostID: ID}
		if r.PostFormValue("type") == "comment" {
			c.PostID = 0
			c.CommentID = ID
			dbfuncs.ChangeComment(&dbfuncs.Comments{ID: ID, IsHaveChild: "1"})
		}

		cID, e := dbfuncs.CreateComment(&c)
		if e != nil {
			data["msg"] = e.Error()
		}
		c.ID = cID
		data["comment"] = c
		doJS(w, data)
	}
}

// HSaveLikeDislike save one like/dislike
func (app *Application) HSaveLikeDislike(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		data := API_RESPONSE{
			Err:  "ok",
			Data: "",
		}

		userID := getUserIDfromReq(w, r)
		if userID == -1 {
			data.Err = "not logged"
			doJS(w, data)
			return
		}

		ID, e := strconv.Atoi(r.PostFormValue("id"))
		if e != nil {
			data.Err = "wrong id"
			doJS(w, data)
			return
		}

		like := &dbfuncs.Likes{UserID: userID}
		typeLike := r.PostFormValue("type")
		if typeLike == "post" {
			like.PostID = ID
		} else if typeLike == "comment" {
			like.CommentID = ID
		} else {
			like.MediaID = ID
		}

		setted, e := dbfuncs.SetLikes(like)
		if e != nil {
			data.Err = "not setted"
			doJS(w, data)
			return
		}

		carmaAny, e := dbfuncs.GetOneFrom(carmaQ(typeLike+"ID", ID))
		if e != nil || carmaAny[0] == nil {
			data.Err = e.Error()
			doJS(w, data)
			return
		}

		data.Data = map[string]interface{}{"carma": carmaAny[0], "isLiked": setted}
		doJS(w, data)
	}
}

// HSaveEventAnswer save one event answer
func (app *Application) HSaveEventAnswer(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		data := API_RESPONSE{
			Err:  "ok",
			Data: "",
		}

		userID := getUserIDfromReq(w, r)
		if userID == -1 {
			data.Err = "not logged"
			doJS(w, data)
			return
		}

		ID, e := strconv.Atoi(r.PostFormValue("id"))
		if e != nil {
			data.Err = "wrong id"
			doJS(w, data)
			return
		}

		if e := dbfuncs.CreateEventAnswer(&dbfuncs.EventAnswers{UserID: userID, EventID: ID, Answer: r.PostFormValue("answer")}); e != nil {
			data.Err = "wrong event answer"
		}

		goingQ := eventAnswerQ(0, ID)
		notGoingQ := eventAnswerQ(1, ID)
		idkQ := eventAnswerQ(2, ID)
		votes, e := dbfuncs.GetWithSubqueries(goingQ, []dbfuncs.SQLSelectParams{notGoingQ, idkQ}, []string{"votes0", "votes1", "votes2"}, struct{}{})
		if e != nil {
			data.Err = e.Error()
		}
		data.Data = votes[0]
		doJS(w, data)
	}
}

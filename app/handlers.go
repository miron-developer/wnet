package app

import (
	"anifor/app/dbfuncs"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// SecureHeaderMiddleware set secure header option
func (app *Application) SecureHeaderMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-XSS-Protection", "1;mode=block")
		w.Header().Set("X-Frame-Options", "deny")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET")
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

// HTTPRedirect redirector to https
func (app *Application) HTTPRedirect(w http.ResponseWriter, r *http.Request) {
	target := "https://" + r.Host[:strings.Index(r.Host, ":")+1] + app.HTTPSport + r.URL.Path
	if len(r.URL.RawQuery) > 0 {
		target += "?" + r.URL.RawQuery
	}
	http.Redirect(w, r, target, http.StatusTemporaryRedirect)
}

// Hindex for handle '/'
func (app *Application) Hindex(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		app.eHandler(w, app.parseHTMLFiles(w, "index.html", nil), "can't load this page", 500)
	}
}

/* ----------------------------------------------- Websocket ---------------------------------------------- */

// CreateWSUser create one WSUser
func (app *Application) CreateWSUser(w http.ResponseWriter, r *http.Request) {
	conn, e := upgrader.Upgrade(w, r, nil)
	if e != nil {
		app.ELog.Println(e, r.RemoteAddr)
		return
	}

	name := StringWithCharset(8)
	user := &WSUser{Conn: conn, Nickname: name, WSName: name}
	app.m.Lock()
	app.OnlineUsers[user.Nickname] = user
	app.m.Unlock()

	go app.OnlineUsers[user.Nickname].HandleUserMsg(app)
	go app.OnlineUsers[user.Nickname].Pinger()
	app.Messages <- &WSMessage{Addresser: "server", Receiver: user.Nickname, MsgType: AuthType, Body: user.Nickname}
}

// ChangeName change logged name
func (app *Application) ChangeName(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		data := map[string]interface{}{"msg": "ok"}
		ID, e := strconv.Atoi(r.PostFormValue("id"))
		if e != nil {
			data["msg"] = e.Error()
			doJS(w, data)
			return
		}

		e = app.ChangeWSUserName(r.PostFormValue("oldname"), r.PostFormValue("newname"), ID)
		if e == nil {
			app.Messages <- &WSMessage{MsgType: UserTypeOnType, Addresser: "server", Receiver: "all", Body: app.findUserByNick(r.PostFormValue("newname"))}
		}
		doJS(w, data)
	}
}

/* ------------------------------------------- API ------------------------------------------------ */

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
		generalGet(w, r, getAPIOption{tableName: "Posts", whatGet: "*", op: op, joins: nil, sample: dbfuncs.Posts{}})
	}
}

// Hpost for handle '/api/post/'
func (app *Application) Hpost(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		ID, e := strconv.Atoi(r.FormValue("postID"))
		if e != nil {
			return
		}
		generalGet(w, r, getAPIOption{tableName: "Posts", whatGet: "*", op: dbfuncs.DoSQLOption("ID=?", "", "", ID), joins: nil, sample: dbfuncs.Posts{}})
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
			getAPIOption{tableName: "Comments as c", whatGet: "c.*, u.photo, u.nickname", op: op,
				joins: []dbfuncs.SQLJoin{join}, sample: dbfuncs.Comments{},
			},
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

		getOp := getAPIOption{tableName: "Users as u", whatGet: "DISTINCT u.*", sample: dbfuncs.Users{}}
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
			getOp.joins = []dbfuncs.SQLJoin{dbfuncs.DoSQLJoin(dbfuncs.LOJOINQ, "Messages as m", "u.id=m.receiverID AND m.senderID=?", userID)}
		}
		op.Args = append(op.Args, first, count)
		getOp.op = op
		generalGet(w, r, getOp)
	}
}

// Huser for handle '/api/user/'
func (app *Application) Huser(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		ID, e := strconv.Atoi(r.FormValue("userID"))
		if e != nil {
			return
		}
		generalGet(w, r, getAPIOption{tableName: "Users", whatGet: "*", op: dbfuncs.DoSQLOption("ID=?", "", "", ID), joins: nil, sample: dbfuncs.Users{}})
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
		receiverID := app.findUserByNick(r.FormValue("nickname")).ID
		addresserID := app.findUserByNick(r.FormValue("username")).ID
		op := dbfuncs.DoSQLOption("senderID=? AND receiverID=? OR senderID=? AND receiverID=?", "datetime(date) DESC", "?,?", addresserID, receiverID, receiverID, addresserID, first, count)
		generalGet(w, r, getAPIOption{tableName: "Messages", whatGet: "*", op: op, joins: nil, sample: dbfuncs.Messages{}})
	}
}

// HonlineUsers for handle '/api/online'
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
// --------------------------------------------- Sign ----------------------------------------

// HcheckUserLogged for handle '/status'
func (app *Application) HcheckUserLogged(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		data := map[string]interface{}{"state": -1}
		userID := getUserIDfromReq(w, r)
		if userID != -1 {
			data["state"] = 1
		}

		if userID != -1 && app.findUserByID(userID) != nil {
			data["state"] = -1
		}
		data["id"] = userID
		doJS(w, data)
	}
}

// HSaveUser for handle '/sign/s/'
func (app *Application) HSaveUser(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		data := map[string]interface{}{"msg": "ok"}
		ID, e := app.saveUser(w, r)
		if e != nil {
			data["msg"] = e.Error()
		} else {
			data["id"] = ID
		}
		doJS(w, data)
	} else {
		app.Hindex(w, r)
	}
}

// HSignUp for handle '/sign/up'
func (app *Application) HSignUp(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		data := map[string]interface{}{"msg": "ok"}
		e := app.signUp(w, r)
		if e != nil {
			data["msg"] = e.Error()
		}
		doJS(w, data)
	} else {
		app.Hindex(w, r)
	}
}

// HSignIn for handle '/sign/in'
func (app *Application) HSignIn(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		data := map[string]interface{}{"msg": "ok"}
		ID, e := app.signIn(w, r)
		if e != nil {
			data["msg"] = e.Error()
		} else {
			data["id"] = ID
		}
		doJS(w, data)
	} else {
		app.Hindex(w, r)
	}
}

// HSavePassword for handle '/sign/r/'
func (app *Application) HSavePassword(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		data := map[string]interface{}{"msg": "ok"}
		e := app.savePassword(w, r)
		if e != nil {
			data["msg"] = e.Error()
		}
		doJS(w, data)
	} else {
		app.Hindex(w, r)
	}
}

// HRestore for handle '/sign/restore'
func (app *Application) HRestore(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		data := map[string]interface{}{"msg": "ok"}
		e := app.restore(w, r)
		if e != nil {
			data["msg"] = e.Error()
		}
		doJS(w, data)
	} else {
		app.Hindex(w, r)
	}
}

// HLogout for handle '/sign/logout'
func (app *Application) HLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		data := map[string]interface{}{"msg": "ok"}
		nickname := r.PostFormValue("nickname")
		newname := StringWithCharset(8)

		app.Messages <- &WSMessage{MsgType: UserTypeOffType, Addresser: "server", Receiver: "all", Body: *app.findUserByNick(nickname)}
		e := app.ChangeWSUserName(nickname, newname, 0)
		if e == nil {
			app.Messages <- &WSMessage{MsgType: AuthType, Addresser: "server", Receiver: newname, Body: newname}
		}

		e = logout(w, r)
		if e != nil {
			data["msg"] = e.Error()
		}
		doJS(w, data)
	} else {
		app.Hindex(w, r)
	}
}

// ------------------------------------------- Profile ------------------------------------------

// HChangeAvatar for handle '/profile/change-avatar'
func (app *Application) HChangeAvatar(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		userID := getUserIDfromReq(w, r)
		if userID == -1 {
			app.Hindex(w, r)
			return
		}

		data := map[string]interface{}{"msg": "ok"}

		newPhoto, e := uploadFile("avatar", "avatar", "img", r)
		if e != nil {
			data["msg"] = e.Error()
			doJS(w, data)
			return
		}
		data["link"] = newPhoto

		e = dbfuncs.ChangeUser(&dbfuncs.Users{ID: userID, Photo: newPhoto}, r.PostFormValue("oldphoto"))
		if e != nil {
			data["msg"] = e.Error()
		}
		doJS(w, data)
	} else {
		app.Hindex(w, r)
	}
}

// HChangeData for handle '/profile/change-profile'
func (app *Application) HChangeData(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		userID := getUserIDfromReq(w, r)
		if userID == -1 {
			app.Hindex(w, r)
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

		e = dbfuncs.ChangeUser(user, "")
		if e != nil {
			data["msg"] = e.Error()
		}
		doJS(w, data)
	} else {
		app.Hindex(w, r)
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
		body := r.PostFormValue("body")
		for {
			st := strings.Index(body, "<")
			fn := strings.Index(body, ">")
			cur := body[st : fn+1]
			if app.XCSSPostBody(cur) != nil {
				data["msg"] = "wrong data"
				doJS(w, data)
				return
			}
			body = body[fn+1:]
			if len(body) == 0 {
				break
			}
		}

		p := dbfuncs.Posts{Title: r.PostFormValue("title"), Body: r.PostFormValue("body"),
			Tags: r.PostFormValue("tags"), UserID: userID, Date: TimeExpire(time.Nanosecond)}
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

		m := dbfuncs.Messages{Date: TimeExpire(time.Nanosecond), Body: r.PostFormValue("body"),
			SenderID: userID, ReceiverID: app.findUserByNick(r.PostFormValue("receiver")).ID}
		data := map[string]interface{}{"msg": "ok"}
		if e := dbfuncs.CreateMessage(&m); e != nil {
			data["msg"] = e.Error()
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
		filename, e := uploadFile("img", "post", "img", r)
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
		filename, e := uploadFile("file", r.PostFormValue("place"), r.PostFormValue("type"), r)
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

		c := dbfuncs.Comments{Body: r.PostFormValue("body"), Date: TimeExpire(time.Nanosecond), UserID: userID, PostID: ID}
		if r.PostFormValue("type") == "comment" {
			c.PostID = 0
			c.CommentID = ID
			dbfuncs.ChangeCommentNestedState(ID)
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

		typeLD := r.PostFormValue("type")
		tablename := "Comments"
		if typeLD == "post" {
			tablename = "Posts"
		} else if typeLD == "user" {
			tablename = "Users"
			typeLD = "whos"
		}

		value := 1
		if r.PostFormValue("value") == "dislike" {
			value = -1
		}
		column := typeLD + "ID"

		ld := dbfuncs.LDGeneral{Value: value, UserID: userID, WhosID: ID, TableName: "LD" + tablename}

		e = dbfuncs.SetLD(&ld, column)
		if e != nil {
			data["msg"] = e.Error()
			doJS(w, data)
			return
		}

		carmaAny, e := dbfuncs.GetOneFrom(ld.TableName, "sum(value)", dbfuncs.DoSQLOption(column+"=?", "", "", ID), nil)
		if e != nil {
			data["msg"] = e.Error()
			doJS(w, data)
			return
		}

		carma := 0
		if carmaAny[0] != nil {
			carma = dbfuncs.FromINT64ToINT(carmaAny[0])
		}

		e = dbfuncs.ChangeCarma(tablename, strconv.Itoa(ID), strconv.Itoa(carma))
		if e != nil {
			data["msg"] = e.Error()
			doJS(w, data)
			return
		}
		data["carma"] = carma
		doJS(w, data)
	}
}

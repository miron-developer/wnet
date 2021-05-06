package app

import (
	"errors"
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
		accessOrigin := "http://localhost:3000"
		if app.IsHeroku {
			accessOrigin = "https://wnet-sn.herokuapp.com"
		}
		w.Header().Set("Access-Control-Allow-Origin", accessOrigin)
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

// HIndex handle all GETs
func (app *Application) HIndex(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("this is server-api"))
}

/* ------------------------------------------- API ------------------------------------------------ */

func (app *Application) HApi(w http.ResponseWriter, r *http.Request, f func(w http.ResponseWriter, r *http.Request) (interface{}, error)) {
	if r.Method == "GET" {
		data := API_RESPONSE{
			Err:  "ok",
			Data: "",
		}

		ids, e := f(w, r)
		if e != nil {
			data.Err = e.Error()
		}
		data.Data = ids
		doJS(w, data)
	}
}

// HNews for handle '/api/news'
func (app *Application) HNews(w http.ResponseWriter, r *http.Request) {
	app.HApi(w, r, app.News)
}

// HPublications for handle '/api/publications'
func (app *Application) HPublications(w http.ResponseWriter, r *http.Request) {
	app.HApi(w, r, app.Publications)
}

// HNotifications for handle '/api/notifications'
func (app *Application) HNotifications(w http.ResponseWriter, r *http.Request) {
	app.HApi(w, r, app.Notifications)
}

// HGallery for handle '/api/gallery'
func (app *Application) HGallery(w http.ResponseWriter, r *http.Request) {
	app.HApi(w, r, app.Gallery)
}

// HUsers for handle '/api/followers'
func (app *Application) HUsers(w http.ResponseWriter, r *http.Request) {
	app.HApi(w, r, app.Users)
}

// HGroups for handle '/api/groups'
func (app *Application) HGroups(w http.ResponseWriter, r *http.Request) {
	app.HApi(w, r, app.Groups)
}

// HChats for handle '/api/chats'
func (app *Application) HChats(w http.ResponseWriter, r *http.Request) {
	app.HApi(w, r, app.Chats)
}

// HEvents for handle '/api/events'
func (app *Application) HEvents(w http.ResponseWriter, r *http.Request) {
	app.HApi(w, r, app.Events)
}

// HMessages for handle '/api/messages'
func (app *Application) HMessages(w http.ResponseWriter, r *http.Request) {
	app.HApi(w, r, app.Messages)
}

// HNotification for handle '/api/notification'
func (app *Application) HNotification(w http.ResponseWriter, r *http.Request) {
	app.HApi(w, r, app.Notification)
}

// HUser for handle '/api/user/'
func (app *Application) HUser(w http.ResponseWriter, r *http.Request) {
	app.HApi(w, r, app.User)
}

// HGroup for handle '/api/group'
func (app *Application) HGroup(w http.ResponseWriter, r *http.Request) {
	app.HApi(w, r, app.Group)
}

// HPost for handle '/api/post'
func (app *Application) HPost(w http.ResponseWriter, r *http.Request) {
	app.HApi(w, r, app.Post)
}

// HEvent for handle '/api/event'
func (app *Application) HEvent(w http.ResponseWriter, r *http.Request) {
	app.HApi(w, r, app.Event)
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
			dbfuncs.Post{},
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
			dbfuncs.Comment{},
			"userPhoto",
			"nickname",
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

// HGetFile save one file
func (app *Application) HGetFile(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		app.GetFile(w, r)
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
			data.Err = "not logged"
			doJS(w, data)
			return
		}

		u := dbfuncs.User{ID: userID, Status: "online"}
		go u.Change()

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

// ------------------------------------------- Change ------------------------------------------

// HConfirmSettings save user settings
func (app *Application) HConfirmSettings(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		data := API_RESPONSE{
			Err:  "ok",
			Data: "",
		}

		if e := app.ConfirmChangeSettings(w, r); e != nil {
			data.Err = e.Error()
		}
		doJS(w, data)
	}
}

// HChangeSettings save user settings
func (app *Application) HChangeSettings(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		data := API_RESPONSE{
			Err:  "ok",
			Data: "",
		}

		if e := app.ChangeSettings(w, r); e != nil {
			data.Err = e.Error()
		}
		doJS(w, data)
	}
}

// HChangeProfile user/group data
func (app *Application) HChangeProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		data := API_RESPONSE{
			Err:  "ok",
			Data: "",
		}

		if e := app.ChangeProfile(w, r); e != nil {
			data.Err = e.Error()
		}
		doJS(w, data)
	}
}

// ------------------------------------------- Save ------------------------------------------

// general handler for all save paths
func (app *Application) HSaves(w http.ResponseWriter, r *http.Request, f func(w http.ResponseWriter, r *http.Request) (interface{}, error)) {
	if r.Method == "POST" {
		data := API_RESPONSE{
			Err:  "ok",
			Data: "",
		}

		datas, e := f(w, r)
		if e != nil {
			data.Err = e.Error()
		}
		data.Data = datas
		doJS(w, data)
	}
}

// HSavePost create post
func (app *Application) HSavePost(w http.ResponseWriter, r *http.Request) {
	app.HSaves(w, r, app.CreatePost)
}

// HSaveRelation create relation
func (app *Application) HSaveRelation(w http.ResponseWriter, r *http.Request) {
	app.HSaves(w, r, app.CreateRelation)
}

// HSaveGroup create group
func (app *Application) HSaveGroup(w http.ResponseWriter, r *http.Request) {
	app.HSaves(w, r, app.CreateGroup)
}

// HSaveEvent create event
func (app *Application) HSaveEvent(w http.ResponseWriter, r *http.Request) {
	app.HSaves(w, r, app.CreateEvent)
}

// HSaveMedia save media
func (app *Application) HSaveMedia(w http.ResponseWriter, r *http.Request) {
	app.HSaves(w, r, app.CreateMedia)
}

// HSaveFile save file
func (app *Application) HSaveFile(w http.ResponseWriter, r *http.Request) {
	app.HSaves(w, r, app.CreateFile)
}

// HSaveLikeDislike save like/dislike
func (app *Application) HSaveLikeDislike(w http.ResponseWriter, r *http.Request) {
	app.HSaves(w, r, app.CreateLike)
}

// HSaveEventAnswer save event answer
func (app *Application) HSaveEventAnswer(w http.ResponseWriter, r *http.Request) {
	app.HSaves(w, r, app.CreateEventAnswer)
}

// HSaveChat save chat
func (app *Application) HSaveChat(w http.ResponseWriter, r *http.Request) {
	app.HSaves(w, r, app.CreateChat)
}

// HSaveComment save comment
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

		c := dbfuncs.Comment{Body: r.PostFormValue("body"), UnixDate: int(time.Now().Unix() * 1000), UserID: userID, PostID: ID}
		if r.PostFormValue("type") == "comment" {
			c.PostID = 0
			c.CommentID = ID
			parentComment := &dbfuncs.Comment{ID: ID, IsHaveChild: "1"}
			parentComment.Change()
		}

		cID, e := c.Create()
		if e != nil {
			data["msg"] = e.Error()
		}
		c.ID = cID
		data["comment"] = c
		doJS(w, data)
	}
}

// HSaveMessage create message
func (app *Application) HSaveMessage(w http.ResponseWriter, r *http.Request) {
	app.HSaves(w, r, app.CreateMessage)
}

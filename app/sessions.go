package app

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"
	"wnet/app/dbfuncs"

	uuid "github.com/satori/go.uuid"
)

const sessionExpire = 24 * time.Hour
const timeSecond = time.Second
const cookieName = "wnetID"

// get uuid for session
func sessionID() string {
	var e error
	u1 := uuid.Must(uuid.NewV4(), e)
	return fmt.Sprint(u1)
}

func setCookie(w http.ResponseWriter, sid string, expire int) {
	sidCook := http.Cookie{
		Name:   cookieName,
		Value:  url.QueryEscape(sid),
		MaxAge: expire,

		Path:     "/",
		SameSite: http.SameSiteNoneMode,
		Secure:   true,
		HttpOnly: true,
	}
	http.SetCookie(w, &sidCook)
}

func updateStatus(userID int, status string) error {
	user := &dbfuncs.User{ID: userID, Status: status}
	return user.Change()
}

// SessionStart start user session
func SessionStart(w http.ResponseWriter, r *http.Request, login string, userID int) error {
	cookie, e := r.Cookie(cookieName)
	sidFromCookie := ""
	sidFromDB := ""
	isToCreate := false

	// get all sids
	if e == nil && cookie.Value != "" {
		sidFromCookie, _ = url.QueryUnescape(cookie.Value)
	}

	res, e := dbfuncs.GetOneFrom(dbfuncs.SQLSelectParams{
		Table:   "Sessions",
		What:    "id",
		Options: dbfuncs.DoSQLOption("userID = ?", "", "", userID),
		Joins:   nil,
	})
	if res != nil && e == nil {
		sidFromDB = res[0].(string)
	}

	// select one sid
	sid := sidFromDB
	if sid == "" {
		isToCreate = true
		sid = sidFromCookie
	}
	if sid == "" {
		sid = sessionID()
	}

	// create or change session
	s := &dbfuncs.Session{ID: sid, Expire: TimeExpire(sessionExpire), UserID: userID}
	if isToCreate {
		e = s.Create()
	} else {
		e = s.Change()
	}
	if e != nil {
		return errors.New("Session error")
	}

	setCookie(w, sid, int(sessionExpire/timeSecond))
	updateStatus(userID, "online")
	return nil
}

// SessionGC delete expired session
func (app *Application) SessionGC() error {
	return dbfuncs.DeleteByParams(dbfuncs.SQLDeleteParams{
		Table:   "Sessions",
		Options: dbfuncs.DoSQLOption("datetime(expire) < datetime('"+TimeExpire(time.Nanosecond)+"')", "", ""),
	})
}

var min = 0

// CheckPerMin call SessionGC per minute that delete expired sessions and do db backup
func (app *Application) CheckPerMin() {
	for {
		timer := time.NewTimer(1 * time.Minute)
		<-timer.C
		app.CurrentRequestCount = 0
		min++
		if min == 60*24 {
			min = 0
			app.UsersCode = map[string]*dbfuncs.User{}
			app.RestoreCode = map[string]string{}
		}
		if min == 30 {
			if e := app.DoBackup(); e == nil {
				app.ILog.Println("backup created!")
			} else {
				app.ELog.Println(e)
			}
		}
		if e := app.SessionGC(); e != nil {
			app.ELog.Println(e)
		}
	}
}

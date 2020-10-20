package app

import (
	"anifor/app/dbfuncs"
	"fmt"
	"net/http"
	"net/url"
	"time"

	uuid "github.com/satori/go.uuid"
)

const sessionExpire = 24 * time.Hour
const timeSecond = time.Second
const cookieName = "sesid"

// get uuid for session
func sessionID() string {
	var e error
	u1 := uuid.Must(uuid.NewV4(), e)
	return fmt.Sprint(u1)
}

func updateCooks(w http.ResponseWriter, sid string) {
	sidCook := http.Cookie{
		Name:     cookieName,
		Value:    url.QueryEscape(sid),
		Path:     "/",
		HttpOnly: true,
		MaxAge:   int(sessionExpire / timeSecond),
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, &sidCook)
}

func updateActivitie(userID int) {
	dbfuncs.ChangeUser(&dbfuncs.Users{ID: userID, LastActivitie: TimeExpire(time.Nanosecond)}, "")
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
	res, e := dbfuncs.GetOneFrom("Sessions", "ID", dbfuncs.DoSQLOption("userID = ?", "", "", userID), nil)
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
	s := &dbfuncs.Sessions{ID: sid, Expire: TimeExpire(sessionExpire), UserID: userID}
	if isToCreate {
		e = dbfuncs.CreateSession(s)
	} else {
		e = dbfuncs.ChangeSession(s)
	}
	if e != nil {
		return e
	}

	updateCooks(w, sid)
	return nil
}

// SessionGC delete expired session
func (app *Application) SessionGC() error {
	return dbfuncs.DeleteSession("datetime(expire) < datetime('" + TimeExpire(time.Nanosecond) + "')")
}

var min = 0

// CheckPerMin call SessionGC per minute that delete expired sessions
func (app *Application) CheckPerMin() {
	for {
		timer := time.NewTimer(1 * time.Minute)
		<-timer.C
		app.CurrentRequestCount = 0
		min++
		if min == 60*24 {
			min = 0
			app.UsersCode = map[string]*dbfuncs.Users{}
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

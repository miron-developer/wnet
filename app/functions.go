package app

import (
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"regexp"
	"time"
	"wnet/app/dbfuncs"
)

func StringWithCharset(length int) string {
	charset := "abcdefghijklmnopqrstuvwxyz" + "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	seededRand := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

// handle error and write it responseWriter
func (app *Application) eHandler(w http.ResponseWriter, e error, msg string, code int) bool {
	if e != nil {
		http.Error(w, msg, code)
		app.ELog.Println(e)
		return true
	}
	return false
}

// write in log each request
func logingReq(r *http.Request) string {
	return fmt.Sprintf("%v: '%v'\n", r.Method, r.URL)
}

// XCSSOther check
func (app *Application) XCSSOther(data string) error {
	if data == "" {
		return nil
	}

	rg := regexp.MustCompile(`^[\w\d@#'?!.():/,-_\s\p{Cyrillic}]+$`)
	if !rg.MatchString(data) {
		return errors.New("wrong data")
	}
	return nil
}

// TimeExpire time.Now().Add(some duration) and return it by string
func TimeExpire(add time.Duration) string {
	return time.Now().Add(add).Format("2006-01-02 15:04:05")
}

// DoBackup make backup every 30 min
func (app *Application) DoBackup() error {
	return dbfuncs.UploadDB()
}

// checkIsLogged check if user is logged
func checkIsLogged(r *http.Request) (string, error) {
	cookie, e := r.Cookie(cookieName)
	if e != nil {
		return "", errors.New("cookie not founded")
	}
	return url.QueryUnescape(cookie.Value)
}

func getUserIDfromReq(w http.ResponseWriter, r *http.Request) int {
	sesID, e := checkIsLogged(r)
	if sesID == "" || e != nil {
		return -1
	}

	userID, e := dbfuncs.GetOneFrom(dbfuncs.SQLSelectParams{
		Table:   "Sessions",
		What:    "userID",
		Options: dbfuncs.DoSQLOption("id = ?", "", "", sesID),
	})
	if e != nil {
		return -1
	}

	// update cooks & sess
	ses := &dbfuncs.Session{ID: sesID, Expire: TimeExpire(sessionExpire)}
	ses.Change()
	setCookie(w, sesID, int(sessionExpire/timeSecond))
	return dbfuncs.FromINT64ToINT(userID[0])
}

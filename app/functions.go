package app

import (
	"anifor/app/dbfuncs"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os/exec"
	"regexp"
	"time"
)

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

// XCSSPostBody check
func (app *Application) XCSSPostBody(data string) error {
	if !regexp.MustCompile(`^<+\w+\s+[class="]+[\w-\d\s]+["]+>+$`).MatchString(data) &&
		!regexp.MustCompile(`^<+\w+\s+[src="]+[\w-\d\/.]+["]+>+$`).MatchString(data) &&
		!regexp.MustCompile(`^<\/\w+>+$`).MatchString(data) {
		return errors.New("wrong data")
	}
	return nil
}

// XCSSOther check
func (app *Application) XCSSOther(data string) error {
	rg := regexp.MustCompile(`^[\w\d@#',_\s]+$`)
	if !rg.MatchString(data) {
		return errors.New("wrong data")
	}
	return nil
}

func (app *Application) parseHTMLFiles(w http.ResponseWriter, file string, data interface{}) error {
	t, ok := app.CachedTemplates[file]
	if !ok {
		return errors.New("Error: Not found cached template")
	}
	return t.Execute(w, data)
}

// TimeExpire time.Now().Add(some duration) and return it by string
func TimeExpire(add time.Duration) string {
	return time.Now().Add(add).Format("2006-01-02 15:04:05")
}

// DoBackup make backup every day
func (app *Application) DoBackup() error {
	cmd := exec.Command("cp", `db/anifor.db`, `db/anifor_backup.db`)
	return cmd.Run()
}

// checkIsLogged check if user is logged
func checkIsLogged(r *http.Request) (string, error) {
	cookie, e := r.Cookie(cookieName)
	if e != nil {
		return "", errors.New("cooks not founded")
	}
	return url.QueryUnescape(cookie.Value)
}

func getUserIDfromReq(w http.ResponseWriter, r *http.Request) int {
	sesID, e := checkIsLogged(r)
	if sesID == "" || e != nil {
		return -1
	}

	userID, e := dbfuncs.GetOneFrom("Sessions", "userID", dbfuncs.DoSQLOption("ID=?", "", "", sesID), nil)
	if e != nil {
		return -1
	}

	// update cooks & sess
	dbfuncs.ChangeSession(&dbfuncs.Sessions{ID: sesID, Expire: TimeExpire(sessionExpire)})
	updateCooks(w, sesID)
	return dbfuncs.FromINT64ToINT(userID[0])
}

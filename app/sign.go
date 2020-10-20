package app

import (
	"anifor/app/dbfuncs"
	"errors"
	"net/http"
	"net/smtp"
	"regexp"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// checkEmailAndNick check if email is empty or not
//	exist = true - user exist in db
func checkEmailAndNick(exist bool, email, nickname string) error {
	results, _ := dbfuncs.GetFrom("Users", "email, nickname", dbfuncs.DoSQLOption("email=? OR nickname=?", "", "", email, nickname), nil)

	if !exist && len(results) > 0 {
		if results[0][0].(string) == email {
			return errors.New("this email is not empty")
		}
		return errors.New("this nickname is not empty")
	}
	if exist && len(results) == 0 {
		return errors.New("wrong login")
	}
	return nil
}

// checkPassword check is password is valid(up) or correct password(in)
//	exist = true - user exist in db
func checkPassword(exist bool, pass, login string) error {
	if !exist {
		if !regexp.MustCompile(`[A-Z]`).MatchString(pass) {
			return errors.New("password must have A-Z")
		}
		if !regexp.MustCompile(`[a-z]`).MatchString(pass) {
			return errors.New("password must have a-z(small)")
		}
		if !regexp.MustCompile(`[0-9]`).MatchString(pass) {
			return errors.New("password must have 0-9")
		}
		if len(pass) < 8 {
			return errors.New("password must have at least 8 character")
		}
	} else {
		dbPass, e := dbfuncs.GetOneFrom("Users", "password", dbfuncs.DoSQLOption("email = ? OR nickname = ?", "", "", login, login), nil)
		if e != nil {
			return errors.New("wrong login")
		}
		return bcrypt.CompareHashAndPassword([]byte(dbPass[0].(string)), []byte(pass))
	}
	return nil
}

// finish signUp proccess
func (app *Application) saveUser(w http.ResponseWriter, r *http.Request) (int, error) {
	user, ok := app.UsersCode[r.PostFormValue("code")]
	if !ok {
		return -1, errors.New("wrong code")
	}

	ID, e := dbfuncs.CreateUser(user)
	if e != nil {
		return -1, e
	}
	return ID, SessionStart(w, r, user.Email, ID)
}

// signUp check validate, start session, send data to db and etc
func (app *Application) signUp(w http.ResponseWriter, r *http.Request) error {
	email := strings.ToLower(r.PostFormValue("email"))
	pass := r.PostFormValue("password")
	nickname := r.PostFormValue("nickname")

	if pass != r.PostFormValue("repeatPassword") {
		return errors.New("password mismatch")
	}
	e := checkEmailAndNick(false, email, nickname)
	if e != nil {
		return e
	}
	e = checkPassword(false, pass, "")
	if e != nil {
		return e
	}
	age, e := strconv.Atoi(r.PostFormValue("age"))
	if e != nil {
		return errors.New("wrong age")
	}
	hashPass, e := bcrypt.GenerateFromPassword([]byte(r.PostFormValue("password")), 4)
	if e != nil {
		return errors.New("password not correct")
	}

	// XCSS
	if app.XCSSOther(nickname) != nil ||
		app.XCSSOther(r.PostFormValue("firstName")) != nil ||
		app.XCSSOther(r.PostFormValue("lastName")) != nil {
		return errors.New("wrond data")
	}

	code := StringWithCharset(8)
	app.m.Lock()
	app.UsersCode[code] = &dbfuncs.Users{FirstName: r.PostFormValue("firstName"), LastName: r.PostFormValue("lastName"),
		Gender: r.PostFormValue("gender"), Age: age, Photo: "static/img/avatar/default.png",
		Role: "2", LastActivitie: TimeExpire(time.Nanosecond),
		NickName: nickname, Email: email, Password: string(hashPass)}
	app.m.Unlock()

	mes := "To: " + email + "\nFrom: " + "ani_for@bk.ru" + "\nSubject: Verification\n\n" +
		"You will be going to register on AniFor. \nEnter this code on site: " + code +
		"\nOr visit this: https://" + r.Host + "/sign/s/" + code +
		"\nBe careful this code retire at the end of the day."
	e = SendMail(mes, []string{email})

	if e != nil {
		return errors.New("wrong email")
	}
	return nil
}

// signIn check password and login from db and request
func (app *Application) signIn(w http.ResponseWriter, r *http.Request) (int, error) {
	login := r.PostFormValue("login")
	pass := r.PostFormValue("password")

	e := checkEmailAndNick(true, login, login)
	if e != nil {
		return -1, e
	}
	e = checkPassword(true, pass, login)
	if e != nil {
		return -1, errors.New("password is not correct")
	}

	res, e := dbfuncs.GetOneFrom("Users", "ID", dbfuncs.DoSQLOption("email = ? OR nickname = ?", "", "", login, login), nil)
	if e != nil {
		return -1, errors.New("wrong login")
	}
	ID := dbfuncs.FromINT64ToINT(res[0])

	if app.findUserByID(ID) != nil {
		return -1, errors.New("user already is online")
	}
	return ID, SessionStart(w, r, login, ID)
}

func logout(w http.ResponseWriter, r *http.Request) error {
	cookie, e := r.Cookie(cookieName)
	if e != nil || cookie.Value == "" {
		return e
	}

	e = dbfuncs.DeleteSession("ID = '" + cookie.Value + "'")
	if e != nil {
		return e
	}

	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	return nil
}

func (app *Application) savePassword(w http.ResponseWriter, r *http.Request) error {
	login, ok := app.RestoreCode[r.PostFormValue("code")]
	if !ok {
		return errors.New("wrong code")
	}

	newPass := r.PostFormValue("password")
	e := checkPassword(false, newPass, "")
	if e != nil {
		return e
	}

	res, e := dbfuncs.GetOneFrom("Users", "ID", dbfuncs.DoSQLOption("email = ? OR nickname = ?", "", "", login, login), nil)
	if e != nil {
		return e
	}

	password, e := bcrypt.GenerateFromPassword([]byte(newPass), 4)
	if e != nil {
		return e
	}
	return dbfuncs.ChangeUser(&dbfuncs.Users{ID: dbfuncs.FromINT64ToINT(res[0]), Password: string(password)}, "")
}

// restore send on email message code to restore password
func (app *Application) restore(w http.ResponseWriter, r *http.Request) error {
	email := r.PostFormValue("email")
	e := checkEmailAndNick(true, email, "")
	if e != nil {
		return e
	}
	code := StringWithCharset(8)
	app.m.Lock()
	app.RestoreCode[code] = email
	app.m.Unlock()

	mes := "To: " + email + "\nFrom: " + "ani_for@bk.ru" + "\nSubject: Restore password\n\n" +
		"You will be going to restore password on AniFor. \nEnter this code on site: " + code +
		"\nOr visit this: https://" + r.Host + "/sign/r/" + code +
		"\nBe careful this code retire at the end of the day."
	e = SendMail(mes, []string{email})
	if e != nil {
		return errors.New("wrong email")
	}
	return nil
}

// SendMail send msg-mail from -> to
func SendMail(msg string, to []string) error {
	host := "smtp.mail.ru"
	from := "ani_for@bk.ru"
	auth := smtp.PlainAuth("", from, "89f90gMiras", host)
	if e := smtp.SendMail(host+":25", auth, from, to, []byte(msg)); e != nil {
		return e
	}
	return nil
}

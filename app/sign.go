package app

import (
	"errors"
	"net/http"
	"net/smtp"
	"regexp"
	"strings"
	"time"
	"wnet/app/dbfuncs"

	"golang.org/x/crypto/bcrypt"
)

// checkEmailAndNick check if email is empty or not
//	exist = true - user exist in db
func checkEmailAndNick(exist bool, email, nickname string) error {
	results, _ := dbfuncs.GetFrom(dbfuncs.SQLSelectParams{
		What:    "email, nName",
		Table:   "Users",
		Options: dbfuncs.DoSQLOption("email=? OR nName=?", "", "", email, nickname),
	})

	if !exist && len(results) > 0 {
		if results[0][0].(string) == email {
			return errors.New("This email is not empty")
		}
		return errors.New("This nickname is not empty")
	}
	if exist && len(results) == 0 {
		return errors.New("Wrong login")
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
		dbPass, e := dbfuncs.GetOneFrom(dbfuncs.SQLSelectParams{
			What:    "password",
			Table:   "Users",
			Options: dbfuncs.DoSQLOption("email = ? OR nName = ?", "", "", login, login),
		})
		if e != nil {
			return errors.New("Wrong login")
		}
		return bcrypt.CompareHashAndPassword([]byte(dbPass[0].(string)), []byte(pass))
	}
	return nil
}

// finish signUp proccess
func (app *Application) SaveUser(w http.ResponseWriter, r *http.Request) (int, error) {
	user, ok := app.UsersCode[r.PostFormValue("code")]
	if !ok {
		return -1, errors.New("Wrong code")
	}

	ID, e := user.Create()
	if e != nil {
		return -1, e
	}
	return ID, SessionStart(w, r, user.Email, ID)
}

func calculateAgeFromDOB(dob string) (int, error) {
	if dob == "" {
		return 0, errors.New("Do not have date of birth!")
	}

	date, e := time.Parse("2006-01-02", dob)
	if e != nil {
		return 0, errors.New("Invalid date!")
	}

	diff := time.Now().Unix() - date.Unix()
	age := int(diff) / 86400 / 365
	if age < 13 {
		return 0, errors.New("Firstly growth to 13 yearl old!")
	}
	return age, nil
}

// SignUp check validate, start session + oauth2
func (app *Application) SignUp(w http.ResponseWriter, r *http.Request) (map[string]interface{}, error) {
	isOauth2Path := strings.Contains(r.URL.Path, "oauth")
	email := strings.ToLower(r.PostFormValue("email"))
	pass := r.PostFormValue("password")
	lname := r.PostFormValue("lastName")
	fname := r.PostFormValue("firstName")
	dob := r.PostFormValue("dob")
	if isOauth2Path {
		pass = StringWithCharset(8) + "1aA"
		lname = "wnet"
		fname = "user"
		dob = "1999-09-09"
	}
	nickname := lname + "_" + fname + StringWithCharset(8)

	if e := checkEmailAndNick(false, email, nickname); e != nil {
		return nil, e
	}
	if e := checkPassword(false, pass, ""); e != nil {
		return nil, e
	}

	age, e := calculateAgeFromDOB(dob)
	if e != nil {
		return nil, e
	}

	hashPass, e := bcrypt.GenerateFromPassword([]byte(pass), 4)
	if e != nil {
		return nil, errors.New("Password do not saved!")
	}

	// XCSS
	if app.XCSSOther(lname) != nil || app.XCSSOther(fname) != nil {
		return nil, errors.New("It's XSS attack!")
	}

	user := &dbfuncs.User{
		FirstName: fname, LastName: lname, NickName: nickname,
		Gender: "Default", Age: age, Avatar: "/img/default-avatar.png", Dob: dob, About: "",
		Status: "online", IsPrivate: "0", Type: "user",
		Email: email, Password: string(hashPass),
	}
	if !isOauth2Path {
		code := StringWithCharset(8)
		app.m.Lock()
		app.UsersCode[code] = user
		app.m.Unlock()

		mes := "To: " + email + "\nFrom: " + "wnet.soc.net@gmail.com" + "\nSubject: Verification\n\n" +
			"You will be going to register on WNET. \nEnter this code on site: " + code +
			"\nOr visit this: " + r.Header.Get("origin") + "/sign/s/" + code +
			"\nBe careful this code expire today."
		return nil, SendMail(mes, []string{email})
	} else {
		ID, e := user.Create()
		if e != nil {
			return nil, e
		}
		return map[string]interface{}{"id": ID, "password": pass}, SessionStart(w, r, user.Email, ID)
	}
}

// SignIn check password and login from db and request + oauth2
func (app *Application) SignIn(w http.ResponseWriter, r *http.Request) (int, error) {
	isOauth2Path := strings.Contains(r.URL.Path, "oauth")
	email := r.PostFormValue("email")
	pass := r.PostFormValue("password")

	if e := checkEmailAndNick(true, email, email); e != nil {
		return -1, e
	}
	if !isOauth2Path {
		if e := checkPassword(true, pass, email); e != nil {
			return -1, errors.New("Password is not correct!")
		}
	}

	res, e := dbfuncs.GetOneFrom(dbfuncs.SQLSelectParams{
		What:    "id",
		Table:   "Users",
		Options: dbfuncs.DoSQLOption("email = ? OR nName = ?", "", "", email, email),
		Joins:   nil,
	})
	if e != nil {
		return -1, errors.New("Wrong login")
	}
	ID := dbfuncs.FromINT64ToINT(res[0])

	// if app.findUserByID(ID) != nil {
	// 	return -1, errors.New("User already is online!")
	// }
	return ID, SessionStart(w, r, email, ID)
}

// Logout user
func (app *Application) Logout(w http.ResponseWriter, r *http.Request) error {
	cookie, e := r.Cookie(cookieName)
	if e != nil || cookie.Value == "" {
		return e
	}

	if e = dbfuncs.DeleteByParams(dbfuncs.SQLDeleteParams{
		Table:   "Sessions",
		Options: dbfuncs.DoSQLOption("id = '"+cookie.Value+"'", "", ""),
	}); e != nil {
		return errors.New("Not logouted")
	}

	setCookie(w, "", -1)
	return nil
}

// SaveNewPassword restore password
func (app *Application) SaveNewPassword(w http.ResponseWriter, r *http.Request) error {
	email, ok := app.RestoreCode[r.PostFormValue("code")]
	if !ok {
		return errors.New("wrong code")
	}

	newPass := r.PostFormValue("password")
	if e := checkPassword(false, newPass, ""); e != nil {
		return e
	}

	res, e := dbfuncs.GetOneFrom(dbfuncs.SQLSelectParams{
		What:    "id",
		Table:   "Users",
		Options: dbfuncs.DoSQLOption("email = ?", "", "", email),
	})
	if e != nil {
		return errors.New("password do not changed")
	}

	password, e := bcrypt.GenerateFromPassword([]byte(newPass), 4)
	if e != nil {
		return errors.New("the new password do not created")
	}
	user := &dbfuncs.User{ID: dbfuncs.FromINT64ToINT(res[0]), Password: string(password)}
	return user.Change()
}

// ResetPassword send on email message code to reset password
func (app *Application) ResetPassword(w http.ResponseWriter, r *http.Request) error {
	email := r.PostFormValue("email")
	if e := checkEmailAndNick(true, email, ""); e != nil {
		return e
	}

	code := StringWithCharset(8)
	app.m.Lock()
	app.RestoreCode[code] = email
	app.m.Unlock()

	mes := "To: " + email + "\nFrom: " + "wnet.soc.net@gmail.com" + "\nSubject: Restore password\n\n" +
		"You will be going to restore password on WNET. \nEnter this code on site: " + code +
		"\nOr visit this: " + r.Header.Get("origin") + "/sign/rst/" + code +
		"\nBe careful this code retire at the end of the day."
	if e := SendMail(mes, []string{email}); e != nil {
		return errors.New("Wrong email")
	}
	return nil
}

// SendMail send msg-mail from -> to
func SendMail(msg string, to []string) error {
	host := "smtp.gmail.com"
	port := ":25"
	from := "wnet.soc.net@gmail.com"
	pass := "tyhcheejbpzatvha"
	auth := smtp.PlainAuth("", from, pass, host)

	if e := smtp.SendMail(host+port, auth, from, to, []byte(msg)); e != nil {
		return errors.New("Wrong email!")
	}
	return nil
}

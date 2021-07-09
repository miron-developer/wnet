package app

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"wnet/pkg/orm"

	"golang.org/x/crypto/bcrypt"
)

func (app *Application) ConfirmChangeSettings(w http.ResponseWriter, r *http.Request) error {
	userID := getUserIDfromReq(w, r)
	if userID == -1 {
		return errors.New("not logged")
	}

	user, ok := app.ChangeCode[r.PostFormValue("code")]
	if !ok {
		return errors.New("Wrong code")
	}

	if e := user.Change(); e != nil {
		fmt.Println(e)
		return errors.New("not changed user")
	}
	return nil
}

func (app *Application) ChangeSettings(w http.ResponseWriter, r *http.Request) error {
	userID := getUserIDfromReq(w, r)
	if userID == -1 {
		return errors.New("not logged")
	}

	email, pass := r.PostFormValue("email"), r.PostFormValue("password")
	ava, isPrivate := r.PostFormValue("avatar"), r.PostFormValue("isPrivate")

	// XCSS
	if checkAllXSS(email, pass, ava, isPrivate) != nil {
		return errors.New("wrong content")
	}
	if email != "" {
		if e := checkEmailAndNick(false, email, ""); e != nil {
			return e
		}
		if e := SendMail("check email", []string{email}); e != nil {
			return e
		}
	}
	if pass != "" {
		if e := checkPassword(false, pass, ""); e != nil {
			return e
		}
	}

	hashPass, e := bcrypt.GenerateFromPassword([]byte(pass), 4)
	if e != nil {
		return errors.New("pass not created")
	}
	u := &orm.User{
		ID: userID, Avatar: ava,
		Email: email, Password: string(hashPass), IsPrivate: isPrivate,
	}

	em, e := orm.GetOneFrom(orm.SQLSelectParams{
		Table:   "Users",
		What:    "email",
		Options: orm.DoSQLOption("id=?", "", "", userID),
	})
	if e != nil || em[0] == nil {
		return errors.New("not logged")
	}
	currentEmail := em[0].(string)

	code := StringWithCharset(8)
	app.m.Lock()
	app.ChangeCode[code] = u
	app.m.Unlock()

	mes := "To: " + currentEmail + "\nFrom: " + "wnet.soc.net@gmail.com" + "\nSubject: Change data\n\n" +
		"You will be going to change data on WNET. \nEnter this code on site: " + code +
		"\nOr visit this: " + r.Header.Get("origin") + "/profile/settings/s/" + code +
		"\nBe careful this code expire today."
	return SendMail(mes, []string{currentEmail})
}

func (app *Application) ChangeProfile(w http.ResponseWriter, r *http.Request) error {
	if getUserIDfromReq(w, r) == -1 {
		return errors.New("not logged")
	}

	id, e := strconv.Atoi(r.PostFormValue("id"))
	if e != nil {
		return errors.New("wrong id")
	}

	isUser := strings.Contains(r.URL.Path, "user")
	if isUser {
		fName, lName, nName := r.PostFormValue("firstName"), r.PostFormValue("lastName"), r.PostFormValue("nickname")
		dob, about, gender := r.PostFormValue("dob"), r.PostFormValue("aboutMe"), r.PostFormValue("gender")
		if checkAllXSS(fName, lName, nName, dob, about, gender) != nil {
			return errors.New("wrong content")
		}

		u := &orm.User{
			ID:  id,
			Dob: dob, Gender: gender, About: about,
			FirstName: fName, LastName: lName, NickName: nName,
		}
		return u.Change()
	}
	title, description := r.PostFormValue("title"), r.PostFormValue("description")
	if checkAllXSS(title, description) != nil {
		return errors.New("wrong content")
	}

	g := &orm.Group{
		ID: id, Title: title, About: description,
	}
	return g.Change()
}

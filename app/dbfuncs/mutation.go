/*
	This file define "graphql" mutation
*/

package dbfuncs

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strconv"
)

// ---------------------Create funcs---------------------------

// CreateUser create one user
func CreateUser(user *Users) (int, error) {
	if user.NickName == "" || user.Password == "" || user.Email == "" {
		return -1, errors.New("n/d")
	}
	r, e := insert("Users", "null,?,?,?,?,?,?,?,?,?,?,?,?", makeArrFromStruct(*user)[1:])
	if e != nil {
		return -1, e
	}
	ID, _ := r.LastInsertId()
	return int(ID), e
}

// CreateSession create new session in db
func CreateSession(ses *Sessions) error {
	if ses.ID == "" {
		return errors.New("n/d")
	}
	_, e := insert("Sessions", "?,?,?", makeArrFromStruct(*ses))
	return e
}

// CreatePost one post and return it's ID
func CreatePost(p *Posts) (int, error) {
	if p.Title == "" || p.Body == "" || p.Tags == "" {
		return -1, errors.New("n/d")
	}
	res, e := insert("Posts", "null,?,?,?,?,?,?,?", makeArrFromStruct(*p)[1:])
	if e != nil {
		return -1, e
	}
	ID, e := res.LastInsertId()
	return int(ID), e
}

// CreateComment create one comment
func CreateComment(c *Comments) (int, error) {
	if c.Body == "" {
		return -1, errors.New("n/d")
	}
	var res sql.Result
	var e error
	arr := makeArrFromStruct(*c)[1:]
	if c.PostID != 0 {
		res, e = insert("Comments", "null,?,?,?,?,?,?,?,null", arr[:len(arr)-1])
	} else {
		arr = arr[:len(arr)-2]
		arr = append(arr, c.CommentID)
		res, e = insert("Comments", "null,?,?,?,?,?,?,null,?", arr)
	}
	if e != nil {
		return -1, e
	}
	ID, e := res.LastInsertId()
	return int(ID), e
}

// SetLD upsert LD by userID and whosID
func SetLD(LD *LDGeneral, collName string) error {
	if LD.TableName == "" || collName == "" {
		return errors.New("n/d")
	}
	op := DoSQLOption("userID=? AND "+collName+"=?", "", "", LD.UserID, LD.WhosID)
	update(LD.TableName, "value='"+strconv.Itoa(LD.Value)+"'", op)
	arr := makeArrFromStruct(*LD)
	op.Where = "(SELECT Changes() = 0)"
	_, e := insertBySelect(LD.TableName, arr[1:len(arr)-1], op)
	return e
}

// CreateMessage create one Message
func CreateMessage(msg *Messages) error {
	if msg.Body == "" {
		return errors.New("n/d")
	}
	_, e := insert("Messages", "null,?,?,?,?,?", makeArrFromStruct(*msg)[1:])
	return e
}

// ---------------------Change funcs---------------------------

// ChangeUser change user profile
func ChangeUser(u *Users, oldphoto string) error {
	if u.ID == 0 {
		return errors.New("ID absent")
	}
	set := ""
	if u.NickName != "" {
		set += " nickname = '" + u.NickName + "',"
	}
	if u.FirstName != "" {
		set += " firstName = '" + u.FirstName + "',"
	}
	if u.LastName != "" {
		set += " lastName = '" + u.LastName + "',"
	}
	if u.Photo != "" {
		if oldphoto != "static/img/avatar/default.png" {
			os.Remove(oldphoto)
		}
		set += " photo = '" + u.Photo + "',"
	}
	if u.Role != "" {
		set += " role = '" + u.Role + "',"
	}
	if u.Friends != "" {
		set += " friends = '" + u.Friends + "',"
	}
	if u.Password != "" {
		set += " password = '" + u.Password + "',"
	}
	if u.LastActivitie != "" {
		set += " lastActivitie = '" + u.LastActivitie + "'"
	}
	_, e := update("Users", set[:len(set)-1], DoSQLOption("ID = ?", "", "", u.ID))
	return e
}

// ChangeSession change expiration
func ChangeSession(s *Sessions) error {
	if s.ID == "" || s.Expire == "" {
		return errors.New("expiration time or ID absent")
	}
	_, e := update("Sessions", "expire = '"+s.Expire+"'", DoSQLOption("ID = ?", "", "", s.ID))
	return e
}

// ChangePost one post
func ChangePost(p *Posts) error {
	if p.ID == 0 {
		return errors.New("ID absent")
	}
	set := ""
	if p.Title != "" {
		set += " title = '" + p.Title + "',"
	}
	if p.Body != "" {
		set += " body = '" + p.Body + "',"
	}
	if p.Tags != "" {
		set += " tags = '" + p.Tags + "',"
	}
	set += " date = '" + p.Date + "',"
	set += " changed = 1"
	_, e := update("Posts", set, DoSQLOption("ID = ?", "", "", p.ID))
	return e
}

// ChangeCommentNestedState change nested state
func ChangeCommentNestedState(ID int) error {
	if ID < 0 {
		return errors.New("wrong ID")
	}
	_, e := update("Comments", "haveChild = 1", DoSQLOption("ID = ?", "", "", ID))
	return e
}

// ChangeComment change setted comment
func ChangeComment(com *Comments) error {
	if com.ID == 0 || com.Body == "" || com.Date == "" {
		return errors.New("ID absent or nothing to change")
	}
	set := fmt.Sprintf("body = '%v', date = '%v', changed = 1", com.Body, com.Date)
	_, e := update("Comments", set, DoSQLOption("ID = ?", "", "", com.ID))
	return e
}

// ChangeCarma change carma by tablename, and ID
func ChangeCarma(tableName, ID, newCarma string) error {
	_, e := update(tableName, "carma = "+newCarma, DoSQLOption("ID = ?", "", "", ID))
	return e
}

// ---------------------Delete funcs---------------------------

// DeleteSession ...
func DeleteSession(condition string) error {
	_, e := deleteSQL("Sessions", DoSQLOption(condition, "", ""))
	return e
}

// DeletePost delete one post by id
func DeletePost(condition, photo string) error {
	if photo != "static/img/avatar/default.png" {
		os.Remove(photo)
	}
	_, e := deleteSQL("Posts", DoSQLOption(condition, "", ""))
	return e
}

// DeleteComment delete one comment by id
func DeleteComment(condition string) error {
	_, e := deleteSQL("Comments", DoSQLOption(condition, "", ""))
	return e
}

// DeleteMessage delete one notifie
func DeleteMessage(condition string) error {
	_, e := deleteSQL("Messages", DoSQLOption(condition, "", ""))
	return e
}

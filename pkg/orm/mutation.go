package orm

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// if have choise whom belong creation data, then this func prepare needed options
// 	for ex post creation. post may belong to user or group.
// 	if to user: null,?,?,?,?,?,null
// 	if to group: null,?,?,?,?,null,?
// 	default = first choose
func prepareDataAndValues(datas string, values []interface{}, options []int) (string, []interface{}) {
	valuesCount := len(values)
	optionsCount := len(options)
	FKsFromIndex := valuesCount - optionsCount

	values = values[:FKsFromIndex]
	arrFromDatas := strings.Split(datas, ",")
	for i := FKsFromIndex; i < len(arrFromDatas); i++ {
		arrFromDatas[i] = "null"
	}

	for i, v := range options {
		if v != 0 {
			arrFromDatas[i+FKsFromIndex] = "?"
			values = append(values, options[i])
		}
	}

	datas = strings.Join(arrFromDatas, ",")
	return datas, values
}

// ---------------------Create funcs---------------------------

// Create create one user
func (user *User) Create() (int, error) {
	if user.NickName == "" || user.Password == "" || user.Email == "" {
		return -1, errors.New("n/d")
	}

	r, e := insertSQL(SQLInsertParams{
		Table:  "Users",
		Datas:  "null,?,?,?,?,?,?,?,?,?,?,?,?,?",
		Values: MakeArrFromStruct(*user)[1:],
	})
	if e != nil {
		return -1, e
	}
	ID, _ := r.LastInsertId()
	return int(ID), e
}

// Create create one group
func (g *Group) Create() (int, error) {
	if g.Title == "" || g.OwnerUserID == 0 {
		return -1, errors.New("n/d")
	}

	r, e := insertSQL(SQLInsertParams{
		Table:  "Groups",
		Datas:  "null,?,?,?,?,?,?,?,?",
		Values: MakeArrFromStruct(*g)[1:],
	})
	if e != nil {
		return -1, e
	}
	ID, _ := r.LastInsertId()
	return int(ID), e
}

// Create create new session in db
func (ses *Session) Create() error {
	if ses.ID == "" || ses.UserID == 0 || ses.Expire == "" {
		return errors.New("n/d")
	}

	_, e := insertSQL(SQLInsertParams{
		Table:  "Sessions",
		Datas:  "?,?,?",
		Values: MakeArrFromStruct(*ses),
	})
	return e
}

// Create one post and return it's ID
func (p *Post) Create() (int, error) {
	if p.Title == "" || p.Body == "" || p.UnixDate == 0 || p.PostType == "" {
		return -1, errors.New("n/d")
	}

	params := SQLInsertParams{
		Table:  "Posts",
		Datas:  "null,?,?,?,?,?,?,?,?,?",
		Values: MakeArrFromStruct(*p),
	}
	params.Datas, params.Values = prepareDataAndValues(params.Datas, params.Values, []int{p.UserID, p.GroupID})
	params.Values = params.Values[1:]

	res, e := insertSQL(params)
	if e != nil {
		return -1, e
	}
	ID, e := res.LastInsertId()
	return int(ID), e
}

// Create create one Message
func (msg *Message) Create() (int, error) {
	if (msg.Body == "" && msg.MessageType == "text") || msg.UnixDate == 0 || msg.MessageType == "" || msg.SenderUserID == 0 {
		return -1, errors.New("n/d")
	}

	params := SQLInsertParams{
		Table:  "Messages",
		Datas:  "null,?,?,?,?,?,?",
		Values: MakeArrFromStruct(*msg),
	}
	params.Datas, params.Values = prepareDataAndValues(params.Datas, params.Values, []int{msg.ReceiverUserID, msg.ReceiverGroupID})
	params.Values = params.Values[1:]

	res, e := insertSQL(params)
	if e != nil {
		return -1, e
	}
	ID, e := res.LastInsertId()
	return int(ID), e
}

// Create create one Chat
func (c *Chat) Create() (int, error) {
	if c.ReceiverGroupID == 0 && c.ReceiverUserID == 0 {
		return -1, errors.New("n/d")
	}

	op := DoSQLOption(
		"(users = ? OR users = ?) AND type = 'user'",
		"",
		"",
		fmt.Sprintf("|%v |%v ", c.SenderUserID, c.ReceiverUserID), fmt.Sprintf("|%v |%v ", c.ReceiverUserID, c.SenderUserID),
	)
	if c.ChatType == "group" {
		op.Where = "users LIKE '%|" + strconv.Itoa(c.SenderUserID) + " %' AND receiverGroupID = ?"
		op.Args = []interface{}{c.ReceiverGroupID}
	}

	if data, e := GetOneFrom(SQLSelectParams{
		Table:   "Chats",
		What:    "id, closed",
		Options: op,
	}); e == nil && data != nil {
		closeds := data[1].(string)
		rg := regexp.MustCompile(`[|\d]+[` + strconv.Itoa(c.SenderUserID) + `]+\s`)
		if rg.MatchString(closeds) {
			closeds = rg.ReplaceAllString(closeds, "")
			return FromINT64ToINT(data[0]), c.Change()
		}
		return FromINT64ToINT(data[0]), nil
	}

	if c.ChatType == "group" {
		data, e := GetFrom(SQLSelectParams{
			Table:   "Users AS u",
			What:    "u.id",
			Options: DoSQLOption("", "", ""),
			Joins:   []SQLJoin{DoSQLJoin(INJOINQ, "Relations AS r", "r.senderUserID = u.id AND r.receiverGroupID = ?", c.ReceiverGroupID)},
		})
		if data == nil && e != nil {
			return -1, e
		}
		c.Users = ""
		for _, v := range data {
			c.Users += fmt.Sprint("|", v[0], " ")
		}
	}

	params := SQLInsertParams{
		Table:  "Chats",
		Datas:  "null,?,?,?,?,?,?",
		Values: MakeArrFromStruct(*c),
	}
	params.Datas, params.Values = prepareDataAndValues(params.Datas, params.Values, []int{c.ReceiverUserID, c.ReceiverGroupID})
	params.Values = params.Values[1:]

	res, e := insertSQL(params)
	if e != nil {
		return -1, e
	}
	ID, e := res.LastInsertId()
	return int(ID), e
}

// Create create one Events
func (evnt *Event) Create() (int, error) {
	if evnt.Title == "" || evnt.UnixDate == 0 || evnt.About == "" {
		return -1, errors.New("n/d")
	}

	params := SQLInsertParams{
		Table:  "Events",
		Datas:  "null,?,?,?,?,?,?",
		Values: MakeArrFromStruct(*evnt),
	}
	params.Datas, params.Values = prepareDataAndValues(params.Datas, params.Values, []int{evnt.UserID, evnt.GroupID})
	params.Values = params.Values[1:]

	res, e := insertSQL(params)
	if e != nil {
		return -1, e
	}
	ID, e := res.LastInsertId()
	return int(ID), e
}

// Create create one EventAnswers
func (ea *EventAnswer) Create() error {
	if ea.Answer == "" || ea.UserID == 0 || ea.EventID == 0 {
		return errors.New("n/d")
	}

	_, e := insertSQL(SQLInsertParams{
		Table:  "EventAnswers",
		Datas:  "null,?,?,?",
		Values: MakeArrFromStruct(*ea)[1:],
	})
	return e
}

// Create create one Relations
func (r *Relation) Create() (int, error) {
	if r.Value == "" {
		return -1, errors.New("n/d")
	}

	params := SQLInsertParams{
		Table:  "Relations",
		Datas:  "null,?,?,?,?,?",
		Values: MakeArrFromStruct(*r),
	}
	params.Datas, params.Values = prepareDataAndValues(params.Datas, params.Values, []int{r.SenderUserID, r.SenderGroupID, r.ReceiverUserID, r.ReceiverGroupID})
	params.Values = params.Values[1:]

	res, e := insertSQL(params)
	if e != nil {
		return -1, e
	}
	ID, e := res.LastInsertId()
	return int(ID), e
}

// Create create one Media(photo&video)
func (m *Media) Create() (int, error) {
	if m.Title == "" || m.MediaType == "" || m.UnixDate == 0 || m.Source == "" {
		return -1, errors.New("n/d")
	}

	params := SQLInsertParams{
		Table:  "Media",
		Datas:  "null,?,?,?,?,?,?,?",
		Values: MakeArrFromStruct(*m),
	}
	params.Datas, params.Values = prepareDataAndValues(params.Datas, params.Values, []int{m.UserID, m.GroupID})
	params.Values = params.Values[1:]

	res, e := insertSQL(params)
	if e != nil {
		return -1, e
	}
	ID, e := res.LastInsertId()
	return int(ID), e
}

// Create create one comment
func (c *Comment) Create() (int, error) {
	if c.Body == "" || c.UnixDate == 0 || c.UserID == 0 {
		return -1, errors.New("n/d")
	}

	params := SQLInsertParams{
		Table:  "Comments",
		Datas:  "null,?,?,?,?,?,?,?,?,?",
		Values: MakeArrFromStruct(*c),
	}
	params.Datas, params.Values = prepareDataAndValues(params.Datas, params.Values, []int{c.PostID, c.CommentID, c.MediaID})
	params.Values = params.Values[1:]

	res, e := insertSQL(params)
	if e != nil {
		return -1, e
	}
	ID, e := res.LastInsertId()
	return int(ID), e
}

// Create create one notification
func (n *Notification) Create() (int, error) {
	if n.NotificationType == "" || n.UnixDate == 0 || n.SenderUserID == 0 {
		return -1, errors.New("n/d")
	}

	params := SQLInsertParams{
		Table:  "Notifications",
		Datas:  "null,?,?,?,?,?,?,?,?,?,?",
		Values: MakeArrFromStruct(*n),
	}
	if n.NotificationType == "6" {
		params.Datas = "null,?,?,?,?,?,null,null,null,?,null"
		params.Values = append(params.Values[1:6], n.GroupID)
	} else {
		if n.ReceiverUserID == 0 {
			params.Datas = "null,?,?,?,null,?,?,?,?,?,?"
		}
		params.Datas, params.Values = prepareDataAndValues(
			params.Datas,
			params.Values,
			[]int{n.RelationID, n.PostID, n.CommentID, n.EventID, n.GroupID, n.MediaID},
		)
		params.Values = params.Values[1:]
		if n.ReceiverUserID == 0 {
			params.Values = append(params.Values[:3], params.Values[4:]...)
		}
	}

	res, e := insertSQL(params)
	if e != nil {
		return -1, e
	}
	ID, e := res.LastInsertId()
	return int(ID), e
}

// Set remove like or create
func (l *Like) Set() (bool, error) {
	if l.UserID == 0 && l.CommentID+l.MediaID+l.PostID == 0 {
		return false, errors.New("n/d")
	}

	deleteParams := SQLDeleteParams{
		Table:   "Likes",
		Options: DoSQLOption("userID=? AND ", "", "", l.UserID),
	}

	if l.PostID != 0 {
		deleteParams.Options.Where += "postID=?"
		deleteParams.Options.Args = append(deleteParams.Options.Args, l.PostID)
	} else if l.CommentID != 0 {
		deleteParams.Options.Where += "commentID=?"
		deleteParams.Options.Args = append(deleteParams.Options.Args, l.CommentID)
	} else {
		deleteParams.Options.Where += "mediaID=?"
		deleteParams.Options.Args = append(deleteParams.Options.Args, l.MediaID)
	}
	res, _ := deleteSQL(deleteParams)

	if affected, e := res.RowsAffected(); affected == 0 && e == nil {
		insertParams := SQLInsertParams{
			Table:  "Likes",
			Datas:  "null,?,?,?,?",
			Values: MakeArrFromStruct(*l),
		}
		insertParams.Datas, insertParams.Values = prepareDataAndValues(insertParams.Datas, insertParams.Values, []int{l.PostID, l.CommentID, l.MediaID})
		insertParams.Values = insertParams.Values[1:]

		_, e := insertSQL(insertParams)
		return true, e
	}
	return false, nil
}

// Create create one file
func (f *ClippedFile) Create() (int, error) {
	if f.FileType == "" || f.UserID == 0 {
		return -1, errors.New("n/d")
	}

	params := SQLInsertParams{
		Table:  "Files",
		Datas:  "null,?,?,?,?,?,?,?",
		Values: MakeArrFromStruct(*f),
	}
	params.Datas, params.Values = prepareDataAndValues(params.Datas, params.Values, []int{f.PostID, f.CommentID, f.MessageID})
	params.Values = params.Values[1:]

	res, e := insertSQL(params)
	if e != nil {
		return -1, e
	}
	ID, e := res.LastInsertId()
	return int(ID), e
}

// ---------------------Change funcs---------------------------

// Change change user profile
func (u *User) Change() error {
	if u.ID == 0 {
		return errors.New("absent/d")
	}

	params := SQLUpdateParams{
		Table:   "Users",
		Couples: map[string]string{},
		Options: DoSQLOption("id=?", "", "", u.ID),
	}

	if u.Email != "" {
		params.Couples["email"] = u.Email
	}
	if u.NickName != "" {
		params.Couples["nName"] = u.NickName
	}
	if u.FirstName != "" {
		params.Couples["fName"] = u.FirstName
	}
	if u.LastName != "" {
		params.Couples["lName"] = u.LastName
	}
	if u.Avatar != "" {
		params.Couples["ava"] = u.Avatar
	}
	if u.IsPrivate != "" {
		params.Couples["isPrivate"] = u.IsPrivate
	}
	if u.Password != "" {
		params.Couples["password"] = u.Password
	}
	if u.Status != "" {
		params.Couples["status"] = u.Status
	}
	if u.Age > 0 {
		params.Couples["age"] = strconv.Itoa(u.Age)
	}
	if u.About != "" {
		params.Couples["about"] = u.About
	}
	if u.Gender != "" {
		params.Couples["gender"] = u.Gender
	}

	_, e := updateSQL(params)
	return e
}

// Change change group profile
func (g *Group) Change() error {
	if g.ID == 0 {
		return errors.New("absent/d")
	}

	params := SQLUpdateParams{
		Table:   "Groups",
		Couples: map[string]string{},
		Options: DoSQLOption("id=?", "", "", g.ID),
	}

	if g.Title != "" {
		params.Couples["title"] = g.Title
	}
	if g.Avatar != "" {
		params.Couples["ava"] = g.Avatar
	}
	if g.IsPrivate != "" {
		params.Couples["isPrivate"] = g.IsPrivate
	}
	if g.Age > 0 {
		params.Couples["age"] = strconv.Itoa(g.Age)
	}
	if g.About != "" {
		params.Couples["about"] = g.About
	}

	_, e := updateSQL(params)
	return e
}

// Change change expiration
func (s *Session) Change() error {
	if s.ID == "" || s.Expire == "" {
		return errors.New("absent/d")
	}

	_, e := updateSQL(SQLUpdateParams{
		Table:   "Sessions",
		Couples: map[string]string{"expire": s.Expire},
		Options: DoSQLOption("id=?", "", "", s.ID),
	})
	return e
}

// Change change post
func (p *Post) Change() error {
	if p.ID == 0 {
		return errors.New("absent/d")
	}

	params := SQLUpdateParams{
		Table:   "Posts",
		Couples: map[string]string{},
		Options: DoSQLOption("id=?", "", "", p.ID),
	}
	if p.Title != "" {
		params.Couples["title"] = p.Title
	}
	if p.Body != "" {
		params.Couples["body"] = p.Body
	}
	if p.PostType != "" {
		params.Couples["type"] = p.PostType
	}
	if p.IsHaveClippedFiles != "" {
		params.Couples["isHaveClippedFiles"] = p.IsHaveClippedFiles
	}

	_, e := updateSQL(params)
	return e
}

// Change change chat
func (c *Chat) Change() error {
	if c.ID == 0 {
		return errors.New("absent/d")
	}

	params := SQLUpdateParams{
		Table:   "Chats",
		Couples: map[string]string{},
		Options: DoSQLOption("id=?", "", "", c.ID),
	}

	if c.Closed != "" {
		params.Couples["closed"] = c.Closed
	}
	if c.Users != "" {
		params.Couples["users"] = c.Users
	}

	_, e := updateSQL(params)
	return e
}

// Change change message
func (msg *Message) Change() error {
	if msg.ID == 0 || msg.Body == "" {
		return errors.New("absent/d")
	}

	_, e := updateSQL(SQLUpdateParams{
		Table:   "Messages",
		Couples: map[string]string{"body": msg.Body},
		Options: DoSQLOption("id=?", "", "", msg.ID),
	})
	return e
}

// Change one event
func (evnt *Event) Change() error {
	if evnt.ID == 0 {
		return errors.New("absent/d")
	}

	params := SQLUpdateParams{
		Table:   "Events",
		Couples: map[string]string{},
		Options: DoSQLOption("id=?", "", "", evnt.ID),
	}
	if evnt.Title != "" {
		params.Couples["title"] = evnt.Title
	}
	if evnt.About != "" {
		params.Couples["about"] = evnt.About
	}

	_, e := updateSQL(params)
	return e
}

// Change change relation
func (r *Relation) Change() error {
	if r.ID == 0 || r.Value == "" {
		return errors.New("absent/d")
	}

	_, e := updateSQL(SQLUpdateParams{
		Table:   "Relations",
		Couples: map[string]string{"value": r.Value},
		Options: DoSQLOption("id=?", "", "", r.ID),
	})
	return e
}

// Change change media
func (m *Media) Change() error {
	if m.ID == 0 || m.Title == "" {
		return errors.New("absent/d")
	}

	_, e := updateSQL(SQLUpdateParams{
		Table:   "Media",
		Couples: map[string]string{"title": m.Title},
		Options: DoSQLOption("id=?", "", "", m.ID),
	})
	return e
}

// Change change comment
func (c *Comment) Change() error {
	if c.ID == 0 {
		return errors.New("absent/d")
	}

	params := SQLUpdateParams{
		Table:   "Comments",
		Couples: map[string]string{},
		Options: DoSQLOption("id=?", "", "", c.ID),
	}

	if c.Body != "" {
		params.Couples["body"] = c.Body
	}
	if c.IsHaveChild != "" {
		params.Couples["isHaveChild"] = c.IsHaveChild
	}
	if c.IsHaveClippedFiles != "" {
		params.Couples["isHaveClippedFiles"] = c.IsHaveClippedFiles
	}

	_, e := updateSQL(params)
	return e
}

// ---------------------Delete funcs---------------------------

// DeleteByParams delete one by id
func DeleteByParams(params SQLDeleteParams) error {
	_, e := deleteSQL(params)
	return e
}

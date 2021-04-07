package dbfuncs

import (
	"errors"
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

	repeat := func(count int) []string {
		res := make([]string, count)
		for i := range res {
			res[i] = "null"
		}
		return res
	}

	choose := func(index int) (string, []interface{}) {
		arrFromDatas := strings.Split(datas, ",")
		arrFromDatas = arrFromDatas[:valuesCount-optionsCount]
		if index > 0 {
			arrFromDatas = append(arrFromDatas, repeat(index)...)
		}
		arrFromDatas = append(arrFromDatas, "?")
		if optionsCount-index-1 > 0 {
			arrFromDatas = append(arrFromDatas, repeat(optionsCount-index-1)...)
		}
		datas = strings.Join(arrFromDatas, ",")
		values = append(values[:valuesCount-optionsCount], options[index])
		return datas, values
	}

	for i, v := range options {
		if v != 0 {
			return choose(i)
		}
	}
	return choose(0)
}

// ---------------------Create funcs---------------------------

// CreateUser create one user
func CreateUser(user *Users) (int, error) {
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

// CreateGroup create one group
func CreateGroup(g *Groups) (int, error) {
	if g.Title == "" || g.OwnerUserID == 0 {
		return -1, errors.New("n/d")
	}

	r, e := insertSQL(SQLInsertParams{
		Table:  "Users",
		Datas:  "null,?,?,?,?,?,?,?,?",
		Values: MakeArrFromStruct(*g)[1:],
	})
	if e != nil {
		return -1, e
	}
	ID, _ := r.LastInsertId()
	return int(ID), e
}

// CreateSession create new session in db
func CreateSession(ses *Sessions) error {
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

// CreatePost one post and return it's ID
func CreatePost(p *Posts) (int, error) {
	if p.Title == "" || p.Body == "" || p.UnixDate == 0 || p.PostType == "" {
		return -1, errors.New("n/d")
	}

	params := SQLInsertParams{
		Table:  "Posts",
		Datas:  "null,?,?,?,?,?,?,?",
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

// CreateMessage create one Message
func CreateMessage(msg *Messages) error {
	if msg.Body == "" || msg.UnixDate == 0 || msg.MessageType == "" || msg.SenderUserID == 0 {
		return errors.New("n/d")
	}

	params := SQLInsertParams{
		Table:  "Messages",
		Datas:  "null,?,?,?,?,?,?",
		Values: MakeArrFromStruct(*msg),
	}
	params.Datas, params.Values = prepareDataAndValues(params.Datas, params.Values, []int{msg.ReceiverUserID, msg.ReceiverGroupID})
	params.Values = params.Values[1:]

	_, e := insertSQL(params)
	return e
}

// CreateEvent create one Events
func CreateEvent(evnt *Events) error {
	if evnt.Title == "" || evnt.UnixDate == 0 || evnt.About == "" {
		return errors.New("n/d")
	}

	params := SQLInsertParams{
		Table:  "Events",
		Datas:  "null,?,?,?,?,?,?",
		Values: MakeArrFromStruct(*evnt),
	}
	params.Datas, params.Values = prepareDataAndValues(params.Datas, params.Values, []int{evnt.UserID, evnt.GroupID})
	params.Values = params.Values[1:]

	_, e := insertSQL(params)
	return e
}

// CreateEventAnswer create one EventAnswers
func CreateEventAnswer(ea *EventAnswers) error {
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

// CreateRelation create one Relations
func CreateRelation(r *Relations) error {
	if r.Value == "" || r.ReceiverUserID == 0 || r.SenderUserID == 0 {
		return errors.New("n/d")
	}

	_, e := insertSQL(SQLInsertParams{
		Table:  "Relations",
		Datas:  "null,?,?,?",
		Values: MakeArrFromStruct(*r)[1:],
	})
	return e
}

// CreateMedia create one Media(photo&video)
func CreateMedia(m *Media) error {
	if m.Title == "" || m.MediaType == "" || m.UnixDate == 0 || m.Source == "" {
		return errors.New("n/d")
	}

	params := SQLInsertParams{
		Table:  "Media",
		Datas:  "null,?,?,?,?,?,?",
		Values: MakeArrFromStruct(*m),
	}
	params.Datas, params.Values = prepareDataAndValues(params.Datas, params.Values, []int{m.UserID, m.GroupID})
	params.Values = params.Values[1:]

	_, e := insertSQL(params)
	return e
}

// CreateComment create one comment
func CreateComment(c *Comments) (int, error) {
	if c.Body == "" || c.UnixDate == 0 || c.UserID == 0 {
		return -1, errors.New("n/d")
	}

	params := SQLInsertParams{
		Table:  "Comments",
		Datas:  "null,?,?,?,?,?,?,?,?",
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

// CreateNotification create one notification
//  if isForAll = true, this notification for all followers
//  else receiverUserID = null
func CreateNotification(n *Notifications, isForAll bool) (int, error) {
	if n.NotificationType == "" || n.UnixDate == 0 || n.SenderUserID == 0 {
		return -1, errors.New("n/d")
	}

	params := SQLInsertParams{
		Table:  "Notifications",
		Datas:  "null,?,?,?,?,?,?,?,?,?",
		Values: MakeArrFromStruct(*n),
	}
	if isForAll {
		params.Datas = "null,?,?,?,null,?,?,?,?,?"
	}
	params.Datas, params.Values = prepareDataAndValues(
		params.Datas,
		params.Values,
		[]int{n.PostID, n.CommentID, n.EventID, n.GroupID, n.MediaID},
	)
	params.Values = params.Values[1:]

	res, e := insertSQL(params)
	if e != nil {
		return -1, e
	}
	ID, e := res.LastInsertId()
	return int(ID), e
}

// SetLikes remove like or create
func SetLikes(l *Likes) (bool, error) {
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

// CreateClippedFile create one file
func CreateClippedFile(f *ClippedFiles) (int, error) {
	if f.FileType == "" || f.UserID == 0 {
		return -1, errors.New("n/d")
	}

	params := SQLInsertParams{
		Table:  "Files",
		Datas:  "null,?,?,?,?,?",
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

// ChangeUser change user profile
func ChangeUser(u *Users) error {
	if u.ID == 0 {
		return errors.New("absent/d")
	}

	params := SQLUpdateParams{
		Table:   "Users",
		Couples: map[string]string{},
		Options: DoSQLOption("id=?", "", "", u.ID),
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

// ChangeGroup change group profile
func ChangeGroup(g *Groups) error {
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

// ChangeSession change expiration
func ChangeSession(s *Sessions) error {
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

// ChangePost change post
func ChangePost(p *Posts) error {
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

	_, e := updateSQL(params)
	return e
}

// ChangeMessage change message
func ChangeMessage(msg *Messages) error {
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

// ChangePost one post
func ChangeEvent(evnt *Events) error {
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

// ChangeRelation change relation
func ChangeRelation(r *Relations) error {
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

// ChangeMedia change media
func ChangeMedia(m *Media) error {
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

// ChangeComment change comment
func ChangeComment(c *Comments) error {
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

	_, e := updateSQL(params)
	return e
}

// ---------------------Delete funcs---------------------------

// DeleteByParams delete one by id
func DeleteByParams(params SQLDeleteParams) error {
	_, e := deleteSQL(params)
	return e
}

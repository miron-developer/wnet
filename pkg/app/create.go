package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
	"wnet/pkg/orm"
)

func checkAllXSS(testers ...string) error {
	for _, v := range testers {
		if e := XCSSOther(v); e != nil {
			return e
		}
	}
	return nil
}

func setObjIDByType(t string, id int, obj interface{}) {
	tmpMap := map[string]interface{}{}
	js, _ := json.Marshal(obj)
	json.Unmarshal(js, &tmpMap)

	chooseType := func() {
		if t == "post" {
			tmpMap["postID"] = id
			return
		}
		if t == "comment" {
			tmpMap["commentID"] = id
			return
		}
		if t == "message" {
			tmpMap["messageID"] = id
		}
	}

	chooseType()
	js, _ = json.Marshal(tmpMap)
	json.Unmarshal(js, &obj)
}

func (app *Application) CreatePost(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	userID := getUserIDfromReq(w, r)
	if userID == -1 {
		return nil, errors.New("not logged")
	}

	title, body := r.PostFormValue("title"), r.PostFormValue("body")
	postType, whichPost := r.PostFormValue("postType"), r.PostFormValue("which")
	allowedUsers, isHaveClippedFiles := r.PostFormValue("choosenFollowers"), r.PostFormValue("isHaveClippedFiles")

	// XCSS
	if checkAllXSS(title, body, postType, allowedUsers) != nil {
		return nil, errors.New("wrong content")
	}

	ids := []int{userID}
	if whichPost == "group" {
		ids = []int{}
		for _, groupID := range strings.Split(r.PostFormValue("choosenGroups"), ",") {
			if id, e := strconv.Atoi(groupID); e == nil {
				ids = append(ids, id)
			}
		}
	}

	datas := []int{}
	for _, id := range ids {
		p := &orm.Post{
			Title: title, Body: body, PostType: postType, Type: "post",
			UnixDate: int(time.Now().Unix() * 1000), AllowedUsers: allowedUsers, IsHaveClippedFiles: isHaveClippedFiles,
		}
		if whichPost == "group" {
			p.GroupID = id
		} else {
			p.UserID = id
		}

		postID, e := p.Create()
		if e != nil {
			return nil, errors.New("not create post")
		}

		// create notification
		go func() {
			n := orm.Notification{
				UnixDate: p.UnixDate, NotificationType: "1",
				SenderUserID: userID, PostID: postID,
			}

			if nid, e := n.Create(); e == nil {
				app.findUserByID(userID).Write(app, &WSMessage{
					MsgType:    WSM_ADD_USER_NOTIFICATION_TYPE,
					ReceiverID: "-2",
					Body: struct {
						ID    int    `json:"id"`
						NType string `json:"type"`
					}{ID: nid, NType: n.NotificationType},
				})
			}
		}()
		datas = append(datas, postID)
	}
	return datas, nil
}

func (app *Application) CreateGroup(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	userID := getUserIDfromReq(w, r)
	if userID == -1 {
		return -1, errors.New("not logged")
	}

	title, description := r.PostFormValue("title"), r.PostFormValue("description")
	groupType, ava := r.PostFormValue("groupType"), r.PostFormValue("avatar")
	inviteUsers := r.PostFormValue("choosenFollowers")
	if ava == "" {
		ava = "/img/default-avatar.png"
	}

	// XCSS
	if checkAllXSS(title, description, groupType, inviteUsers, ava) != nil {
		return -1, errors.New("wrong content")
	}

	g := &orm.Group{
		Title: title, About: description, Type: "group",
		CreationDate: strings.Split(TimeExpire(time.Nanosecond), " ")[0], Age: 0,
		Avatar: ava, IsPrivate: groupType, OwnerUserID: userID,
	}
	id, e := g.Create()
	if e != nil {
		return -1, errors.New("not create group")
	}

	rl := &orm.Relation{
		Value:           "1",
		SenderUserID:    userID,
		ReceiverGroupID: id,
	}
	if _, e := rl.Create(); e != nil {
		orm.DeleteByParams(orm.SQLDeleteParams{Table: "Groups", Options: orm.DoSQLOption("id=?", "", "", id)})
		return -1, errors.New("not create group: no relation between user & group")
	}

	// create notification
	go func() {
		n := orm.Notification{
			UnixDate: int(time.Now().Unix() * 1000), NotificationType: "3",
			SenderUserID: userID, GroupID: id,
		}
		if nid, e := n.Create(); e == nil {
			app.findUserByID(userID).Write(app, &WSMessage{
				MsgType:    WSM_ADD_USER_NOTIFICATION_TYPE,
				ReceiverID: "-2",
				Body: struct {
					ID    int    `json:"id"`
					NType string `json:"type"`
				}{ID: nid, NType: n.NotificationType},
			})
		}
	}()

	if inviteUsers != "" {
		for _, inviterStringID := range strings.Split(inviteUsers, ",") {
			inviterID, e := strconv.Atoi(inviterStringID)
			if e != nil {
				continue
			}

			rlsh := &orm.Relation{
				Value:          "-1",
				SenderGroupID:  id,
				ReceiverUserID: inviterID,
			}
			rlsh.Create()

			// create notification
			go func() {
				n := orm.Notification{
					UnixDate: int(time.Now().Unix() * 1000), NotificationType: "4",
					SenderUserID: userID, ReceiverUserID: inviterID, GroupID: id,
				}

				if nid, e := n.Create(); e == nil {
					app.findUserByID(userID).Write(app, &WSMessage{
						MsgType:    WSM_ADD_USER_NOTIFICATION_TYPE,
						ReceiverID: strconv.Itoa(inviterID),
						Body: struct {
							ID    int    `json:"id"`
							NType string `json:"type"`
						}{ID: nid, NType: n.NotificationType},
					})
				}
			}()
		}
	}

	return []int{id}, nil
}

func (app *Application) CreateEvent(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	userID := getUserIDfromReq(w, r)
	if userID == -1 {
		return nil, errors.New("not logged")
	}

	title, description := r.PostFormValue("title"), r.PostFormValue("description")
	datetime, groups := r.PostFormValue("datetime"), strings.Split(r.PostFormValue("choosenGroups"), ",")

	dt, e := strconv.Atoi(datetime)
	if e != nil {
		return nil, errors.New("wrong datetime")
	}

	// XCSS
	if checkAllXSS(title, description) != nil {
		return nil, errors.New("wrong content")
	}

	ids := []int{}
	for _, idString := range groups {
		id, e := strconv.Atoi(idString)
		evnt := &orm.Event{
			Title: title, About: description, Type: "event",
			UnixDate: dt, UserID: userID, GroupID: id,
		}

		eventID, e := evnt.Create()
		if e != nil {
			return nil, errors.New("not create event")
		}

		// create notification
		go func() {
			n := orm.Notification{
				UnixDate: evnt.UnixDate, NotificationType: "2",
				SenderUserID: userID, EventID: eventID,
			}
			if nid, e := n.Create(); e == nil {
				app.findUserByID(userID).Write(app, &WSMessage{
					MsgType:    WSM_ADD_USER_NOTIFICATION_TYPE,
					ReceiverID: "-2",
					Body: struct {
						ID    int    `json:"id"`
						NType string `json:"type"`
					}{ID: nid, NType: n.NotificationType},
				})
			}
		}()
		ids = append(ids, eventID)
	}

	return ids, nil
}

func (app *Application) CreateMedia(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	userID := getUserIDfromReq(w, r)
	if userID == -1 {
		return nil, errors.New("not logged")
	}

	title, mediaType := r.PostFormValue("title"), r.PostFormValue("type")
	which, src, preview := r.PostFormValue("which"), r.PostFormValue("src"), r.PostFormValue("preview")

	// XCSS
	if checkAllXSS(title, mediaType, which) != nil {
		return nil, errors.New("wrong content")
	}

	ids := []int{userID}
	if which == "group" {
		ids = []int{}
		for _, groupID := range strings.Split(r.PostFormValue("choosenGroups"), ",") {
			if id, e := strconv.Atoi(groupID); e == nil {
				ids = append(ids, id)
			}
		}
	}

	datas := []int{}
	for _, id := range ids {
		m := &orm.Media{
			Title: title, MediaType: mediaType, UnixDate: int(time.Now().Unix() * 1000),
			Source: src, Preview: preview,
		}
		if which == "group" {
			m.GroupID = id
		} else {
			m.UserID = id
		}

		mediaID, e := m.Create()
		if e != nil {
			return nil, errors.New("not create " + mediaType)
		}
		datas = append(datas, mediaID)
	}

	return datas, nil
}

func getFileType(ftype string) string {
	if ftype == "image" || ftype == "photo" {
		return orm.DRIVE_IMAGE_TYPE
	}
	if ftype == "audio" {
		return orm.DRIVE_AUDIO_TYPE
	}
	if ftype == "video" {
		return orm.DRIVE_VIDEO_TYPE
	}
	return orm.DRIVE_FILE_TYPE
}

func (app *Application) CreateFile(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	userID := getUserIDfromReq(w, r)
	if userID == -1 {
		return "", errors.New("not logged")
	}

	fType := getFileType(r.PostFormValue("type"))
	link, fName, e := uploadFile("file", fType, r)
	if e != nil {
		return "", e
	}

	whomIDString := r.PostFormValue("whomID")
	if whomIDString != "" {
		whom := r.PostFormValue("whomFile")
		whomID, e := strconv.Atoi(whomIDString)
		if e != nil {
			return "", errors.New("not saved file")
		}

		cfile := &orm.ClippedFile{
			FileType: fType,
			Name:     fName,
			Source:   link,
			UserID:   userID,
		}
		setObjIDByType(whom, whomID, cfile)

		if _, e = cfile.Create(); e != nil {
			return "", errors.New("not saved file")
		}
	}
	return map[string]string{"src": link}, nil
}

func (app *Application) CreateLike(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	userID := getUserIDfromReq(w, r)
	errRes := map[string]interface{}{"carma": 0, "isLiked": false}
	if userID == -1 {
		return errRes, errors.New("not logged")
	}

	ID, e := strconv.Atoi(r.PostFormValue("id"))
	if e != nil {
		return errRes, errors.New("wrong id")
	}

	l := &orm.Like{UserID: userID, MediaID: ID}
	n := orm.Notification{
		UnixDate: int(time.Now().Unix() * 1000), NotificationType: "12",
		SenderUserID: userID, MediaID: ID,
	}

	typeLike := r.PostFormValue("type")
	noteTable := "Media"
	if typeLike == "post" {
		l.MediaID = 0
		l.PostID = ID
		n.MediaID = 0
		n.PostID = ID
		n.NotificationType = "10"
		noteTable = "Posts"
	} else if typeLike == "comment" {
		l.MediaID = 0
		l.CommentID = ID
		n.MediaID = 0
		n.CommentID = ID
		n.NotificationType = "11"
		noteTable = "Comments"
	}

	setted, e := l.Set()
	if e != nil {
		return errRes, errors.New("not setted")
	}
	if typeLike == "video" {
		n.NotificationType = "13"
	}
	if typeLike == "photo" || typeLike == "video" {
		typeLike = "media"
	}

	// create notification
	go func() {
		if data, e := orm.GetOneFrom(orm.SQLSelectParams{
			Table:   noteTable,
			What:    "userID",
			Options: orm.DoSQLOption("id=?", "", "", ID),
		}); e == nil && data != nil {
			n.ReceiverUserID = orm.FromINT64ToINT(data[0])

			if nid, e := n.Create(); e == nil {
				app.findUserByID(userID).Write(app, &WSMessage{
					MsgType:    WSM_ADD_USER_NOTIFICATION_TYPE,
					ReceiverID: strconv.Itoa(n.ReceiverUserID),
					Body: struct {
						ID    int    `json:"id"`
						NType string `json:"type"`
					}{ID: nid, NType: n.NotificationType},
				})
			}
		}
	}()

	carmaAny, e := orm.GetOneFrom(carmaCountQ(typeLike+"ID", ID))
	if e != nil || carmaAny[0] == nil {
		return errRes, e
	}
	return map[string]interface{}{"carma": carmaAny[0], "isLiked": setted}, nil
}

func (app *Application) CreateRelation(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	senderID, e := getUserID(w, r, r.PostFormValue("senderID"))
	if senderID == -1 || e != nil {
		return nil, errors.New("not logged")
	}

	ID, e := strconv.Atoi(r.PostFormValue("receiverID"))
	if e != nil {
		return nil, errors.New("wrong id")
	}

	rlsh := &orm.Relation{SenderUserID: senderID, ReceiverGroupID: ID}
	cond := "GroupID"
	if r.PostFormValue("isUser") == "true" {
		rlsh.ReceiverGroupID = 0
		rlsh.ReceiverUserID = ID
		cond = "UserID"
	}

	typeRlsh, e := strconv.Atoi(r.PostFormValue("type"))
	if e != nil {
		return nil, errors.New("wrong relation")
	}

	op := 1 // 1 - create; -1 - remove; 0 - change
	if typeRlsh == 0 {
		rlsh.Value = "1"
	} else if typeRlsh == 1 {
		op = -1
	} else if typeRlsh == 2 {
		rlsh.Value = "-1"
	} else if typeRlsh == 3 {
		op = -1
	} else if typeRlsh == 4 {
		op = 0
		rlsh.Value = "0"
	} else {
		op = -1
	}

	if op == 1 {
		id, e := rlsh.Create()

		// create notification
		go func() {
			if rlsh.Value == "-1" {
				n := orm.Notification{
					UnixDate: int(time.Now().Unix() * 1000), NotificationType: "5",
					SenderUserID: senderID, ReceiverUserID: rlsh.ReceiverUserID, RelationID: id,
				}
				if cond == "GroupID" {
					n.NotificationType = "6"
					n.GroupID = rlsh.ReceiverGroupID
					data, e := orm.GetOneFrom(orm.SQLSelectParams{
						Table:   "Groups",
						What:    "ownerUserID",
						Options: orm.DoSQLOption("id=?", "", "", rlsh.ReceiverGroupID),
					})
					if e != nil || data == nil {
						return
					}
					n.ReceiverUserID = orm.FromINT64ToINT(data[0])
				}

				if nid, e := n.Create(); e == nil {
					app.findUserByID(senderID).Write(app, &WSMessage{
						MsgType:    WSM_ADD_USER_NOTIFICATION_TYPE,
						ReceiverID: strconv.Itoa(n.ReceiverUserID),
						Body: struct {
							ID    int    `json:"id"`
							NType string `json:"type"`
						}{ID: nid, NType: n.NotificationType},
					})
				}
			}
		}()
		return id, e
	} else if op == -1 {
		return nil, orm.DeleteByParams(orm.SQLDeleteParams{
			Table: "Relations",
			Options: orm.DoSQLOption(
				"(senderUserID = ? AND receiver"+cond+"=?) OR (receiverUserID=? AND sender"+cond+"=?)",
				"",
				"",
				senderID, ID, senderID, ID,
			),
		})
	}

	id, e := orm.GetOneFrom(orm.SQLSelectParams{
		Table: "Relations",
		What:  "id",
		Options: orm.DoSQLOption(
			"(senderUserID = ? AND receiver"+cond+"=?) OR (receiverUserID=? AND sender"+cond+"=?)",
			"",
			"",
			senderID, ID, senderID, ID,
		),
	})
	if e != nil || id[0] == nil {
		return nil, errors.New("not changed")
	}
	rlsh.ID = orm.FromINT64ToINT(id[0])
	return nil, rlsh.Change()
}

func (app *Application) CreateEventAnswer(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	userID := getUserIDfromReq(w, r)
	if userID == -1 {
		return nil, errors.New("not logged")
	}

	ID, e := strconv.Atoi(r.PostFormValue("id"))
	if e != nil {
		return nil, errors.New("wrong id")
	}

	ea := &orm.EventAnswer{UserID: userID, EventID: ID, Answer: r.PostFormValue("answer")}
	if e := ea.Create(); e != nil {
		return nil, errors.New("wrong event answer")
	}

	goingQ := eventAnswerQ(0, ID)
	notGoingQ := eventAnswerQ(1, ID)
	idkQ := eventAnswerQ(2, ID)
	votes, e := orm.GetWithSubqueries(goingQ, []orm.SQLSelectParams{notGoingQ, idkQ}, []string{}, []string{"votes0", "votes1", "votes2"}, struct{}{})
	if e != nil {
		return nil, e
	}
	return votes[0], nil
}

func (app *Application) CreateChat(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	userID := getUserIDfromReq(w, r)
	if userID == -1 {
		return nil, errors.New("not logged")
	}

	ID, e := strconv.Atoi(r.PostFormValue("id"))
	if e != nil {
		return nil, errors.New("wrong id")
	}

	chat := &orm.Chat{
		ChatType: "user", Users: fmt.Sprintf("|%v |%v ", userID, ID),
		SenderUserID: userID, ReceiverUserID: ID,
	}
	typeChat := r.PostFormValue("type")
	if typeChat == "group" {
		chat.ChatType = "group"
		chat.ReceiverUserID = 0
		chat.ReceiverGroupID = ID
	}

	id, e := chat.Create()
	if e != nil {
		return -1, errors.New("not created chat")
	}
	return []int{id}, nil
}

func (app *Application) CreateMessage(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	userID := getUserIDfromReq(w, r)
	if userID == -1 {
		return nil, errors.New("not logged")
	}

	ID, e := strconv.Atoi(r.PostFormValue("id"))
	if e != nil {
		return nil, errors.New("wrong id")
	}

	body, messageType := r.PostFormValue("body"), r.PostFormValue("messageType")
	if checkAllXSS(body, messageType) != nil {
		return nil, errors.New("wrong content")
	}

	m := orm.Message{
		UnixDate: int(time.Now().Unix() * 1000), Body: body,
		SenderUserID: userID, ReceiverUserID: ID, MessageType: messageType,
	}
	c := orm.Chat{
		ChatType: "user", Users: fmt.Sprintf("|%v |%v ", userID, ID),
		SenderUserID: userID, ReceiverUserID: m.ReceiverUserID, ReceiverGroupID: m.ReceiverGroupID,
	}

	typeChat := r.PostFormValue("type")
	if typeChat == "group" {
		c.ChatType = "group"
		c.ReceiverUserID = 0
		c.ReceiverGroupID = ID
		m.ReceiverUserID = 0
		m.ReceiverGroupID = ID
	}

	if _, e = c.Create(); e != nil {
		return -1, errors.New("not created message: not create chat")
	}

	id, e := m.Create()
	if e != nil {
		return -1, errors.New("not created message")
	}
	return []int{id}, nil
}

func (app *Application) CreateComment(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	userID := getUserIDfromReq(w, r)
	if userID == -1 {
		return nil, errors.New("not logged")
	}
	ID, e := strconv.Atoi(r.PostFormValue("id"))
	if e != nil {
		return nil, errors.New("wrong id")
	}

	commentType, body := r.PostFormValue("type"), r.PostFormValue("body")
	isAnswer, isHaveClippedFiles := r.PostFormValue("isAnswer"), r.PostFormValue("isHaveClippedFiles")
	if checkAllXSS(body, isHaveClippedFiles, isAnswer) != nil {
		return nil, errors.New("wrong content")
	}

	c := orm.Comment{
		Body: body, UnixDate: int(time.Now().Unix() * 1000),
		IsHaveChild: "0", IsAnswer: isAnswer, IsHaveClippedFiles: isHaveClippedFiles,
		UserID: userID, PostID: ID,
	}
	n := orm.Notification{
		UnixDate: int(time.Now().Unix() * 1000), NotificationType: "20",
		SenderUserID: userID, PostID: ID,
	}
	noteTable := "Posts"
	if commentType == "comment" {
		c.PostID = 0
		c.CommentID = ID
		n.PostID = 0
		n.CommentID = ID
		n.NotificationType = "21"

		if r.PostFormValue("isHaveChild") == "0" {
			parentComment := orm.Comment{ID: ID, IsHaveChild: "1"}
			if e := parentComment.Change(); e != nil {
				return -1, errors.New("not created comment, parent not changed")
			}
		}
	} else if commentType == "media" || commentType == "video" || commentType == "photo" {
		c.PostID = 0
		c.MediaID = ID
		n.PostID = 0
		n.MediaID = ID
		n.NotificationType = "22"
	}

	// create notification
	go func() {
		if commentType == "video" {
			n.NotificationType = "23"
		}
		if data, e := orm.GetOneFrom(orm.SQLSelectParams{
			Table:   noteTable,
			What:    "userID",
			Options: orm.DoSQLOption("id=?", "", "", ID),
		}); e == nil && data != nil {
			n.ReceiverUserID = orm.FromINT64ToINT(data[0])

			if nid, e := n.Create(); e == nil {
				app.findUserByID(userID).Write(app, &WSMessage{
					MsgType:    WSM_ADD_USER_NOTIFICATION_TYPE,
					ReceiverID: strconv.Itoa(n.ReceiverUserID),
					Body: struct {
						ID    int    `json:"id"`
						NType string `json:"type"`
					}{ID: nid, NType: n.NotificationType},
				})
			}
		}
	}()

	id, e := c.Create()
	if e != nil {
		return -1, errors.New("not created comment")
	}
	return []int{id}, nil
}

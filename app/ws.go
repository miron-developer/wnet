package app

import (
	"anifor/app/dbfuncs"
	"errors"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

// WSMessage types
const (
	AuthType               = 0  // when created ws connection, logout\login
	ChatMessageType        = 1  // when send chat message
	CommentType            = 2  // when new comment in real time
	PostType               = 3  // when new post in real time
	UserTypeOnType         = 4  // when some user online
	UserTypeOffType        = 5  // when some user offline
	StartSendingImageType  = 6  // when send image to other user: start
	ImageDataSendingType   = 7  // send image data(pocket) per 256kb for ex
	FinishSendingImageType = 8  // finished send image
	TypingStartType        = 9  // user typing
	TypingFinishType       = 10 // user finish typing
	ErrorType              = -1 // if error was

	writeWait  = 10 * time.Second
	pongWait   = 60 * time.Second
	pingPeriod = (pongWait * 9) / 10
)

// WSMessage one message from ws connection to users
type WSMessage struct {
	MsgType   int         `json:"msgType"`   // ws message type
	Addresser string      `json:"addresser"` // who sended
	Receiver  string      `json:"receiver"`  // who get
	Body      interface{} `json:"body"`      // message body
}

// WSUser is one ws connection user
type WSUser struct {
	Conn     *websocket.Conn
	WSName   string
	Nickname string
	ID       int
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  2048,
	WriteBufferSize: 2048,
	CheckOrigin:     func(*http.Request) bool { return true },
}

func (app *Application) findUserByNick(finder string) *WSUser {
	for _, user := range app.OnlineUsers {
		if user.Nickname == finder {
			return user
		}
	}
	return nil
}

func (app *Application) findUserByID(finder int) *WSUser {
	for _, user := range app.OnlineUsers {
		if user.ID == finder {
			return user
		}
	}
	return nil
}

// ChangeWSUserName change username if it logged
func (app *Application) ChangeWSUserName(oldName, newName string, ID int) error {
	user := app.findUserByNick(oldName)
	if user == nil {
		return errors.New("user not founded")
	}
	user.Nickname = newName
	user.ID = ID
	return nil
}

// WSWork work with channel
func (app *Application) WSWork() {
	for {
		msg := <-app.Messages
		// fmt.Println("msg:", msg, msg.Body)
		if msg.Receiver == "all" {
			for _, v := range app.OnlineUsers {
				v.Conn.WriteJSON(msg)
			}
		} else {
			receiver := app.findUserByNick(msg.Receiver)
			addresser := app.findUserByNick(msg.Addresser)
			if receiver == nil || (msg.Addresser != "server" && addresser == nil) {
				continue
			}

			if msg.MsgType == ChatMessageType {
				if e := dbfuncs.CreateMessage(&dbfuncs.Messages{
					Date:       TimeExpire(time.Nanosecond),
					Body:       msg.Body.(string),
					ReceiverID: receiver.ID,
					SenderID:   addresser.ID,
				}); e != nil {
					addresser.Conn.WriteJSON(WSMessage{Addresser: "server", Receiver: addresser.Nickname,
						Body: "something wrong: " + e.Error(), MsgType: ErrorType})
				}
			}

			if e := receiver.Conn.WriteJSON(msg); e != nil {
				addresser.Conn.WriteJSON(WSMessage{Addresser: "server", Receiver: addresser.Nickname,
					Body: "something wrong: " + e.Error(), MsgType: ErrorType})
			}
		}
	}
}

// HandleUserMsg handle received msg from front user
func (user *WSUser) HandleUserMsg(app *Application) {
	user.Conn.SetPongHandler(
		func(string) error {
			user.Conn.SetReadDeadline(time.Now().Add(pongWait))
			return nil
		},
	)

	for {
		msg := &WSMessage{}
		if e := user.Conn.ReadJSON(msg); e != nil {
			if user.ID > 0 {
				app.Messages <- &WSMessage{MsgType: UserTypeOffType, Addresser: "server", Receiver: "all", Body: user}
			}

			app.m.Lock()
			delete(app.OnlineUsers, user.WSName)
			app.m.Unlock()
			app.ELog.Println(e)
			user.Conn.Close()
			return
		}
		app.Messages <- msg
	}
}

// Pinger ping every pingPeriod
func (user *WSUser) Pinger() {
	ticker := time.NewTicker(pingPeriod)
	for {
		<-ticker.C
		if err := user.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
			return
		}
	}
}

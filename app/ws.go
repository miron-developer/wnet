package app

import (
	"net/http"
	"strconv"
	"time"
	"wnet/app/dbfuncs"

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
	MsgType     int         `json:"msgType"`   // ws message type
	AddresserID string      `json:"addresser"` // who sended, if = -1, then send server
	ReceiverID  string      `json:"receiver"`  // who get, if = -2, then get all, -1 server
	Body        interface{} `json:"body"`      // message body
}

// WSUser is one ws connection user
type WSUser struct {
	Conn *websocket.Conn
	ID   int
}

// ChatRoom is room between users
type ChatRoom struct {
	RoomID          string
	Type            string // group chat or p2p
	Users           map[int]*WSUser
	Messages        chan *WSMessage
	LastMsgUnixTime int64
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  2048,
	WriteBufferSize: 2048,
	CheckOrigin:     func(*http.Request) bool { return true },
}

func (app *Application) findUserByID(finder int) *WSUser {
	for _, user := range app.OnlineUsers {
		if user.ID == finder {
			return user
		}
	}
	return nil
}

func (room *ChatRoom) Work(app *Application) {
	for {
		msg := WSMessage{}
		if msg.ReceiverID == "all" {
			for _, v := range room.Users {
				v.Conn.WriteJSON(msg)
			}
		} else {
			receiverID, recErr := strconv.Atoi(msg.ReceiverID)
			addresserID, addErr := strconv.Atoi(msg.AddresserID)
			if recErr != nil || addErr != nil {
				continue
			}

			receiver := app.findUserByID(receiverID)
			addresser := app.findUserByID(addresserID)
			if receiver == nil || (msg.AddresserID == "" && addresser == nil) {
				continue
			}

			if msg.MsgType == ChatMessageType {
				body := dbfuncs.MakeArrFromStruct(msg.Body)
				chatMsg := &dbfuncs.Message{
					UnixDate:     int(time.Now().Unix()),
					Body:         body[0].(string),
					SenderUserID: addresserID,
					MessageType:  body[1].(string),
				}
				if room.Type == "group" {
					chatMsg.ReceiverGroupID = receiverID
				}
				if e := chatMsg.Create(); e != nil {
					addresser.Conn.WriteJSON(WSMessage{AddresserID: "server", ReceiverID: strconv.Itoa(addresser.ID),
						Body: "something wrong: " + e.Error(), MsgType: ErrorType})
				}
			}

			if e := receiver.Conn.WriteJSON(msg); e != nil {
				addresser.Conn.WriteJSON(WSMessage{AddresserID: "server", ReceiverID: strconv.Itoa(addresser.ID),
					Body: "something wrong: " + e.Error(), MsgType: ErrorType})
			}
		}
	}
}

// WSWork work with channel
func (app *Application) WSWork() {
	for {
		msg := WSMessage{}
		if msg.ReceiverID == "all" {
			for _, v := range app.OnlineUsers {
				v.Conn.WriteJSON(msg)
			}
		} else {
			receiverID, recErr := strconv.Atoi(msg.ReceiverID)
			addresserID, addErr := strconv.Atoi(msg.AddresserID)
			if recErr != nil || addErr != nil {
				continue
			}

			receiver := app.findUserByID(receiverID)
			addresser := app.findUserByID(addresserID)
			if receiver == nil || (msg.AddresserID == "" && addresser == nil) {
				continue
			}

			// if msg.MsgType == ChatMessageType {
			// 	if e := dbfuncs.CreateMessage(&dbfuncs.Messages{
			// 		Date:       TimeExpire(time.Nanosecond),
			// 		Body:       msg.Body.(string),
			// 		ReceiverID: receiver.ID,
			// 		SenderID:   addresser.ID,
			// 	}); e != nil {
			// 		addresser.Conn.WriteJSON(WSMessage{AddresserID: "server", ReceiverID: strconv.Itoa(addresser.ID),
			// 			Body: "something wrong: " + e.Error(), MsgType: ErrorType})
			// 	}
			// }

			if e := receiver.Conn.WriteJSON(msg); e != nil {
				addresser.Conn.WriteJSON(WSMessage{AddresserID: "server", ReceiverID: strconv.Itoa(addresser.ID),
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
				app.Messages <- &WSMessage{MsgType: UserTypeOffType, AddresserID: "server", ReceiverID: "all", Body: user}
			}

			app.m.Lock()
			delete(app.OnlineUsers, user.ID)
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
